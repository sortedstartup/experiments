package main

import (
	"database/sql"
	// "embed"
	"log"
	"net"
	"net/http"
	"os"
	"fmt"

	"sortedstartup/otel/dao"
	"sortedstartup/otel/otel/api"
	"sortedstartup/otel/otel/proto"
	"sortedstartup/otel/mono/utils"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	grpcPort = ":8000"
	httpPort = ":8080" // Only gRPC-Web and static UI, no extra upload server
)


// var staticUIFS embed.FS

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using system env")
	}

	dbPath := os.Getenv("SQLITE_DB_PATH")
	fmt.Println("Using DB path:", dbPath)
	if dbPath == "" {
		dbPath = "./app.db"
	}
	if err := dao.MigrateSQLite(dbPath); err != nil {
		log.Fatalf("DB migration failed: %v", err)
	}
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to open DB: %v", err)
	}
	defer db.Close()

	listener, err := net.Listen("tcp", grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	apiServer := api.NewServer(db)
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
