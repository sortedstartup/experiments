package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v4"
)

type OpenAIRealtimeProxy struct {
	peerConnection *webrtc.PeerConnection
	openAIConn     *websocket.Conn
	audioTrack     *webrtc.TrackLocalStaticRTP
	mu             sync.Mutex
	sequenceNumber uint16
	timestamp      uint32
}

func main() {
	proxy := &OpenAIRealtimeProxy{}

	http.HandleFunc("/offer", proxy.handleOffer)
	http.HandleFunc("/", serveIndex)

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func (p *OpenAIRealtimeProxy) handleOffer(w http.ResponseWriter, r *http.Request) {
	var offer webrtc.SessionDescription
	if err := json.NewDecoder(r.Body).Decode(&offer); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Create WebRTC peer connection
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{URLs: []string{"stun:stun.l.google.com:19302"}},
		},
	}

	var err error
	p.peerConnection, err = webrtc.NewPeerConnection(config)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create audio track for sending audio to client
	p.audioTrack, err = webrtc.NewTrackLocalStaticRTP(
		webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus},
		"audio", "openai-proxy",
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Add audio track to peer connection
	if _, err = p.peerConnection.AddTrack(p.audioTrack); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Handle incoming audio from client
	p.peerConnection.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		log.Printf("Received track: %s", track.Kind())

		if track.Kind() == webrtc.RTPCodecTypeAudio {
			go p.handleIncomingAudio(track)
		}
	})

	// Set remote description (offer from client)
	if err = p.peerConnection.SetRemoteDescription(offer); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create answer
	answer, err := p.peerConnection.CreateAnswer(nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set local description
	if err = p.peerConnection.SetLocalDescription(answer); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Connect to OpenAI Realtime API
	go p.connectToOpenAI()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(answer)
}

func (p *OpenAIRealtimeProxy) connectToOpenAI() {
	// First, get ephemeral token from OpenAI
	token, err := p.getEphemeralToken()
	if err != nil {
		log.Printf("Failed to get ephemeral token: %v", err)
		return
	}

	// Connect to OpenAI WebSocket endpoint
	url := "wss://api.openai.com/v1/realtime?model=gpt-4o-realtime-preview-2024-10-01"

	header := http.Header{}
	header.Set("Authorization", "Bearer "+token)
	header.Set("OpenAI-Beta", "realtime=v1")

	conn, _, err := websocket.DefaultDialer.Dial(url, header)
	if err != nil {
		log.Printf("Failed to connect to OpenAI: %v", err)
		return
	}
	defer conn.Close()

	p.mu.Lock()
	p.openAIConn = conn
	p.mu.Unlock()

	// Send session configuration
	sessionConfig := map[string]interface{}{
		"type": "session.update",
		"session": map[string]interface{}{
			"modalities":   []string{"text", "audio"},
			"instructions": "You are a helpful assistant. Respond naturally to voice conversations.",
			"voice":        "alloy",
			"turn_detection": map[string]interface{}{
				"type": "server_vad",
			},
		},
	}

	if err := conn.WriteJSON(sessionConfig); err != nil {
		log.Printf("Failed to send session config: %v", err)
		return
	}

	// Handle messages from OpenAI
	p.handleOpenAIMessages(conn)
}

func (p *OpenAIRealtimeProxy) getEphemeralToken() (string, error) {
	// Replace with your actual OpenAI API key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("OPENAI_API_KEY environment variable not set")
	}

	// In production, you'd make an HTTP request to OpenAI's session endpoint
	// to get an ephemeral token. For now, return the API key
	return apiKey, nil
}

func (p *OpenAIRealtimeProxy) handleIncomingAudio(track *webrtc.TrackRemote) {
	log.Println("Started handling incoming audio from client")

	for {
		rtpPacket, _, err := track.ReadRTP()
		if err != nil {
			log.Printf("Error reading RTP: %v", err)
			return
		}

		// Forward audio data to OpenAI
		p.forwardAudioToOpenAI(rtpPacket.Payload)
	}
}

func (p *OpenAIRealtimeProxy) forwardAudioToOpenAI(audioData []byte) {
	p.mu.Lock()
	conn := p.openAIConn
	p.mu.Unlock()

	if conn == nil {
		return
	}

	// Send audio to OpenAI
	audioMessage := map[string]interface{}{
		"type":  "input_audio_buffer.append",
		"audio": audioData, // In practice, you'd encode this properly as base64
	}

	if err := conn.WriteJSON(audioMessage); err != nil {
		log.Printf("Failed to send audio to OpenAI: %v", err)
	}
}

func (p *OpenAIRealtimeProxy) handleOpenAIMessages(conn *websocket.Conn) {
	for {
		var message map[string]interface{}
		if err := conn.ReadJSON(&message); err != nil {
			log.Printf("Error reading from OpenAI: %v", err)
			return
		}

		messageType, ok := message["type"].(string)
		if !ok {
			continue
		}

		switch messageType {
		case "response.audio.delta":
			// Handle audio response from OpenAI
			if audioData, ok := message["delta"].(string); ok {
				p.sendAudioToClient([]byte(audioData))
			}
		case "response.audio_transcript.delta":
			log.Printf("OpenAI transcript: %v", message["delta"])
		case "error":
			log.Printf("OpenAI error: %v", message["error"])
		default:
			log.Printf("Received message type: %s", messageType)
		}
	}
}

func (p *OpenAIRealtimeProxy) sendAudioToClient(audioData []byte) {
	if p.audioTrack == nil {
		return
	}

	// Create RTP packet with audio data using the pion/rtp package
	p.sequenceNumber++
	p.timestamp += 960 // 20ms of audio at 48kHz (typical for Opus)

	packet := &rtp.Packet{
		Header: rtp.Header{
			Version:        2,
			PayloadType:    111, // Opus payload type
			SequenceNumber: p.sequenceNumber,
			Timestamp:      p.timestamp,
			SSRC:           12345,
		},
		Payload: audioData,
	}

	if err := p.audioTrack.WriteRTP(packet); err != nil {
		log.Printf("Failed to send audio to client: %v", err)
	}
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	html := `
<!DOCTYPE html>
<html>
<head>
    <title>OpenAI WebRTC Audio Proxy</title>
</head>
<body>
    <h1>OpenAI Realtime Audio</h1>
    <button id="start">Start Audio Chat</button>
    <button id="stop" disabled>Stop</button>
    <div id="status">Ready to connect</div>
    <audio id="remoteAudio" autoplay></audio>
    
    <script>
        const startButton = document.getElementById('start');
        const stopButton = document.getElementById('stop');
        const status = document.getElementById('status');
        const remoteAudio = document.getElementById('remoteAudio');
        let peerConnection;
        let localStream;

        startButton.onclick = async () => {
            try {
                status.textContent = 'Getting microphone access...';
                
                // Get user media (microphone)
                localStream = await navigator.mediaDevices.getUserMedia({ 
                    audio: {
                        echoCancellation: true,
                        noiseSuppression: true,
                        autoGainControl: true
                    }
                });
                
                status.textContent = 'Creating connection...';
                
                // Create peer connection
                peerConnection = new RTCPeerConnection({
                    iceServers: [{ urls: 'stun:stun.l.google.com:19302' }]
                });

                // Add local stream to peer connection
                localStream.getTracks().forEach(track => {
                    peerConnection.addTrack(track, localStream);
                });

                // Handle remote stream
                peerConnection.ontrack = event => {
                    remoteAudio.srcObject = event.streams[0];
                    status.textContent = 'Connected! Speak into your microphone.';
                };

                // Handle connection state changes
                peerConnection.onconnectionstatechange = () => {
                    console.log('Connection state:', peerConnection.connectionState);
                    if (peerConnection.connectionState === 'connected') {
                        status.textContent = 'Connected to OpenAI! Speak into your microphone.';
                    }
                };

                // Create offer
                const offer = await peerConnection.createOffer();
                await peerConnection.setLocalDescription(offer);

                status.textContent = 'Connecting to server...';

                // Send offer to server
                const response = await fetch('/offer', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(offer)
                });

                if (!response.ok) {
                    throw new Error('Failed to connect to server');
                }

                const answer = await response.json();
                await peerConnection.setRemoteDescription(answer);

                startButton.disabled = true;
                stopButton.disabled = false;
                
            } catch (error) {
                console.error('Error:', error);
                status.textContent = 'Error: ' + error.message;
            }
        };

        stopButton.onclick = () => {
            if (peerConnection) {
                peerConnection.close();
            }
            if (localStream) {
                localStream.getTracks().forEach(track => track.stop());
            }
            startButton.disabled = false;
            stopButton.disabled = true;
            status.textContent = 'Disconnected';
        };
    </script>
</body>
</html>`
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}
