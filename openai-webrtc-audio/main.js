import express from "express";
import cors from "cors";

import { OPENAI_API_KEY } from "./key.js";

const app = express();
app.use(cors());
const sessionConfig = JSON.stringify({
    session: {
        type: "realtime",
        model: "gpt-4o-mini-realtime-preview",  
        audio: {
            input: {
                turn_detection: null
            },
            output: {
                voice: "marin",
            },
        },
        instructions: "You are a helpful assistant. Please answer the user's questions in ENGLISH only.",
    },
});

// An endpoint which would work with the client code above - it returns
// the contents of a REST API request to this protected endpoint
app.get("/token", async (req, res) => {
    try {
        const response = await fetch(
            "https://api.openai.com/v1/realtime/client_secrets",
            {
                method: "POST",
                headers: {
                    Authorization: `Bearer ${OPENAI_API_KEY}`,
                    "Content-Type": "application/json",
                },
                body: sessionConfig,
            }
        );

        const data = await response.json();
        res.json(data);
    } catch (error) {
        console.error("Token generation error:", error);
        res.status(500).json({ error: "Failed to generate token" });
    }
});

app.listen(3000);



// tokens endpoint returns something like this:
// {
//     "value": "ek_68c261c617ac8191baa41dbf5aa9b94e",
//     "expires_at": 1757570078,
//     "session": {
//         "type": "realtime",
//         "object": "realtime.session",
//         "id": "sess_CEUOkSHp7srjd1XsxJaBf",
//         "model": "gpt-4o-mini-realtime-preview",
//         "output_modalities": [
//             "audio"
//         ],
//         "instructions": "You are a helpful assistant. Please answer the user's questions in ENGLISH only.",
//         "tools": [],
//         "tool_choice": "auto",
//         "max_output_tokens": "inf",
//         "tracing": null,
//         "truncation": "auto",
//         "prompt": null,
//         "expires_at": 0,
//         "audio": {
//             "input": {
//                 "format": {
//                     "type": "audio/pcm",
//                     "rate": 24000
//                 },
//                 "transcription": null,
//                 "noise_reduction": null,
//                 "turn_detection": {
//                     "type": "server_vad",
//                     "threshold": 0.5,
//                     "prefix_padding_ms": 300,
//                     "silence_duration_ms": 200,
//                     "idle_timeout_ms": null,
//                     "create_response": true,
//                     "interrupt_response": true
//                 }
//             },
//             "output": {
//                 "format": {
//                     "type": "audio/pcm",
//                     "rate": 24000
//                 },
//                 "voice": "marin",
//                 "speed": 1
//             }
//         },
//         "include": null
//     }
// }