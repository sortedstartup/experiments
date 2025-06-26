package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() { // run only for testing for application run rename it and use the main function of main.go
	// Load env
	err := godotenv.Load()
	if err != nil {
		log.Println("Could not load .env, using environment variables")
	}

	to := os.Getenv("ALERT_RECIPIENT")
	subject := "Test Email - Domain Monitor"
	body := "This is a test email to verify your SendGrid " + os.Getenv("EMAIL_METHOD") + " setup."

	err = sendEmail(to, subject, body)
	if err != nil {
		log.Fatalf("Failed to send test email: %v", err)
	}
	log.Println("Test email sent successfully to", to)
}
