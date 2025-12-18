package main

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"os"
	"strings"
)

func main() {
	smtpHost := "smtp.email.ap-hyderabad-1.oci.oraclecloud.com"
	smtpPort := "587"

	username := strings.TrimSpace(os.Getenv("SMTP_USERNAME"))
	if username == "" {
		fmt.Println("Error: SMTP_USERNAME not set")
		os.Exit(1)
	}

	password := strings.TrimSpace(os.Getenv("SMTP_PASSWORD"))
	if password == "" {
		fmt.Println("Error: SMTP_PASSWORD not set")
		os.Exit(1)
	}

	from := strings.TrimSpace(os.Getenv("SMTP_FROM"))
	if from == "" {
		fmt.Println("Error: SMTP_FROM not set")
		os.Exit(1)
	}

	toEnv := strings.TrimSpace(os.Getenv("SMTP_TO"))
	if toEnv == "" {
		fmt.Println("Error: SMTP_TO not set")
		os.Exit(1)
	}

	to := strings.Split(toEnv, ",")
	for i := range to {
		to[i] = strings.TrimSpace(to[i])
	}

	subject := "Test Email from OCI"
	body := "This is a test email sent via Golang using OCI Email Delivery."

	toList := strings.Join(to, ", ")
	msg := []byte("From: " + from + "\r\n" +
		"To: " + toList + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n" +
		"\r\n" +
		body)

	auth := smtp.PlainAuth("", username, password, smtpHost)

	client, err := smtp.Dial(smtpHost + ":" + smtpPort)
	if err != nil {
		fmt.Printf("Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	if ok, _ := client.Extension("STARTTLS"); ok {
		config := &tls.Config{
			ServerName: smtpHost,
		}
		if err = client.StartTLS(config); err != nil {
			fmt.Printf("Failed to start TLS: %v\n", err)
			os.Exit(1)
		}
	}

	if err = client.Auth(auth); err != nil {
		fmt.Printf("Authentication failed: %v\n", err)
		os.Exit(1)
	}

	if err = client.Mail(from); err != nil {
		fmt.Printf("Failed to set sender: %v\n", err)
		os.Exit(1)
	}

	for _, recipient := range to {
		if err = client.Rcpt(recipient); err != nil {
			fmt.Printf("Failed to set recipient %s: %v\n", recipient, err)
			os.Exit(1)
		}
	}

	writer, err := client.Data()
	if err != nil {
		fmt.Printf("Failed to open data writer: %v\n", err)
		os.Exit(1)
	}

	_, err = writer.Write(msg)
	if err != nil {
		fmt.Printf("Failed to write message: %v\n", err)
		os.Exit(1)
	}

	if err = writer.Close(); err != nil {
		fmt.Printf("Failed to close writer: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Email sent successfully!")
}