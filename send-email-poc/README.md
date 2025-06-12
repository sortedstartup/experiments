# Email Sender App using SendGrid (Go)
A simple API in Go that sends emails using [SendGrid](https://sendgrid.com/) via a POST request.

## Features
- Send email via SendGrid
- Accepts recipient, subject, and body
- .env-based config
- Clean and extensible

## Setup Instructions

### Test Sending Email
```bash
curl -X POST http://localhost:8080/send-email \
  -H "Content-Type: application/json" \
  -d '{
    "to": "you@example.com",
    "subject": "Welcome to Sorted Stream",
    "body": "Hello and welcome! This is an email sent via Go using SendGrid."
  }'
```

### How to Get a SendGrid API Key
1. Sign up at [SendGrid](https://sendgrid.com/)
2. Go to Settings > API Keys
3. Click Create API Key, give it a name, and choose Full Access
4. Copy the key and paste it into .env

### .env.example
SENDGRID_API_KEY=your_sendgrid_api_key
EMAIL_SENDER=your_verified_sender@example.com

