# How to Run Auth Service

## Prerequisites

- Go 1.22 or later
- SQLite (optional, for inspecting the database)

## Configuration

The service uses environment variables for configuration. You need to set up your OAuth providers using the following format:

```bash
export OAUTH_PROVIDER_<NAME>_CLIENT_ID="your-client-id"
export OAUTH_PROVIDER_<NAME>_CLIENT_SECRET="your-client-secret"
export OAUTH_PROVIDER_<NAME>_REDIRECT_URL="http://localhost:8080/callback/<name>"
export OAUTH_PROVIDER_<NAME>_ISSUER_URL="https://issuer.url" # Required for OIDC
export OAUTH_PROVIDER_<NAME>_SCOPES="openid email profile" # Space-separated scopes
# Optional: Manual endpoints for non-OIDC providers
# export OAUTH_PROVIDER_<NAME>_AUTH_URL="https://..."
# export OAUTH_PROVIDER_<NAME>_TOKEN_URL="https://..."
```

### Example: Google

```bash
export OAUTH_PROVIDER_GOOGLE_CLIENT_ID="your-google-client-id"
export OAUTH_PROVIDER_GOOGLE_CLIENT_SECRET="your-google-client-secret"
export OAUTH_PROVIDER_GOOGLE_REDIRECT_URL="http://localhost:8080/callback/google"
export OAUTH_PROVIDER_GOOGLE_ISSUER_URL="https://accounts.google.com"
export OAUTH_PROVIDER_GOOGLE_SCOPES="openid email profile"
```

### Example: GitHub

```bash
export OAUTH_PROVIDER_GITHUB_CLIENT_ID="your-github-client-id"
export OAUTH_PROVIDER_GITHUB_CLIENT_SECRET="your-github-client-secret"
export OAUTH_PROVIDER_GITHUB_REDIRECT_URL="http://localhost:8080/callback/github"
# GitHub is not an OIDC provider, so we might need to set Auth/Token URLs if not using a discovery proxy, 
# but for now the service expects OIDC discovery or manual config.
# If using manual config (supported in code):
export OAUTH_PROVIDER_GITHUB_AUTH_URL="https://github.com/login/oauth/authorize"
export OAUTH_PROVIDER_GITHUB_TOKEN_URL="https://github.com/login/oauth/access_token"
```

## Running the Service

1.  Navigate to the backend directory:
    ```bash
    cd backend
    ```

2.  Run the application:
    ```bash
    go run .
    ```

    The service will start on port `8080` (HTTP) and `8000` (gRPC).
    It will automatically create and migrate the SQLite database (`db.sqlite`).

## Testing

1.  Open your browser and navigate to:
    [http://localhost:8080/login](http://localhost:8080/login)

2.  You should see "Sign in with..." buttons for each configured provider.

3.  Click a button to start the OAuth flow. Upon success, you will see a page with your JWT.
