package main

import (
	"fmt"
	"os"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// sendWithSendGrid sends an email using SendGrid
func sendWithSendGrid(to, subject, body string) error {
	apiKey := os.Getenv("SENDGRID_API_KEY")
	sender := os.Getenv("EMAIL_SENDER")

	if apiKey == "" || sender == "" {
		return fmt.Errorf("SendGrid API key or sender email not set")
	}

	from := mail.NewEmail("Domain Monitor", sender)
	toEmail := mail.NewEmail("", to)
	message := mail.NewSingleEmail(from, subject, toEmail, body, body)

	client := sendgrid.NewSendClient(apiKey)
	response, err := client.Send(message)

	if err != nil {
		return fmt.Errorf("SendGrid send error: %v", err)
	}

	if response.StatusCode >= 400 {
		return fmt.Errorf("SendGrid error: %s", response.Body)
	}

	return nil
}
