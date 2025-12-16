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

	"sortedstartup/common/auth"
	paymentApi "sortedstartup/sortedpay/paymentservice/api"
	paymentDao "sortedstartup/sortedpay/paymentservice/dao"
	paymentProto "sortedstartup/sortedpay/paymentservice/proto"
)

const (
	defaultGrpcPort = "8000"
	defaultHttpPort = "8080"
	defaultHost     = ""
)

func main() {
	// Set up logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	// Parse command line flags
	host := flag.String("host", defaultHost, "Host to bind the server to (default: all interfaces)")
	grpcPort := flag.String("grpc-port", defaultGrpcPort, "Port for gRPC server")
	httpPort := flag.String("http-port", defaultHttpPort, "Port for HTTP server")
	flag.Parse()

	// Build addresses
	grpcAddr := net.JoinHostPort(*host, *grpcPort)
	httpAddr := net.JoinHostPort(*host, *httpPort)

	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system env")
	}

	// Create listener for gRPC
	listener, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", grpcAddr, err)
	}

	// Set up JWT validator and auth interceptors
	jwtSecret := os.Getenv("APP_JWT_SECRET")
	issuer := os.Getenv("APP_ISSUER")

	// Default values if not set
	if jwtSecret == "" {
		jwtSecret = "your-secret-key" // Change this in production
	}
	if issuer == "" {
		issuer = "sortedstartup"
	}

	validator := auth.NewJWTValidator([]byte(jwtSecret), issuer)

	// Create gRPC auth interceptor
	authInterceptor := auth.NewGRPCAuthInterceptor(validator, true) // requireAuth = true

	// Skip authentication for health check
	authInterceptor.SkipMethods([]string{
		"/grpc.health.v1.Health/Check",
	})

	// Create gRPC server with interceptors
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptor.UnaryInterceptor()),
		grpc.StreamInterceptor(authInterceptor.StreamInterceptor()),
	)

	// Load payment service configuration
	config, err := paymentDao.LoadConfig()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		log.Fatalf("Failed to load configuration: %v", err)
	}

	slog.Info("Payment service configuration loaded",
		"database_type", config.Database.Type,
		"postgres_host", config.Database.Postgres.Host,
		"postgres_port", config.Database.Postgres.Port,
		"sqlite_url", config.Database.SQLite.URL)

	// Create DAO factory
	daoFactory, err := paymentDao.NewDAOFactory(config)
	if err != nil {
		slog.Error("Failed to create DAO factory", "error", err)
		log.Fatalf("Failed to create DAO factory: %v", err)
	}
	defer func() {
		if err := daoFactory.Close(); err != nil {
			log.Printf("Error closing DAO factory: %v", err)
		}
	}()

	// Create HTTP mux
	mux := http.NewServeMux()

	// Create payment service API
	paymentServiceApi := paymentApi.NewPaymentServiceAPI(mux, daoFactory)
	if paymentServiceApi == nil {
		slog.Error("Failed to create payment service API")
		log.Fatal("Failed to create payment service API")
	}

	// Initialize payment service (runs migrations and seeds)
	if err := paymentServiceApi.Init(config); err != nil {
		slog.Error("Failed to initialize payment service", "error", err)
		log.Fatalf("Failed to initialize payment service: %v", err)
	}

	// Register payment service with gRPC server
	paymentProto.RegisterPaymentServiceServer(grpcServer, paymentServiceApi)

	// Enable reflection (for testing/debugging)
	reflection.Register(grpcServer)

	// Add health endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// gRPC-Web wrapper
	wrappedGrpc := grpcweb.WrapServer(grpcServer)

	// HTTP handler that routes gRPC-Web requests to gRPC server, otherwise uses mux
	httpHandler := func(w http.ResponseWriter, r *http.Request) {
		if wrappedGrpc.IsGrpcWebRequest(r) || wrappedGrpc.IsAcceptableGrpcCorsRequest(r) {
			wrappedGrpc.ServeHTTP(w, r)
			return
		}
		// For non-gRPC requests, use the mux
		mux.ServeHTTP(w, r)
	}

	// Create HTTP auth middleware
	authMiddleware := auth.NewHTTPAuthMiddleware(validator, false) // requireAuth = false for webhooks

	// Skip authentication for webhooks and health check
	authMiddleware.SkipPaths([]string{
		"/health",
		"/stripe-webhook",
		"/razorpay-webhook",
	})

	// HTTP server with auth middleware
	httpServer := &http.Server{
		Addr:    httpAddr,
		Handler: authMiddleware.Middleware(http.HandlerFunc(httpHandler)),
	}

	// Run both servers in parallel
	serverErr := make(chan error, 2)

	go func() {
		log.Printf("Starting gRPC server on %s", grpcAddr)
		serverErr <- grpcServer.Serve(listener)
	}()

	go func() {
		log.Printf("Starting HTTP server on %s", httpAddr)
		serverErr <- httpServer.ListenAndServe()
	}()

	// Wait for server error
	err = <-serverErr
	if err != nil {
		slog.Error("Server error", "error", err)
		log.Fatalf("Server error: %v", err)
	}
}
