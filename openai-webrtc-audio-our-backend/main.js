import express from "express";
import cors from "cors";
import wrtc from "wrtc";
import { OPENAI_API_KEY } from "./key.js";

const app = express();
app.use(cors());
app.use(express.json());

// Global variables - single user
let clientPC = null;
let openaiPC = null;
let clientDataChannel = null;
let openaiDataChannel = null;

// ICE candidate endpoint
app.post("/ice-candidate", async (req, res) => {
  const { candidate } = req.body;

  if (clientPC && candidate) {
    try {
      await clientPC.addIceCandidate(new wrtc.RTCIceCandidate(candidate));
      console.log("✅ Added ICE candidate");
    } catch (error) {
      // Normal ICE errors, ignore
    }
  }

  res.json({ success: true });
});

// WebRTC offer endpoint
app.post("/webrtc-offer", async (req, res) => {
  const { offer } = req.body;

  try {
    console.log("🚀 Setting up WebRTC with auto VAD...");

    // Create peer connections
    clientPC = new wrtc.RTCPeerConnection();

    openaiPC = new wrtc.RTCPeerConnection();

    // ICE candidate forwarding
    clientPC.onicecandidate = (event) => {
      if (event.candidate) {
        openaiPC
          .addIceCandidate(new wrtc.RTCIceCandidate(event.candidate))
          .catch(() => {});
      }
    };

    openaiPC.onicecandidate = (event) => {
      if (event.candidate) {
        clientPC
          .addIceCandidate(new wrtc.RTCIceCandidate(event.candidate))
          .catch(() => {});
      }
    };

    // Connection monitoring
    clientPC.onconnectionstatechange = () => {
      console.log("📡 Client:", clientPC.connectionState);
    };

    openaiPC.onconnectionstatechange = () => {
      console.log("🤖 OpenAI:", openaiPC.connectionState);
    };

    // Audio forwarding: Client -> OpenAI
    clientPC.ontrack = (event) => {
      console.log("🎤 Audio: Client -> OpenAI");

      console.log("🎤 Audio: Client -> OpenAI");
      console.log("Track details:", {
        kind: event.track.kind,
        enabled: event.track.enabled,
        muted: event.track.muted,
        readyState: event.track.readyState,
        id: event.track.id,
        label: event.track.label,
      });

      if (event.track) {
        const stream = new wrtc.MediaStream([event.track]);
        openaiPC.addTrack(event.track, stream);
      }
    };

    // Audio forwarding: OpenAI -> Client
    openaiPC.ontrack = (event) => {
      console.log("🔊 Audio: OpenAI -> Client");
      console.log("🔊 Audio: OpenAI -> Client");
      console.log("OpenAI Track details:", {
        kind: event.track.kind,
        enabled: event.track.enabled,
        muted: event.track.muted,
        readyState: event.track.readyState,
        id: event.track.id,
        label: event.track.label,
      });
      if (event.track) {
        const stream = new wrtc.MediaStream([event.track]);
        clientPC.addTrack(event.track, stream);
      }
    };

    // Data channel setup
    clientPC.ondatachannel = (event) => {
      clientDataChannel = event.channel;
      console.log("📡 Client data channel received");

      clientDataChannel.onmessage = (event) => {
        const message = JSON.parse(event.data);
        console.log("📤 Clientss:", message.type);

        // Forward to OpenAI
        if (openaiDataChannel && openaiDataChannel.readyState === "open") {
          openaiDataChannel.send(event.data);
        }
      };
    };

    openaiDataChannel = openaiPC.createDataChannel("openai");
    openaiDataChannel.onopen = () => {
      console.log("🤖 OpenAI data channel open");
    };
    openaiDataChannel.onmessage = (event) => {
      const message = JSON.parse(event.data);
      console.log("📥 OpenAI:", message.type);

      if (message.type === "response.output_audio_transcript.delta") {
        console.log("📝 Transcript:", message.delta);
      }
    };

    // Handle client offer
    await clientPC.setRemoteDescription(new wrtc.RTCSessionDescription(offer));
    const clientAnswer = await clientPC.createAnswer();
    await clientPC.setLocalDescription(clientAnswer);

    // Create OpenAI offer
    const openaiOffer = await openaiPC.createOffer();
    await openaiPC.setLocalDescription(openaiOffer);

    // Get OpenAI token with VAD configuration
    console.log("🔑 Getting OpenAI token with VAD...");
    const tokenResponse = await fetch(
      "https://api.openai.com/v1/realtime/client_secrets",
      {
        method: "POST",
        headers: {
          Authorization: `Bearer ${OPENAI_API_KEY}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          session: {
            type: "realtime",
            model: "gpt-4o-mini-realtime-preview",
            audio: {
              output: {
                voice: "alloy",
              },
              input: {
                turn_detection: {
                  type: "server_vad",
                },
              },
            },
            instructions:
              "You are a helpful AI assistant. Respond naturally when you detect the user has finished speaking. Answer in ENGLISH only.",
          },
        }),
      }
    );

    if (!tokenResponse.ok) {
      throw new Error(`Token failed: ${tokenResponse.status}`);
    }

    const tokenData = await tokenResponse.json();
    const EPHEMERAL_KEY = tokenData.value;
    console.log("✅ VAD token received");

    // Connect to OpenAI
    console.log("🔗 Connecting to OpenAI with VAD...");
    const sdpResponse = await fetch(
      "https://api.openai.com/v1/realtime/calls?model=gpt-4o-mini-realtime-preview",
      {
        method: "POST",
        body: openaiOffer.sdp,
        headers: {
          Authorization: `Bearer ${EPHEMERAL_KEY}`,
          "Content-Type": "application/sdp",
        },
      }
    );

    if (!sdpResponse.ok) {
      throw new Error(`OpenAI failed: ${sdpResponse.status}`);
    }

    const openaiSDP = await sdpResponse.text();
    await openaiPC.setRemoteDescription(
      new wrtc.RTCSessionDescription({
        type: "answer",
        sdp: openaiSDP,
      })
    );

    console.log("✅ VAD setup complete - Auto speech detection enabled!");
    res.json({ answer: clientAnswer });
  } catch (error) {
    console.error("❌ Setup failed:", error);
    res.status(500).json({ error: error.message });
  }
});

// Cleanup endpoint
app.post("/cleanup", (req, res) => {
  if (clientPC) {
    clientPC.close();
    clientPC = null;
  }
  if (openaiPC) {
    openaiPC.close();
    openaiPC = null;
  }
  clientDataChannel = null;
  openaiDataChannel = null;

  console.log("🧹 Cleanup complete");
  res.json({ success: true });
});

app.listen(3000, () => {
  console.log("🚀 Auto VAD WebRTC server running on port 3000");
});



