package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"

	"github.com/pion/rtp/codecs"
	"github.com/pion/webrtc/v3"
	"gopkg.in/hraban/opus.v2"
)

// Global
var (
	clientPC    *webrtc.PeerConnection
	opusDecoder *opus.Decoder
)

// WebRTC message types
type ICECandidateRequest struct {
	Candidate webrtc.ICECandidateInit `json:"candidate"`
}
type WebRTCOfferRequest struct {
	Offer webrtc.SessionDescription `json:"offer"`
}
type WebRTCOfferResponse struct {
	Answer webrtc.SessionDescription `json:"answer"`
}

// Whisper response
type WhisperResponse struct {
	Text string `json:"text"`
}

const htmlContent = `<!DOCTYPE html>
<html>
  <body>
    <h1> Whisper Live Test</h1>
    <button id="connectBtn">Connect</button>
    <pre id="log"></pre>
    <script>
    let pc;
    let log = msg => document.getElementById('log').innerText += msg + "\n";

    document.getElementById('connectBtn').onclick = async () => {
      pc = new RTCPeerConnection();
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
      stream.getTracks().forEach(t => pc.addTrack(t, stream));

      pc.onicecandidate = e => {
        if (e.candidate) {
          fetch('/ice-candidate',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({candidate:e.candidate})});
        }
      };

      const offer = await pc.createOffer();
      await pc.setLocalDescription(offer);
      const resp = await fetch('/webrtc-offer',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({offer})});
      const {answer} = await resp.json();
      await pc.setRemoteDescription(answer);
      log("Connected!");
    }
    </script>
  </body>
</html>`

func main() {
	var err error
	opusDecoder, err = opus.NewDecoder(48000, 1)
	if err != nil {
		log.Fatalf("Opus decoder error: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(htmlContent))
	})
	mux.HandleFunc("/ice-candidate", handleICECandidate)
	mux.HandleFunc("/webrtc-offer", handleWebRTCOffer)

	fmt.Println("🚀 Server running at http://localhost:3000")
	log.Fatal(http.ListenAndServe(":3000", mux))
}

func handleICECandidate(w http.ResponseWriter, r *http.Request) {
	var req ICECandidateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad candidate", 400)
		return
	}
	if clientPC != nil && req.Candidate.Candidate != "" {
		clientPC.AddICECandidate(req.Candidate)
	}
	w.Write([]byte(`{"ok":true}`))
}

func handleWebRTCOffer(w http.ResponseWriter, r *http.Request) {
	var req WebRTCOfferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad offer", 400)
		return
	}

	config := webrtc.Configuration{ICEServers: []webrtc.ICEServer{{URLs: []string{"stun:stun.l.google.com:19302"}}}}
	var err error
	clientPC, err = webrtc.NewPeerConnection(config)
	if err != nil {
		http.Error(w, "pc error", 500)
		return
	}

	clientPC.OnTrack(func(track *webrtc.TrackRemote, recv *webrtc.RTPReceiver) {
		if track.Kind() == webrtc.RTPCodecTypeAudio {
			go handleClientAudio(track)
		}
	})

	if err := clientPC.SetRemoteDescription(req.Offer); err != nil {
		http.Error(w, "set remote fail", 500)
		return
	}
	answer, _ := clientPC.CreateAnswer(nil)
	clientPC.SetLocalDescription(answer)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(WebRTCOfferResponse{Answer: answer})
}

func handleClientAudio(track *webrtc.TrackRemote) {
	opusPacket := &codecs.OpusPacket{}
	pcmBuffer := make([]int16, 0, 48000*3)

	for {
		rtpPacket, _, err := track.ReadRTP()
		if err != nil {
			log.Printf("RTP read error: %v", err)
			return
		}
		opusData, err := opusPacket.Unmarshal(rtpPacket.Payload)
		if err != nil {
			continue
		}
		pcmData := make([]int16, 960)
		n, err := opusDecoder.Decode(opusData, pcmData)
		if err != nil {
			continue
		}
		pcmBuffer = append(pcmBuffer, pcmData[:n]...)

		if len(pcmBuffer) >= 144000 {
			// Write to wav + send
			wavBuf := pcmToWav(pcmBuffer, 48000)
			go sendToWhisper(wavBuf)
			pcmBuffer = pcmBuffer[:0]
		}
	}
}

// Encode PCM to minimal WAV
func pcmToWav(samples []int16, sampleRate int) []byte {
	buf := new(bytes.Buffer)

	// WAV header
	byteRate := sampleRate * 2
	blockAlign := 2
	dataSize := len(samples) * 2
	binary.Write(buf, binary.LittleEndian, []byte("RIFF"))
	binary.Write(buf, binary.LittleEndian, uint32(36+dataSize))
	binary.Write(buf, binary.LittleEndian, []byte("WAVEfmt "))
	binary.Write(buf, binary.LittleEndian, uint32(16)) // PCM header size
	binary.Write(buf, binary.LittleEndian, uint16(1))  // PCM format
	binary.Write(buf, binary.LittleEndian, uint16(1))  // mono
	binary.Write(buf, binary.LittleEndian, uint32(sampleRate))
	binary.Write(buf, binary.LittleEndian, uint32(byteRate))
	binary.Write(buf, binary.LittleEndian, uint16(blockAlign))
	binary.Write(buf, binary.LittleEndian, uint16(16)) // bits
	binary.Write(buf, binary.LittleEndian, []byte("data"))
	binary.Write(buf, binary.LittleEndian, uint32(dataSize))
	for _, s := range samples {
		binary.Write(buf, binary.LittleEndian, s)
	}
	return buf.Bytes()
}

// Send wav to Whisper HTTP API
func sendToWhisper(wavData []byte) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	fw, _ := writer.CreateFormFile("file", "chunk.wav")
	io.Copy(fw, bytes.NewReader(wavData))
	writer.WriteField("temperature", "0.2")
	writer.WriteField("response-format", "json")
	writer.WriteField("audio_format", "wav")
	writer.Close()

	req, _ := http.NewRequest("POST", "http://127.0.0.1:8080/inference", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Whisper call error: %v", err)
		return
	}
	defer resp.Body.Close()

	var out WhisperResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err == nil {
		fmt.Printf("📝 Whisper: %s\n", out.Text)
	} else {
		b, _ := io.ReadAll(resp.Body)
		log.Printf("Whisper parse fail: %s", string(b))
	}
}

/**
cd whisper.cpp

#Downloads the model
sh ./models/download-ggml-model.sh base.en

#Build
cmake -B build
cmake --build build -j --config Release

# Running server
./build/bin/whisper-server -m models/ggml-base.en.bin

**/
