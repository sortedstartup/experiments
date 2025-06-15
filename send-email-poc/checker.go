package main

import (
	"fmt"
	"time"

	"github.com/likexian/whois"
	whoisparser "github.com/likexian/whois-parser"
)

func CheckDomainExpiry(domain string) (time.Time, error) {
	raw, err := whois.Whois(domain)
	if err != nil {
		return time.Time{}, fmt.Errorf("WHOIS error: %w", err)
	}

	result, err := whoisparser.Parse(raw)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse error: %w", err)
	}

	expiryStr := result.Domain.ExpirationDate
	if expiryStr == "" {
		return time.Time{}, fmt.Errorf("no expiration date found for domain: %s", domain)
	}

	// Try common date formats
	formats := []string{
		"2006-01-02T15:04:05Z", // ISO 8601
		"2006-01-02",           // Common plain format
		"02-Jan-2006",          // WHOIS format
		"2006.01.02",           // Some ccTLDs
		"2006/01/02",           // Just in case
	}

	var parsed time.Time
	for _, format := range formats {
		parsed, err = time.Parse(format, expiryStr)
		if err == nil {
			return parsed, nil
		}
	}

	return time.Time{}, fmt.Errorf("could not parse expiration date: %s", expiryStr)
}
