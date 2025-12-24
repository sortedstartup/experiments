package service

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenService struct {
	appJWTSecret []byte
	tokenTTL     time.Duration
	appIssuer    string
}

func NewTokenService(secret string, issuer string, ttl time.Duration) *TokenService {
	return &TokenService{
		appJWTSecret: []byte(secret),
		appIssuer:    issuer,
		tokenTTL:     ttl,
	}
}

func (s *TokenService) GenerateToken(userID, email, name, roles string) (string, error) {
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss":   s.appIssuer,
		"sub":   userID,
		"email": email,
		"roles": []string{roles},
		"iat":   now.Unix(),
		"exp":   now.Add(s.tokenTTL).Unix(),
		"name":  name,
	})

	return token.SignedString(s.appJWTSecret)
}
