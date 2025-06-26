package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	err := checkAndNotify()
	if err != nil {
		log.Fatalf("Error during check: %v", err)
	}
}

func checkAndNotify() error {
	file, err := os.Open("domains.txt")
	if err != nil {
		return fmt.Errorf("error opening domains.txt: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var expiring []string

	for scanner.Scan() {
		domain := strings.TrimSpace(scanner.Text())
		if domain == "" {
			continue
		}

		expiry, err := CheckDomainExpiry(domain)
		if err != nil {
			log.Println("WHOIS error for domain", domain, ":", err)
			continue
		}

		daysLeft := int(time.Until(expiry).Hours() / 24)
		log.Printf("Checked domain: %s | Expiry: %s | Days Left: %d\n", domain, expiry.Format("2006-01-02"), daysLeft)

		if daysLeft <= 30 {
			expiring = append(expiring, fmt.Sprintf("%s is expiring in %d days (%s)", domain, daysLeft, expiry.Format("2006-01-02")))
		}
	}

	if len(expiring) > 0 {
		log.Println("Expiring domains found. Sending alert email...")

		body := "The following domains are expiring soon:\n\n" + strings.Join(expiring, "\n")
		err := sendEmail(os.Getenv("ALERT_RECIPIENT"), "Domain Expiry Alert", body)
		if err != nil {
			return fmt.Errorf("error sending email: %w", err)
		}
		log.Println("Email sent successfully.")
	} else {
		log.Println("No expiring domains found.")
	}

	return nil
}
