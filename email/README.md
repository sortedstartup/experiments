# Email Sender - OCI Email Delivery

Simple Go program to send emails using Oracle Cloud Infrastructure (OCI) Email Delivery.

## Prerequisites

- Go 1.21 or later
- OCI Email Delivery configured with:
  - Verified domain
  - SMTP credentials generated
  - Approved sender email address

## Setup

1. Set environment variables:

```bash
export SMTP_USERNAME='ocid1.user.oc1..aaaaaa...'
export SMTP_PASSWORD='your-smtp-password'
export SMTP_FROM='sender@your-verified-domain.com'
export SMTP_TO='recipient1@example.com,recipient2@example.com'
```

**Important:** Use single quotes if password contains special characters like `$`, `!`, `@`.

2. Run:

```bash
go run send-email.go
```

Or build and run:

```bash
go build -o send-email send-email.go
./send-email
```

## Environment Variables

- `SMTP_USERNAME` - SMTP credential username (OCID format)
- `SMTP_PASSWORD` - SMTP credential password
- `SMTP_FROM` - Sender email (must be in Approved Senders)
- `SMTP_TO` - Recipient emails (comma-separated)

Note: Subject and body are defined in the code. Edit `send-email.go` to change them.

## Troubleshooting

**535 Authentication failed:**
- Verify SMTP credentials are correct
- Check password has no extra spaces (use single quotes)

**535 Authorization failed:**
- Ensure `SMTP_FROM` address is in OCI Approved Senders
- Check SMTP credentials and Approved Senders are in same compartment

## Security

Never commit SMTP credentials to version control. Always use environment variables.
