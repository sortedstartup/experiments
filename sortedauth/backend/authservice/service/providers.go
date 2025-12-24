package service

import (
	"context"
	"log/slog"

	"sortedstartup/authservice/dao"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// OAuthProvider holds the runtime configuration for a single OAuth provider
type OAuthProvider struct {
	Config   oauth2.Config
	Verifier *oidc.IDTokenVerifier
	Name     string
}

// InitProviders initializes OAuth providers from configuration
func InitProviders(config *dao.Config) (map[string]*OAuthProvider, error) {
	ctx := context.Background()
	providers := make(map[string]*OAuthProvider)

	if config == nil || len(config.OAuth.Providers) == 0 {
		slog.Warn("No OAuth providers configured in dao.Config")
		return providers, nil
	}

	for name, pCfg := range config.OAuth.Providers {
		slog.Info("Initializing OAuth provider", "name", name, "issuer", pCfg.IssuerURL)

		provider, err := oidc.NewProvider(ctx, pCfg.IssuerURL)
		if err != nil {
			slog.Error("Failed to create OIDC provider", "name", name, "error", err)
			continue
		}

		verifier := provider.Verifier(&oidc.Config{ClientID: pCfg.ClientID})

		oauthCfg := oauth2.Config{
			ClientID:     pCfg.ClientID,
			ClientSecret: pCfg.ClientSecret,
			RedirectURL:  pCfg.RedirectURL,
			Endpoint:     provider.Endpoint(),
			Scopes:       pCfg.Scopes,
		}

		// Override endpoints if manually specified (e.g. for non-OIDC providers)
		if pCfg.AuthURL != "" && pCfg.TokenURL != "" {
			oauthCfg.Endpoint = oauth2.Endpoint{
				AuthURL:  pCfg.AuthURL,
				TokenURL: pCfg.TokenURL,
			}
		}

		providers[name] = &OAuthProvider{
			Config:   oauthCfg,
			Verifier: verifier,
			Name:     name,
		}
	}

	return providers, nil
}

// BuildLoginURL constructs the OAuth authorization URL with the given state.
// The state is used to prevent CSRF attacks.
func (p *OAuthProvider) BuildLoginURL(state string) string {
	return p.Config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}
