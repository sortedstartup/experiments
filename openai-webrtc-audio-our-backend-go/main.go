package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/pion/webrtc/v3"
	"github.com/rs/cors"
)

// Global variables - single user
var (
	clientPC          *webrtc.PeerConnection
	openaiPC          *webrtc.PeerConnection
	clientDataChannel *webrtc.DataChannel
	openaiDataChannel *webrtc.DataChannel
	OPENAI_API_KEY    = ""
	openAIConnected   = false
)

type ICECandidateRequest struct {
	Candidate webrtc.ICECandidateInit `json:"candidate"`
}

type WebRTCOfferRequest struct {
	Offer webrtc.SessionDescription `json:"offer"`
}

type WebRTCOfferResponse struct {
	Answer webrtc.SessionDescription `json:"answer"`
}

const htmlContent = `<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>OpenAI Realtime WebRTC - Simple</title>
  </head>
  <body>
    <h1>Realtime Audio with WebRTC - Simple</h1>
    <button id="connection">Connect With LLM</button>
    <div id="status">Click Connect to start</div>

    <script>
      let audioElement;
      let dataChannel;
      let isConnected = false;
      let pc = null;

      async function Connection() {
        try {
          document.getElementById("status").textContent = "Connecting...";
          pc = new RTCPeerConnection();

          pc.onicecandidate = (event) => {
            if (event.candidate) {
              fetch("/ice-candidate", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ candidate: event.candidate }),
              }).catch(() => {});
            }
          };

          pc.onconnectionstatechange = () => {
            console.log("Connection state:", pc.connectionState);
            document.getElementById("status").textContent = ` + "`Connection: ${pc.connectionState}`" + `;
          };

          pc.ontrack = (e) => {
            console.log("🔊 Audio received");
            if (!audioElement) {
              audioElement = document.createElement("audio");
              audioElement.autoplay = true;
              document.body.appendChild(audioElement);
            }
            audioElement.srcObject = e.streams[0];
          };

          // Add microphone
          const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
          stream.getTracks().forEach(track => {
            pc.addTrack(track, stream);
          });
          
          // Data channel
          dataChannel = pc.createDataChannel("oai-events");
          dataChannel.onopen = () => {
            console.log("✅ Data channel open");
            dataChannel.send(JSON.stringify({
              type: "session.update",
              session: {
                modalities: ["audio"],
                voice: "alloy",
                turn_detection: { type: "server_vad" },
                instructions: "You are helpful. Answer in ENGLISH only."
              }
            }));
          };

          dataChannel.onmessage = (event) => {
            const msg = JSON.parse(event.data);
            console.log("📨", msg.type);
            if (msg.type === "session.created") {
              document.getElementById("status").textContent = "✅ Ready - Speak!";
              isConnected = true;
            }
          };

          const offer = await pc.createOffer();
          await pc.setLocalDescription(offer);

          const response = await fetch("/webrtc-offer", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ offer: offer }),
          });

          if (!response.ok) {
            throw new Error(` + "`HTTP ${response.status}`" + `);
          }

          const { answer } = await response.json();
          await pc.setRemoteDescription(answer);

        } catch (error) {
          console.error("Failed:", error);
          document.getElementById("status").textContent = ` + "`Failed: ${error.message}`" + `;
        }
      }

      document.getElementById("connection").onclick = () => {
        if (!isConnected) Connection();
      };
    </script>
  </body>
</html>`

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(htmlContent))
	})
	mux.HandleFunc("/ice-candidate", handleICECandidate)
	mux.HandleFunc("/webrtc-offer", handleWebRTCOffer)

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
	json.NewDecoder(r.Body).Decode(&req)

	if clientPC != nil && req.Candidate.Candidate != "" {
		clientPC.AddICECandidate(req.Candidate)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func handleWebRTCOffer(w http.ResponseWriter, r *http.Request) {
	var req WebRTCOfferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	fmt.Println("🚀 Setting up WebRTC...")

	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{URLs: []string{"stun:stun.l.google.com:19302"}},
		},
	}

	var err error
	clientPC, err = webrtc.NewPeerConnection(config)
	if err != nil {
		http.Error(w, "Client PC failed", http.StatusInternalServerError)
		return
	}

	openaiPC, err = webrtc.NewPeerConnection(config)
	if err != nil {
		http.Error(w, "OpenAI PC failed", http.StatusInternalServerError)
		return
	}

	// ICE forwarding
	clientPC.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c != nil {
			openaiPC.AddICECandidate(c.ToJSON())
		}
	})

	openaiPC.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c != nil {
			clientPC.AddICECandidate(c.ToJSON())
		}
	})

	// Connection status
	clientPC.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		fmt.Printf("📡 Client: %s\n", s)
	})

	openaiPC.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		fmt.Printf("🤖 OpenAI: %s\n", s)
	})

	// THIS IS THE KEY: Just like Node.js - add tracks when received from client
	clientPC.OnTrack(func(remoteTrack *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		fmt.Println("🎤 Audio: Client -> OpenAI")
		fmt.Printf("Track details: kind=%s, id=%s\n", remoteTrack.Kind(), remoteTrack.ID())

		// Create local track for OpenAI (like Node.js does)
		localTrack, err := webrtc.NewTrackLocalStaticRTP(
			remoteTrack.Codec().RTPCodecCapability,
			remoteTrack.ID(),
			remoteTrack.StreamID(),
		)
		if err != nil {
			fmt.Printf("Error creating track: %v\n", err)
			return
		}

		// Add track to OpenAI peer connection (like Node.js addTrack)
		if _, err := openaiPC.AddTrack(localTrack); err != nil {
			fmt.Printf("Error adding track: %v\n", err)
			return
		}

		// Start relaying audio data
		go func() {
			rtpBuf := make([]byte, 1400)
			for {
				i, _, readErr := remoteTrack.Read(rtpBuf)
				if readErr != nil {
					if readErr == io.EOF {
						break
					}
					break
				}
				localTrack.Write(rtpBuf[:i])
			}
		}()

		// NOW connect to OpenAI after we have audio - just like Node.js!
		if !openAIConnected {
			connectToOpenAIWithAudio()
		}
	})

	// Audio from OpenAI -> Client
	openaiPC.OnTrack(func(remoteTrack *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		fmt.Println("🔊 Audio: OpenAI -> Client")

		localTrack, err := webrtc.NewTrackLocalStaticRTP(
			remoteTrack.Codec().RTPCodecCapability,
			remoteTrack.ID(),
			remoteTrack.StreamID(),
		)
		if err != nil {
			return
		}

		clientPC.AddTrack(localTrack)

		go func() {
			rtpBuf := make([]byte, 1400)
			for {
				i, _, readErr := remoteTrack.Read(rtpBuf)
				if readErr != nil {
					break
				}
				localTrack.Write(rtpBuf[:i])
			}
		}()
	})

	// Data channel setup
	clientPC.OnDataChannel(func(dc *webrtc.DataChannel) {
		clientDataChannel = dc
		fmt.Println("📡 Client data channel")

		dc.OnMessage(func(msg webrtc.DataChannelMessage) {
			if openaiDataChannel != nil {
				openaiDataChannel.Send(msg.Data)
			}
		})
	})

	openaiDataChannel, _ = openaiPC.CreateDataChannel("openai", nil)
	openaiDataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
		// Parse the JSON message
		var message map[string]interface{}
		if err := json.Unmarshal(msg.Data, &message); err != nil {
			fmt.Printf("📥 OpenAI (raw): %s\n", string(msg.Data))
			return
		}

		// Print the message type
		msgType, _ := message["type"].(string)
		fmt.Printf("📥 OpenAI Event: %s\n", msgType)

		// Print full message for detailed debugging
		prettyJSON, err := json.MarshalIndent(message, "", "  ")
		if err == nil {
			fmt.Printf("📥 OpenAI Full Message:\n%s\n", string(prettyJSON))
		}

		// Handle specific message types with detailed info
		switch msgType {
		case "session.created":
			fmt.Println("🎯 Session created successfully!")

		case "session.updated":
			fmt.Println("📋 Session configuration updated")

		case "input_audio_buffer.speech_started":
			fmt.Println("🎤 Speech detection: Started")

		case "input_audio_buffer.speech_stopped":
			fmt.Println("🔇 Speech detection: Stopped")

		case "response.created":
			fmt.Println("🤖 OpenAI is generating response")

		case "response.audio.delta":
			fmt.Println("🔊 Receiving audio chunk from OpenAI")

		case "response.audio_transcript.delta":
			if delta, ok := message["delta"].(string); ok {
				fmt.Printf("📝 AI Transcript: '%s'\n", delta)
			}

		case "response.output_audio_transcript.delta":
			if delta, ok := message["delta"].(string); ok {
				fmt.Printf("📝 AI Speaking: '%s'\n", delta)
			}

		case "response.done":
			fmt.Println("✅ Response completed")

		case "error":
			if errorMsg, ok := message["error"].(map[string]interface{}); ok {
				fmt.Printf("❌ OpenAI Error: %v\n", errorMsg)
			}

		default:
			fmt.Printf("❓ Unknown OpenAI event: %s\n", msgType)
		}

		fmt.Println("─────────────────────────────────────") // Separator line

		// Forward to client
		if clientDataChannel != nil && clientDataChannel.ReadyState() == webrtc.DataChannelStateOpen {
			clientDataChannel.Send(msg.Data)
		}
	})

	// Handle client offer/answer first
	clientPC.SetRemoteDescription(req.Offer)
	clientAnswer, _ := clientPC.CreateAnswer(nil)
	clientPC.SetLocalDescription(clientAnswer)

	fmt.Println("✅ Setup complete - waiting for audio from client")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(WebRTCOfferResponse{Answer: clientAnswer})
}

func connectToOpenAIWithAudio() {
	openAIConnected = true
	fmt.Println("🔗 Now connecting to OpenAI with audio...")

	// Create offer AFTER we have audio tracks (like Node.js)
	openaiOffer, err := openaiPC.CreateOffer(nil)
	if err != nil {
		fmt.Printf("Error creating OpenAI offer: %v\n", err)
		return
	}

	openaiPC.SetLocalDescription(openaiOffer)
	fmt.Printf("OpenAI Offer created with %d characters\n", len(openaiOffer.SDP))

	// Get token
	token, err := getOpenAIToken()
	if err != nil {
		fmt.Printf("Token error: %v\n", err)
		return
	}

	// Connect to OpenAI
	if err := connectToOpenAI(openaiOffer, token); err != nil {
		fmt.Printf("OpenAI error: %v\n", err)
		return
	}

	fmt.Println("✅ Connected to OpenAI!")
}

func getOpenAIToken() (string, error) {
	config := map[string]interface{}{
		"session": map[string]interface{}{
			"type":  "realtime",
			"model": "gpt-4o-mini-realtime-preview",
			"audio": map[string]interface{}{
				"output": map[string]interface{}{"voice": "alloy"},
				"input":  map[string]interface{}{"turn_detection": map[string]interface{}{"type": "server_vad"}},
			},
			"instructions": "You are helpful. Answer briefly in ENGLISH only.",
		},
	}

	data, _ := json.Marshal(config)
	req, _ := http.NewRequest("POST", "https://api.openai.com/v1/realtime/client_secrets", bytes.NewBuffer(data))
	req.Header.Set("Authorization", "Bearer "+OPENAI_API_KEY)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if value, ok := result["value"].(string); ok {
		return value, nil
	}
	return "", fmt.Errorf("no token received")
}

func connectToOpenAI(offer webrtc.SessionDescription, token string) error {
	req, _ := http.NewRequest("POST",
		"https://api.openai.com/v1/realtime/calls?model=gpt-4o-mini-realtime-preview",
		bytes.NewBufferString(offer.SDP))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/sdp")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	responseBody := buf.String()

	fmt.Printf("Response Status: %d\n", resp.StatusCode)
	fmt.Printf("Response Body Length: %d chars\n", len(responseBody))

	// Fix: Accept both 200 and 201 as success
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return fmt.Errorf("status %d: %s", resp.StatusCode, responseBody)
	}

	// Success! Set the SDP answer from OpenAI
	fmt.Println("✅ Got SDP answer from OpenAI!")
	return openaiPC.SetRemoteDescription(webrtc.SessionDescription{
		Type: webrtc.SDPTypeAnswer,
		SDP:  responseBody,
	})
}
