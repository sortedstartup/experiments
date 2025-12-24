package service

import (
	"context"
	"fmt"
	"html/template"
	"log/slog"
)

type AuthService interface {
	LoginUser(ctx context.Context, email, name, roles, oAuthProvider, oAuthUserID string, isFederated bool) (string, error)
	GetOAuthConfigForFrontend() OAuthFrontendConfig
	GetProvider(name string) (*OAuthProvider, bool)
	VerifyAndExchangeCode(ctx context.Context, providerName, code string) (string, error)
	VerifyIDToken(ctx context.Context, providerName, rawIDToken string) (string, error)
	GetCallbackTemplate() *template.Template
}

type authService struct {
	providers    map[string]*OAuthProvider
	tokenService *TokenService
	userService  *UserService

	callbackResponseTemplate *template.Template
}

func NewAuthService(userService *UserService, tokenService *TokenService, providers map[string]*OAuthProvider) AuthService {
	slog.Debug("authservice:service:NewAuthService")

	// Create the callback HTML template
	callbackHTML := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Authentication Success - SortedChat</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', sans-serif;
            background: #f9fafb;
            margin: 0;
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
        }
        .container {
            text-align: center;
            color: #374151;
        }
        h1 {
            margin-bottom: 0.5rem;
            font-size: 1.5rem;
        }
        p {
            color: #6b7280;
            margin: 0;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>Authentication Successful!</h1>
        <p>Redirecting...</p>
    </div>
    
    <script>
        // Set JWT token in localStorage and redirect immediately
        localStorage.setItem('sortedchat.jwt', '{{.JWT}}');
        console.log('JWT token set in localStorage');
        window.location.href = '/';
    </script>
</body>
</html>`

	callbackTemplate, err := template.New("callback").Parse(callbackHTML)
	if err != nil {
		slog.Error("authservice:service:NewAuthService", "step", "failed to parse callback template", "error", err)
	}

	return &authService{
		userService:              userService,
		tokenService:             tokenService,
		providers:                providers,
		callbackResponseTemplate: callbackTemplate,
	}
}

// LoginUser handles the core login logic: verifying email, creating user, and issuing JWT
func (s *authService) LoginUser(ctx context.Context, email, name, roles, oAuthProvider, oAuthUserID string, isFederated bool) (string, error) {
	slog.Info("authservice:service:LoginUser", "email", email)

	// Check if email is in the allowlist
	// Note: isEmailAllowed should probably be injected or moved to a policy service,
	// but keeping it here for now as it was a helper function.
	// Assuming isEmailAllowed is defined in this package (it was used in the original file).
	if !isEmailAllowed(email) {
		slog.Debug("authservice:service:LoginUser", "step", "Login attempt from unauthorized email", "email", email)
		return "", fmt.Errorf("access denied: email not authorized")
	}

	userID, err := s.userService.CreateUserIfNotExists(email, name, roles, oAuthProvider, oAuthUserID, isFederated)
	if err != nil {
		slog.Error("authservice:service:LoginUser", "step", "user creation failed", "error", err)
		return "", fmt.Errorf("user creation failed: %w", err)
	}

	// Mint your app JWT using TokenService
	appJWT, err := s.tokenService.GenerateToken(userID, email, name, roles)
	if err != nil {
		slog.Error("authservice:service:LoginUser", "step", "jwt issue failed", "error", err)
		return "", fmt.Errorf("jwt issue failed: %w", err)
	}

	return appJWT, nil
}

type OAuthFrontendConfig struct {
	Providers map[string]ProviderFrontendConfig `json:"providers"`
}

type ProviderFrontendConfig struct {
	ClientID         string `json:"client_id"`
	OAuthProviderURL string `json:"oauth_provider_url"` // For frontend logic if needed
	OAuthRedirectURL string `json:"oauth_redirect_url"`
}

func (s *authService) GetOAuthConfigForFrontend() OAuthFrontendConfig {
	slog.Info("GetOAuthConfigForFrontend")

	configs := make(map[string]ProviderFrontendConfig)

	for name, p := range s.providers {
		configs[name] = ProviderFrontendConfig{
			ClientID:         p.Config.ClientID,
			OAuthRedirectURL: p.Config.RedirectURL,
		}
	}

	return OAuthFrontendConfig{
		Providers: configs,
	}
}

// GetProvider returns the OAuthProvider by name
func (s *authService) GetProvider(name string) (*OAuthProvider, bool) {
	p, ok := s.providers[name]
	return p, ok
}

// VerifyAndExchangeCode exchanges an OAuth code for a token, verifies the ID token, and logs the user in.
func (s *authService) VerifyAndExchangeCode(ctx context.Context, providerName, code string) (string, error) {
	provider, exists := s.GetProvider(providerName)
	if !exists {
		return "", fmt.Errorf("unknown provider: %s", providerName)
	}

	// Exchange code for token
	tok, err := provider.Config.Exchange(ctx, code)
	if err != nil {
		return "", fmt.Errorf("code exchange failed: %w", err)
	}

	// Extract ID Token
	rawIDToken, ok := tok.Extra("id_token").(string)
	if !ok || rawIDToken == "" {
		return "", fmt.Errorf("no id_token in token response")
	}

	return s.verifyIDTokenAndLogin(ctx, provider, rawIDToken)
}

// VerifyIDToken verifies a raw ID token (e.g. from Google One Tap) and logs the user in.
func (s *authService) VerifyIDToken(ctx context.Context, providerName, rawIDToken string) (string, error) {
	provider, exists := s.GetProvider(providerName)
	if !exists {
		return "", fmt.Errorf("unknown provider: %s", providerName)
	}

	return s.verifyIDTokenAndLogin(ctx, provider, rawIDToken)
}

func (s *authService) verifyIDTokenAndLogin(ctx context.Context, provider *OAuthProvider, rawIDToken string) (string, error) {
	// Verify ID token
	idToken, err := provider.Verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return "", fmt.Errorf("id token verification failed: %w", err)
	}

	// Extract claims
	var claims struct {
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return "", fmt.Errorf("claims extraction failed: %w", err)
	}

	if claims.Sub == "" {
		return "", fmt.Errorf("no subject in claims")
	}

	// Login user
	return s.LoginUser(ctx, claims.Email, claims.Name, "user", provider.Name, claims.Sub, true)
}

// GetCallbackTemplate returns the compiled HTML template for callback response
func (s *authService) GetCallbackTemplate() *template.Template {
	return s.callbackResponseTemplate
}

// Helper functions
// isEmailAllowed, getDefaults, getEnvOrDefault are defined in other files in this package.
