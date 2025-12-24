package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"sortedstartup/authservice/service"
)

type AuthAPI struct {
	mux     *http.ServeMux
	service service.AuthService
}

func NewAuthAPI(mux *http.ServeMux, service service.AuthService) *AuthAPI {
	slog.Debug("authservice:api:NewAuthAPI")
	return &AuthAPI{
		mux:     mux,
		service: service,
	}
}

func (a *AuthAPI) initCore() {
	slog.Debug("authservice:api:initCore")
	//main end point, called by oauth providers like google etc.
	a.mux.HandleFunc("/callback/{provider}", a.OAuthCallbackHandler)
	a.mux.HandleFunc("/login", a.loginHandler)
	a.mux.HandleFunc("/oauth-config", a.oAuthConfigHandler)
	a.mux.HandleFunc("/google-one-tap-callback", a.GoogleOneTapHandler)
}

func (a *AuthAPI) oAuthConfigHandler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("authservice:api:oAuthConfigHandler", "path", r.URL.Path, "method", r.Method)
	// Get the OAuth configuration from the service
	config := a.service.GetOAuthConfigForFrontend()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(config); err != nil {
		slog.Error("authservice:api:oAuthConfigHandler encode failed", "err", err)
		http.Error(w, "failed to encode config", http.StatusInternalServerError)
		return
	}
}

func (a *AuthAPI) loginHandler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("authservice:api:loginHandler", "path", r.URL.Path, "method", r.Method)

	// Get the OAuth config to know which providers are available
	config := a.service.GetOAuthConfigForFrontend()

	// Simple HTML login page with dynamic buttons
	htmlHead := `<!DOCTYPE html>
<html>
<head>
    <title>Login - SortedChat</title>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            margin: 0;
            padding: 0;
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
        }
        .login-container {
            background: white;
            padding: 2rem;
            border-radius: 10px;
            box-shadow: 0 10px 25px rgba(0,0,0,0.1);
            text-align: center;
            max-width: 400px;
            width: 100%;
        }
        h1 {
            color: #333;
            margin-bottom: 1.5rem;
        }
        .auth-btn {
            display: inline-flex;
            align-items: center;
            justify-content: center;
            background: #f0f0f0;
            color: #333;
            border: 1px solid #ddd;
            padding: 12px 24px;
            border-radius: 5px;
            font-size: 16px;
            font-weight: 500;
            text-decoration: none;
            transition: all 0.3s;
            cursor: pointer;
            margin-bottom: 10px;
            width: 100%;
            box-sizing: border-box;
        }
        .auth-btn:hover {
            background: #e0e0e0;
        }
        .auth-btn.google {
            background: #4285f4;
            color: white;
            border: none;
        }
        .auth-btn.google:hover {
            background: #3367d6;
        }
        .auth-btn svg {
            margin-right: 8px;
        }
    </style>
</head>
<body>
    <div class="login-container">
        <h1>Welcome to SortedChat</h1>
        <p>Sign in to continue</p>
`
	htmlFoot := `    </div>
</body>
</html>`

	var buttonsHtml string

	// If no providers, show a message
	if len(config.Providers) == 0 {
		buttonsHtml = "<p>No authentication providers configured.</p>"
	} else {
		for name := range config.Providers {
			// Get provider to build URL
			provider, exists := a.service.GetProvider(name)
			if !exists {
				continue
			}
			// TODO: Use a real random state for CSRF protection
			authURL := provider.BuildLoginURL("state")

			// Customize button based on provider name
			btnClass := "auth-btn"
			btnText := "Sign in with " + name
			icon := "" // Default no icon

			if name == "google" {
				btnClass += " google"
				btnText = "Sign in with Google"
				icon = `<svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
                <path d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z" fill="#4285F4"/>
                <path d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" fill="#34A853"/>
                <path d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z" fill="#FBBC05"/>
                <path d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" fill="#EA4335"/>
            </svg>`
			}

			buttonsHtml += `<a href="` + authURL + `" class="` + btnClass + `">` + icon + btnText + `</a>`
		}
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(htmlHead + buttonsHtml + htmlFoot))
}

func (a *AuthAPI) OAuthCallbackHandler(w http.ResponseWriter, r *http.Request) {
	slog.Info("authservice:api:OAuthCallbackHandler")
	ctx := r.Context()
	q := r.URL.Query()

	code := q.Get("code")

	// Extract provider from path: /callback/{provider}
	providerName := r.PathValue("provider")

	if providerName == "" {
		slog.Error("Provider not specified in path", "path", r.URL.Path)
		http.Error(w, "Provider not specified", http.StatusBadRequest)
		return
	}

	// Exchange code and verify token via Service
	appJWT, err := a.service.VerifyAndExchangeCode(ctx, providerName, code)
	if err != nil {
		slog.Error("authservice:api:OAuthCallbackHandler", "step", "VerifyAndExchangeCode failed", "error", err)
		if strings.Contains(err.Error(), "access denied") {
			http.Error(w, "Access denied: Your email is not authorized", http.StatusForbidden)
		} else {
			http.Error(w, "Login failed", http.StatusInternalServerError)
		}
		return
	}

	// Return HTML page with JWT embedded in JavaScript
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	data := struct {
		JWT string
	}{
		JWT: appJWT,
	}

	if err := a.service.GetCallbackTemplate().Execute(w, data); err != nil {
		http.Error(w, "template execution failed", http.StatusInternalServerError)
		return
	}
}

type TokenReq struct {
	Token string `json:"token"`
}

func (a *AuthAPI) GoogleOneTapHandler(w http.ResponseWriter, r *http.Request) {
	var req TokenReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("authservice:api:GoogleOneTapHandler", "step", "failed to decode request", "error", err)
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Verify token via Service
	appJWT, err := a.service.VerifyIDToken(ctx, "google", req.Token)
	if err != nil {
		slog.Error("authservice:api:GoogleOneTapHandler", "step", "VerifyIDToken failed", "error", err)
		if strings.Contains(err.Error(), "access denied") {
			http.Error(w, "Access denied: Your email is not authorized", http.StatusForbidden)
		} else {
			http.Error(w, "Login failed: "+err.Error(), http.StatusUnauthorized)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "jwt": appJWT})
}
