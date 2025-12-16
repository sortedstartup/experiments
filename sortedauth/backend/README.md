# SortedAuth - Simple HTTP Auth Service

A minimal HTTP authentication service based on OAuth2 with JWT tokens, built in Go.

## Features

- **OAuth2 Authentication**: Supports Google OAuth (with fake OAuth for development)
- **JWT Token Management**: Issues and validates JWT tokens
- **SQLite/PostgreSQL Support**: Configurable database backend
- **HTTP Middleware**: Easy-to-use authentication middleware
- **Protected Endpoints**: Example of how to protect routes
- **Health Check**: Built-in health monitoring

## Quick Start

### 1. Install Dependencies

```bash
cd backend
go mod tidy
```

### 2. Configuration

Copy the example configuration:
```bash
cp config.example .env
```

Edit `.env` file with your settings (optional - defaults work for development).

### 3. Run the Service

```bash
# Default port 8080
go run main.go

# Custom port
go run main.go --http-port=8081

# Custom host and port
go run main.go --host=0.0.0.0 --http-port=8081
```

## Available Endpoints

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/health` | Health check | No |
| GET | `/login` | Login page with OAuth button | No |
| GET | `/callback` | OAuth callback handler | No |
| GET | `/oauth-config` | OAuth configuration for frontend | No |
| GET | `/protected` | Example protected endpoint | Yes |

## Testing the Service

### Health Check
```bash
curl http://localhost:8081/health
# Response: OK
```

### Login Page
```bash
curl http://localhost:8081/login
# Returns HTML login page with Google OAuth button
```

### OAuth Configuration
```bash
curl http://localhost:8081/oauth-config
# Returns JSON with OAuth configuration
```

### Protected Endpoint (without auth)
```bash
curl http://localhost:8081/protected
# Response: 401 Unauthorized - Authentication required
```

### Protected Endpoint (with JWT)
```bash
# First get a JWT token through the OAuth flow, then:
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" http://localhost:8081/protected
# Response: Hello user@example.com! Your user ID is: uuid-here
```

## Configuration Options

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_TYPE` | `sqlite` | Database type (`sqlite` or `postgres`) |
| `SQLITE_URL` | `auth.db` | SQLite database file path |
| `APP_JWT_SECRET` | `fake_jwt_secret_for_dev_only` | JWT signing secret |
| `APP_ISSUER` | `sortedchat-dev` | JWT issuer |
| `OAUTH_ISSUER_URL` | `http://localhost:8080/fakeoauth` | OAuth issuer URL |
| `GOOGLE_CLIENT_ID` | `fake_client_id` | OAuth client ID |
| `GOOGLE_CLIENT_SECRET` | `fake_client_secret` | OAuth client secret |
| `GOOGLE_REDIRECT_URL` | `http://localhost:8080/callback` | OAuth redirect URL |
| `ALLOWED_LOGIN_EMAILS` | (empty - allows all) | Comma-separated list of allowed emails |

### PostgreSQL Configuration

To use PostgreSQL instead of SQLite:

```bash
export DB_TYPE=postgres
export POSTGRES_HOST=localhost
export POSTGRES_PORT=5432
export POSTGRES_DATABASE=sortedauth
export POSTGRES_USERNAME=postgres
export POSTGRES_PASSWORD=yourpassword
export POSTGRES_SSL_MODE=disable
```

## Development vs Production

### Development Mode
- Uses fake OAuth provider for testing
- Default JWT secret (change in production!)
- SQLite database for simplicity

### Production Mode
- Set `APP_JWT_SECRET` to a strong random value
- Configure real Google OAuth credentials
- Use PostgreSQL for better performance
- Set `ALLOWED_LOGIN_EMAILS` to restrict access

## Architecture

The service is built with a clean architecture:

- **main.go**: HTTP server setup and routing
- **authservice/**: OAuth and user management
- **common/auth/**: JWT validation and HTTP middleware
- **authservice/dao/**: Database access layer

## Security Features

- JWT token validation with configurable secret and issuer
- Email allowlist for access control
- Secure OAuth2 flow with state parameter
- HTTP middleware for easy route protection
- Configurable authentication requirements per endpoint

## Adding Protected Routes

```go
// In your main.go or router setup
mux.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
    // Get user from context (set by auth middleware)
    userClaims, ok := auth.GetUserFromContext(r.Context())
    if !ok {
        http.Error(w, "Authentication required", http.StatusUnauthorized)
        return
    }
    
    // Your protected logic here
    fmt.Fprintf(w, "Hello %s!", userClaims.Email)
})
```

The auth middleware automatically validates JWT tokens and adds user information to the request context for protected routes.
