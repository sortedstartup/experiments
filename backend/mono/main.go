package main

import (
	"embed"
	"log"
	"net"
	"net/http"

	"sortedstartup/otel/mono/utils"
	"sortedstartup/otel/otel/api"
	"sortedstartup/otel/otel/proto"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	grpcPort = ":8000"
	httpPort = ":8080" // Only gRPC-Web and static UI, no extra upload server
)

//go:embed public
var staticUIFS embed.FS

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using system env")
	}

	listener, err := net.Listen("tcp", grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	apiServer := api.NewServer()                          // fix: no arguments
	proto.RegisterSortedtestServer(grpcServer, apiServer) // interface fix handled in api.go
	reflection.Register(grpcServer)

	wrappedGrpc := grpcweb.WrapServer(grpcServer)

	// Comment out staticUIFS if public folder does not exist
	// publicFS, err := fs.Sub(staticUIFS, "public")
	// if err != nil {
	// 	log.Fatalf("Failed to load static UI: %v", err)
	// }
	// staticUI := http.FileServer(http.FS(publicFS))

	httpMux := http.NewServeMux()
	httpMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if wrappedGrpc.IsGrpcWebRequest(r) || wrappedGrpc.IsAcceptableGrpcCorsRequest(r) {
			utils.EnableCORS(wrappedGrpc).ServeHTTP(w, r)
			return
		}
		// staticUI.ServeHTTP(w, r)
	})

	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: utils.EnableCORS(httpMux),
	}

	go func() {
		log.Println("Starting gRPC server on", grpcPort)
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("gRPC server error: %v", err)
		}
	}()

	log.Println("Starting gRPC-Web/Frontend server on :8080")
	if err := httpServer.ListenAndServe(); err != nil {
		log.Fatalf("HTTP server error: %v", err)
	}
}
