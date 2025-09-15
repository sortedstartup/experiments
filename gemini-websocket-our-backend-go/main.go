package main

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/pion/rtp"
	"github.com/pion/rtp/codecs"
	"github.com/pion/webrtc/v3"
	"github.com/rs/cors"
	"gopkg.in/hraban/opus.v2"
)

// Global variables
var (
	clientPC            *webrtc.PeerConnection
	geminiWS            *websocket.Conn
	clientOutboundTrack *webrtc.TrackLocalStaticRTP
	sequenceNumber      uint16 = 1
	timestamp           uint32 = 1
	timestampMutex      sync.Mutex
	opusEncoder         *opus.Encoder
	opusDecoder         *opus.Decoder
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

// CORRECTED: Gemini message types based on API docs
type GeminiMessage struct {
	Setup         *GeminiSetup         `json:"setup,omitempty"`
	RealtimeInput *GeminiRealtimeInput `json:"realtimeInput,omitempty"`
}

type GeminiSetup struct {
	Model            string                  `json:"model"`
	GenerationConfig *GeminiGenerationConfig `json:"generationConfig,omitempty"`
}

type GeminiGenerationConfig struct {
	ResponseModalities []string `json:"responseModalities"`
}

// CORRECTED: Use realtimeInput for audio streaming
type GeminiRealtimeInput struct {
	MediaChunks []GeminiMediaChunk `json:"mediaChunks"`
}

type GeminiMediaChunk struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"`
}

const htmlContent = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Gemini Live WebRTC Audio Chat</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            max-width: 800px;
            margin: 0 auto;
            padding: 20px;
            background-color: #f5f5f5;
        }
        .container {
            background: white;
            padding: 30px;
            border-radius: 15px;
            box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
            text-align: center;
        }
        h1 { color: #333; margin-bottom: 30px; }
        button {
            padding: 15px 30px;
            margin: 10px;
            border: none;
            border-radius: 25px;
            font-size: 16px;
            cursor: pointer;
            transition: all 0.3s ease;
        }
        .connect-btn { background: #4CAF50; color: white; }
        .connect-btn:hover { background: #45a049; }
        .connect-btn:disabled { background: #cccccc; cursor: not-allowed; }
        .status {
            margin: 20px 0;
            padding: 15px;
            border-radius: 10px;
            font-weight: bold;
        }
        .status.connecting { background: #fff3cd; color: #856404; }
        .status.connected { background: #d4edda; color: #155724; }
        .status.disconnected { background: #f8d7da; color: #721c24; }
        .log {
            background: #f8f9fa;
            padding: 15px;
            border-radius: 5px;
            height: 200px;
            overflow-y: auto;
            margin-top: 20px;
            font-size: 12px;
            text-align: left;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🎤 Gemini Live WebRTC Audio Chat</h1>
        <p>Simple audio streaming between browser and Gemini Live API via WebRTC</p>
        
        <button id="connectBtn" class="connect-btn">Connect With Gemini</button>
        <div id="status" class="status disconnected">Click Connect to start</div>
        
        <audio id="remoteAudio" autoplay playsinline controls></audio>
        
        <div id="log" class="log">Waiting to connect...</div>
    </div>

    <script>
    let pc = null;
    let isConnected = false;

    const connectBtn = document.getElementById('connectBtn');
    const status = document.getElementById('status');
    const remoteAudio = document.getElementById('remoteAudio');
    const log = document.getElementById('log');

    function addLog(message) {
        const time = new Date().toLocaleTimeString();
        log.innerHTML += '<div>[' + time + '] ' + message + '</div>';
        log.scrollTop = log.scrollHeight;
    }

    connectBtn.onclick = () => {
        if (!isConnected) connectToGemini();
    };

    async function connectToGemini() {
        try {
            updateStatus('Connecting...', 'connecting');
            connectBtn.disabled = true;
            addLog('Starting WebRTC connection...');

            pc = new RTCPeerConnection({
                iceServers: [{ urls: 'stun:stun.l.google.com:19302' }]
            });

            pc.onicecandidate = (event) => {
                if (event.candidate) {
                    fetch('/ice-candidate', {
                        method: 'POST',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify({ candidate: event.candidate })
                    }).catch(console.error);
                }
            };

            pc.onconnectionstatechange = () => {
                addLog('Connection state: ' + pc.connectionState);
                if (pc.connectionState === 'connected') {
                    updateStatus('✅ Connected - Speak to Gemini!', 'connected');
                    isConnected = true;
                    addLog('🎤 Ready for audio chat!');
                } else if (pc.connectionState === 'disconnected' || pc.connectionState === 'failed') {
                    updateStatus('❌ Disconnected', 'disconnected');
                    isConnected = false;
                    connectBtn.disabled = false;
                    addLog('Connection lost');
                }
            };

            // FIXED: Proper audio track handling
            pc.ontrack = (event) => {
                addLog('🔊 Audio track received from server');
                const [remoteStream] = event.streams;
                remoteAudio.srcObject = remoteStream;
                
                // Enable audio and force play
                remoteAudio.muted = false;
                remoteAudio.volume = 1.0;
                remoteAudio.autoplay = true;
                
                // Try to play immediately
                remoteAudio.play().then(() => {
                    addLog('✅ Audio playback started');
                }).catch(e => {
                    addLog('⚠️ Audio autoplay blocked - click to enable');
                    // Create a play button for user interaction
                    const playBtn = document.createElement('button');
                    playBtn.textContent = '🔊 Enable Audio';
                    playBtn.onclick = () => {
                        remoteAudio.play().then(() => {
                            addLog('✅ Audio enabled manually');
                            playBtn.remove();
                        }).catch(err => addLog('❌ Audio play error: ' + err.message));
                    };
                    document.querySelector('.container').appendChild(playBtn);
                });
            };

            // Get microphone with specific settings for 48kHz
            const stream = await navigator.mediaDevices.getUserMedia({ 
                audio: {
                    sampleRate: 48000,
                    channelCount: 1,
                    echoCancellation: false,
                    noiseSuppression: false,
                    autoGainControl: false
                }
            });

            addLog('✅ Microphone access granted');
            
            stream.getTracks().forEach(track => {
                pc.addTrack(track, stream);
                addLog('Added track: ' + track.kind);
            });

            const offer = await pc.createOffer();
            await pc.setLocalDescription(offer);
            addLog('Created WebRTC offer');

            const response = await fetch('/webrtc-offer', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ offer: offer })
            });

            if (!response.ok) {
                throw new Error('HTTP ' + response.status + ': ' + response.statusText);
            }

            const { answer } = await response.json();
            await pc.setRemoteDescription(answer);
            addLog('✅ WebRTC negotiation complete');

        } catch (error) {
            console.error('Connection failed:', error);
            addLog('❌ Error: ' + error.message);
            updateStatus('❌ Failed: ' + error.message, 'disconnected');
            connectBtn.disabled = false;
        }
    }

    function updateStatus(message, className) {
        status.textContent = message;
        status.className = 'status ' + className;
    }
</script>

</body>
</html>`

func main() {
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		log.Fatal("GOOGLE_API_KEY environment variable is required")
	}

	// Initialize Opus encoder/decoder
	var err error
	opusEncoder, err = opus.NewEncoder(48000, 1, opus.AppVoIP)
	if err != nil {
		log.Fatalf("Failed to create Opus encoder: %v", err)
	}
	opusEncoder.SetBitrate(64000) // 64kbps

	opusDecoder, err = opus.NewDecoder(48000, 1)
	if err != nil {
		log.Fatalf("Failed to create Opus decoder: %v", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(htmlContent))
	})

	mux.HandleFunc("/ice-candidate", handleICECandidate)
	mux.HandleFunc("/webrtc-offer", func(w http.ResponseWriter, r *http.Request) {
		handleWebRTCOffer(w, r, apiKey)
	})

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	})

	handler := c.Handler(mux)

	fmt.Println("🚀 Server running on http://localhost:3000")
	log.Fatal(http.ListenAndServe(":3000", handler))
}

func handleICECandidate(w http.ResponseWriter, r *http.Request) {
	var req ICECandidateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if clientPC != nil && req.Candidate.Candidate != "" {
		if err := clientPC.AddICECandidate(req.Candidate); err != nil {
			log.Printf("❌ AddICECandidate error: %v", err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func handleWebRTCOffer(w http.ResponseWriter, r *http.Request, apiKey string) {
	var req WebRTCOfferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	fmt.Println("🚀 Setting up WebRTC connection...")

	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{URLs: []string{"stun:stun.l.google.com:19302"}},
		},
	}

	var err error
	clientPC, err = webrtc.NewPeerConnection(config)
	if err != nil {
		http.Error(w, "Failed to create peer connection", http.StatusInternalServerError)
		return
	}

	clientPC.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		fmt.Printf("📡 WebRTC Connection: %s\n", s)
		if s == webrtc.PeerConnectionStateConnected {
			go connectToGemini(apiKey)
		}
	})

	// Handle incoming audio from client
	clientPC.OnTrack(func(remoteTrack *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		fmt.Printf("🎤 Receiving %s track from client\n", remoteTrack.Kind())
		if remoteTrack.Kind() == webrtc.RTPCodecTypeAudio {
			go handleClientAudio(remoteTrack)
		}
	})

	// Create track for sending audio back to client
	outboundTrack, err := webrtc.NewTrackLocalStaticRTP(
		webrtc.RTPCodecCapability{MimeType: "audio/opus", ClockRate: 48000, Channels: 1},
		"gemini-audio", "gemini-stream",
	)
	if err != nil {
		http.Error(w, "Failed to create outbound track", http.StatusInternalServerError)
		return
	}

	clientOutboundTrack = outboundTrack
	if _, err := clientPC.AddTrack(clientOutboundTrack); err != nil {
		http.Error(w, "Failed to add outbound track", http.StatusInternalServerError)
		return
	}

	if err := clientPC.SetRemoteDescription(req.Offer); err != nil {
		http.Error(w, "SetRemoteDescription failed", http.StatusInternalServerError)
		return
	}

	answer, err := clientPC.CreateAnswer(nil)
	if err != nil {
		http.Error(w, "CreateAnswer failed", http.StatusInternalServerError)
		return
	}

	if err := clientPC.SetLocalDescription(answer); err != nil {
		http.Error(w, "SetLocalDescription failed", http.StatusInternalServerError)
		return
	}

	fmt.Println("✅ WebRTC setup complete")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(WebRTCOfferResponse{Answer: answer})
}

func handleClientAudio(track *webrtc.TrackRemote) {
	opusPacket := &codecs.OpusPacket{}
	pcmBuffer := make([]int16, 0, 48000) // Buffer for 1 second at 48kHz

	for {
		rtpPacket, _, err := track.ReadRTP()
		if err != nil {
			fmt.Printf("❌ Error reading RTP: %v\n", err)
			return
		}

		// Extract Opus payload
		opusData, err := opusPacket.Unmarshal(rtpPacket.Payload)
		if err != nil {
			continue
		}

		// Decode Opus to PCM
		pcmData := make([]int16, 960) // 20ms at 48kHz
		n, err := opusDecoder.Decode(opusData, pcmData)
		if err != nil {
			continue
		}

		// Accumulate PCM data
		pcmBuffer = append(pcmBuffer, pcmData[:n]...)

		// Send to Gemini when we have ~500ms of data
		if len(pcmBuffer) >= 24000 { // 0.5 seconds at 48kHz
			// Downsample to 16kHz for Gemini
			pcm16kHz := downsample48to16(pcmBuffer)

			// Convert to bytes
			pcmBytes := make([]byte, len(pcm16kHz)*2)
			for i, sample := range pcm16kHz {
				binary.LittleEndian.PutUint16(pcmBytes[i*2:], uint16(sample))
			}

			sendAudioToGemini(pcmBytes)
			pcmBuffer = pcmBuffer[:0] // Clear buffer
		}
	}
}

// Simple downsampling from 48kHz to 16kHz (3:1 ratio)
func downsample48to16(input []int16) []int16 {
	output := make([]int16, len(input)/3)
	for i := 0; i < len(output); i++ {
		output[i] = input[i*3] // Simple decimation
	}
	return output
}

// FIXED: Proper upsampling from 24kHz to 48kHz (1:2 ratio)
func upsample24to48(input []int16) []int16 {
	output := make([]int16, len(input)*2)
	for i := 0; i < len(input); i++ {
		output[i*2] = input[i] // Original sample
		if i*2+1 < len(output) {
			output[i*2+1] = input[i] // Duplicate sample
		}
	}
	return output
}

func connectToGemini(apiKey string) {
	fmt.Println("🔗 Connecting to Gemini Live API...")

	wsURL := "wss://generativelanguage.googleapis.com/ws/google.ai.generativelanguage.v1beta.GenerativeService.BidiGenerateContent?key=" + apiKey

	var err error
	geminiWS, _, err = websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		log.Printf("❌ Failed to connect to Gemini: %v", err)
		return
	}

	fmt.Println("✅ Connected to Gemini WebSocket")

	setupMsg := GeminiMessage{
		Setup: &GeminiSetup{
			Model: "models/gemini-live-2.5-flash-preview",
			GenerationConfig: &GeminiGenerationConfig{
				ResponseModalities: []string{"AUDIO"},
			},
		},
	}

	if err := geminiWS.WriteJSON(setupMsg); err != nil {
		log.Printf("❌ Failed to send setup: %v", err)
		return
	}

	// Wait for setup acknowledgment
	var setupResponse map[string]interface{}
	if err := geminiWS.ReadJSON(&setupResponse); err != nil {
		log.Printf("❌ Setup failed: %v", err)
		return
	}

	fmt.Println("✅ Gemini setup complete")
	go handleGeminiResponses()
}

// CORRECTED: Use realtimeInput with mediaChunks
func sendAudioToGemini(audioData []byte) {
	if geminiWS == nil || len(audioData) == 0 {
		return
	}

	encodedAudio := base64.StdEncoding.EncodeToString(audioData)

	msg := GeminiMessage{
		RealtimeInput: &GeminiRealtimeInput{
			MediaChunks: []GeminiMediaChunk{
				{
					MimeType: "audio/pcm;rate=16000",
					Data:     encodedAudio,
				},
			},
		},
	}

	if err := geminiWS.WriteJSON(msg); err != nil {
		log.Printf("❌ Error sending to Gemini: %v", err)
	} else {
		fmt.Printf("📤 Sent %d bytes to Gemini\n", len(audioData))
	}
}

// CORRECTED: Response parsing for serverContent
func handleGeminiResponses() {
	for {
		var response map[string]interface{}
		if err := geminiWS.ReadJSON(&response); err != nil {
			log.Printf("❌ Gemini connection closed: %v", err)
			return
		}

		// Debug: Print full response structure
		responseBytes, _ := json.MarshalIndent(response, "", "  ")
		fmt.Printf("🔍 Gemini Response: %s\n", string(responseBytes))

		// Extract audio from serverContent response
		if serverContent, ok := response["serverContent"].(map[string]interface{}); ok {
			if modelTurn, ok := serverContent["modelTurn"].(map[string]interface{}); ok {
				if parts, ok := modelTurn["parts"].([]interface{}); ok {
					for _, part := range parts {
						if partMap, ok := part.(map[string]interface{}); ok {
							if inlineData, ok := partMap["inlineData"].(map[string]interface{}); ok {
								if audioData, ok := inlineData["data"].(string); ok {
									fmt.Printf("📥 Received %d chars of audio from Gemini\n", len(audioData))
									sendAudioToClient(audioData)
								}
							}
						}
					}
				}
			}
		}

		// Check for setup acknowledgment
		if setupComplete, ok := response["setupComplete"]; ok {
			fmt.Printf("✅ Gemini setup acknowledged: %+v\n", setupComplete)
		}
	}
}

// CORRECTED: Audio processing for 24kHz input
func sendAudioToClient(base64Audio string) {
	if clientOutboundTrack == nil {
		return
	}

	// Decode Gemini's 24kHz PCM audio
	pcmData, err := base64.StdEncoding.DecodeString(base64Audio)
	if err != nil || len(pcmData) == 0 {
		fmt.Printf("❌ Failed to decode audio: %v\n", err)
		return
	}

	fmt.Printf("📥 Decoded %d bytes of PCM audio\n", len(pcmData))

	// Convert bytes to int16 samples
	pcmSamples := make([]int16, len(pcmData)/2)
	for i := 0; i < len(pcmSamples); i++ {
		pcmSamples[i] = int16(binary.LittleEndian.Uint16(pcmData[i*2:]))
	}

	// Upsample from 24kHz to 48kHz
	pcm48kHz := upsample24to48(pcmSamples)

	// Encode PCM to Opus in 20ms frames (960 samples at 48kHz)
	const frameSize = 960
	for i := 0; i < len(pcm48kHz); i += frameSize {
		end := i + frameSize
		if end > len(pcm48kHz) {
			// Pad last frame with zeros
			frame := make([]int16, frameSize)
			copy(frame, pcm48kHz[i:])
			pcm48kHz = append(pcm48kHz[:i], frame...)
			end = len(pcm48kHz)
		}

		frame := pcm48kHz[i:end]

		// Encode to Opus
		opusData := make([]byte, 4000)
		n, err := opusEncoder.Encode(frame, opusData)
		if err != nil {
			log.Printf("❌ Opus encoding error: %v", err)
			continue
		}

		// Create RTP packet with proper timing
		timestampMutex.Lock()
		rtpPacket := &rtp.Packet{
			Header: rtp.Header{
				Version:        2,
				PayloadType:    111, // Opus
				SequenceNumber: sequenceNumber,
				Timestamp:      timestamp,
				SSRC:           1234,
			},
			Payload: opusData[:n],
		}
		sequenceNumber++
		timestamp += 960 // 20ms worth of samples at 48kHz
		timestampMutex.Unlock()

		if err := clientOutboundTrack.WriteRTP(rtpPacket); err != nil {
			log.Printf("❌ Error sending RTP packet: %v", err)
			break
		}
	}

	fmt.Printf("📤 Sent %d bytes of audio to client\n", len(pcmData))
}
