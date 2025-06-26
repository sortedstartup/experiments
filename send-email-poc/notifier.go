package main

import (
	"fmt"
	"net/smtp"
	"os"
	"strings"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// API-based SendGrid sender
func sendWithSendGrid(to, subject, body string) error {
	apiKey := os.Getenv("SENDGRID_API_KEY")
	sender := os.Getenv("EMAIL_SENDER")

	if apiKey == "" || sender == "" {
		return fmt.Errorf("SendGrid API key or sender email not set")
	}

	name := os.Getenv("EMAIL_SENDER_NAME")
	if name == "" {
		name = "Domain Monitor"
	}

	from := mail.NewEmail(name, sender)
	toEmail := mail.NewEmail("", to)
	message := mail.NewSingleEmail(from, subject, toEmail, body, body)

	client := sendgrid.NewSendClient(apiKey)
	response, err := client.Send(message)

	if err != nil {
		return fmt.Errorf("SendGrid send error: %v", err)
	}

	if response.StatusCode >= 400 {
		return fmt.Errorf("SendGrid error (%d): %s", response.StatusCode, response.Body)
	}

	return nil
}

// SMTP-based sender
func sendWithSMTP(to, subject, body string) error {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	username := os.Getenv("SMTP_USERNAME")
	password := os.Getenv("SMTP_PASSWORD")
	sender := os.Getenv("EMAIL_SENDER")

	if host == "" || port == "" || username == "" || password == "" || sender == "" {
		return fmt.Errorf("SMTP credentials are not set properly")
	}

	addr := host + ":" + port
	auth := smtp.PlainAuth("", username, password, host)

	msg := []byte("From: " + sender + "\r\n" + 
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"Content-Type: text/plain; charset=\"UTF-8\"\r\n" +
		"\r\n" +
		body + "\r\n")

	err := smtp.SendMail(addr, auth, sender, []string{to}, msg)
	if err != nil {
		return fmt.Errorf("SMTP send error: %v", err)
	}
	return nil
}

// Unified sender
func sendEmail(to, subject, body string) error {
	method := strings.ToLower(os.Getenv("EMAIL_METHOD"))
	switch method {
	case "smtp":
		return sendWithSMTP(to, subject, body)
	case "api", "":
		return sendWithSendGrid(to, subject, body)
	default:
		return fmt.Errorf("unknown email method: %s", method)
	}
}
