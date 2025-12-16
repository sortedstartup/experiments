package main

import (
	"flag"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	authApi "sortedstartup/authservice/api"
	authDao "sortedstartup/authservice/dao"
	authService "sortedstartup/authservice/service"
	auth "sortedstartup/common/auth"
)

const (
	defaultHttpPort = "8080"
	defaultHost     = ""
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	// Parse command line flags
	host := flag.String("host", defaultHost, "Host to bind the server to (default: all interfaces)")
	httpPort := flag.String("http-port", defaultHttpPort, "Port for HTTP server")
	flag.Parse()

	// Build address
	httpAddr := net.JoinHostPort(*host, *httpPort)

	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using system env")
	}

	// Get JWT configuration
	jwtSecret := os.Getenv("APP_JWT_SECRET")
	issuer := os.Getenv("APP_ISSUER")

	// Use defaults if not set
	defaultJwtSecret, defaultIssuer := getJWTDefaults()
	if jwtSecret == "" {
		jwtSecret = defaultJwtSecret
	}
	if issuer == "" {
		issuer = defaultIssuer
	}

	// Create JWT validator
	validator := auth.NewJWTValidator([]byte(jwtSecret), issuer)

	// Create HTTP auth middleware
	authMiddleware := auth.NewHTTPAuthMiddleware(validator, false) // requireAuth = false for flexibility

	// Skip authentication for certain paths
	authMiddleware.SkipPaths([]string{
		"/health",
		"/login",
		"/callback",
		"/oauth-config",
	})

	// Create HTTP mux
	mux := http.NewServeMux()

	authConfig, err := authDao.LoadConfig()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		log.Fatalf("Failed to load configuration: %v", err)
	}

	authDaoFactory, err := authDao.NewDAOFactory(authConfig)
	if err != nil {
		slog.Error("Failed to create user service DAO", "error", err)
		log.Fatalf("Failed to create user service DAO: %v", err)
	}

	userServiceDao, err := authDaoFactory.CreateDAO()
	if err != nil {
		slog.Error("Failed to create user service DAO", "error", err)
		log.Fatalf("Failed to create user service DAO: %v", err)
	}

	userService := authService.NewUserService(userServiceDao)
	userService.Init(authConfig)
	authService := authService.NewAuthService(userService)
	authServiceApi := authApi.NewAuthServiceAPI(mux, authService)
	authServiceApi.Init()

	// Add health endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Wrap handler with auth middleware
	wrappedHandler := authMiddleware.Middleware(mux)

	// Create HTTP server
	httpServer := &http.Server{
		Addr:    httpAddr,
		Handler: wrappedHandler,
	}

	slog.Info("HTTP server starting", "address", httpAddr)

	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("HTTP server error", "error", err, "address", httpAddr)
		log.Fatalf("HTTP server error: %v", err)
	}
}
