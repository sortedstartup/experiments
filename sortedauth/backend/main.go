package main

import (
	"flag"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	authApi "sortedstartup/authservice/api"
	authDao "sortedstartup/authservice/dao"
	proto "sortedstartup/authservice/proto"
	authService "sortedstartup/authservice/service"
	auth "sortedstartup/common/auth"
	util "sortedstartup/utils"
)

const (
	defaultGrpcPort = "8000"
	defaultHttpPort = "8080"
	defaultHost     = ""
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	// Parse command line flags
	host := flag.String("host", defaultHost, "Host to bind the server to (default: all interfaces)")
	httpPort := flag.String("http-port", defaultHttpPort, "Port for HTTP server")
	grpcPort := flag.String("grpc-port", defaultGrpcPort, "Port for gRPC server")
	flag.Parse()

	// Build addresses
	grpcAddr := net.JoinHostPort(*host, *grpcPort)
	httpAddr := net.JoinHostPort(*host, *httpPort)

	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using system env")
	}

	// Create gRPC listener
	grpcListener, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", grpcAddr, err)
	}

	// Get JWT configuration
	jwtSecret := os.Getenv("APP_JWT_SECRET")
	issuer := os.Getenv("APP_ISSUER")
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
	authMiddleware := auth.NewHTTPAuthMiddleware(validator, false)
	authMiddleware.SkipPaths([]string{
		"/health",
		"/login",
		"/callback",
		"/oauth-config",
		"/google-one-tap-callback",
	})

	// Load auth config and create services
	authConfig, err := authDao.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	authDaoFactory, err := authDao.NewDAOFactory(authConfig)
	if err != nil {
		log.Fatalf("Failed to create user service DAO: %v", err)
	}

	userServiceDao, err := authDaoFactory.CreateDAO()
	if err != nil {
		log.Fatalf("Failed to create user service DAO: %v", err)
	}

	userService := authService.NewUserService(userServiceDao)
	userService.Init(authConfig)
	userServiceAPI := authApi.NewUserServiceAPI(userService)
	authSvc := authService.NewAuthService(userService)

	// Create HTTP mux and register HTTP routes
	mux := http.NewServeMux()
	authServiceApi := authApi.NewAuthServiceAPI(mux, authSvc)
	authServiceApi.Init()

	// Add health endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Create gRPC server and register services
	grpcServer := grpc.NewServer()
	proto.RegisterUserServiceServer(grpcServer, userServiceAPI)
	reflection.Register(grpcServer)

	// Wrap gRPC server for gRPC-Web support
	wrappedGrpc := grpcweb.WrapServer(grpcServer)

	// Combined handler: gRPC-Web + HTTP
	combinedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if wrappedGrpc.IsGrpcWebRequest(r) || wrappedGrpc.IsAcceptableGrpcCorsRequest(r) {
			wrappedGrpc.ServeHTTP(w, r)
			return
		}
		// Regular HTTP requests go to the mux
		mux.ServeHTTP(w, r)
	})

	// Wrap with auth middleware
	wrappedHandler := authMiddleware.Middleware(util.EnableCORS(combinedHandler))

	// Channel to catch server errors
	serverErr := make(chan error, 2)

	// Start gRPC server in goroutine
	go func() {
		slog.Info("gRPC server starting", "address", grpcAddr)
		serverErr <- grpcServer.Serve(grpcListener)
	}()

	// Start HTTP server (with gRPC-Web support) in goroutine
	go func() {
		slog.Info("HTTP server starting", "address", httpAddr)
		httpServer := &http.Server{
			Addr:    httpAddr,
			Handler: wrappedHandler,
		}
		serverErr <- httpServer.ListenAndServe()
	}()

	// Wait for any server error
	if err := <-serverErr; err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
