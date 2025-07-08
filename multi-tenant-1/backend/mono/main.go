package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"net/http"
	"os"

	"sortedstartup/multi-tenant/dao"
	"sortedstartup/multi-tenant/mono/utils"
	"sortedstartup/multi-tenant/test/api"
	"sortedstartup/multi-tenant/test/proto"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/joho/godotenv"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	otellog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.34.0"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	grpcPort = ":8000"
	httpPort = ":8080"
)

func main() {
	ctx := context.Background()

	// Initialize OpenTelemetry resource
	res, err := newResource()
	if err != nil {
		log.Fatalf("Failed to create OTel resource: %v", err)
	}

	// Logger provider
	loggerProvider, err := newLoggerProvider(ctx, res)
	if err != nil {
		log.Fatalf("Failed to create OTel logger provider: %v", err)
	}
	defer loggerProvider.Shutdown(ctx)
	global.SetLoggerProvider(loggerProvider)

	// Tracer provider
	tp, err := initTracer()
	if err != nil {
		log.Fatal(err)
	}
	defer tp.Shutdown(ctx)

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system env")
	}

	// Setup DB
	dbPath := os.Getenv("SQLITE_DB_PATH")
	if dbPath == "" {
		dbPath = "./app.db"
	}
	log.Println("Using DB path:", dbPath)

	if err := dao.MigrateSQLite(dbPath); err != nil {
		log.Fatalf("DB migration failed: %v", err)
	}

	superDB, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to open super DB: %v", err)
	}
	defer superDB.Close()

	if err := dao.InitTenantDBs(superDB, "../mono"); err != nil {
		log.Fatalf("Failed to initialize tenant DBs: %v", err)
	}

	// Set up gRPC server
	listener, err := net.Listen("tcp", grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	apiServer := api.NewServer(superDB, loggerProvider)
	proto.RegisterSortedtestServer(grpcServer, apiServer)
	reflection.Register(grpcServer)

	// Wrap with gRPC-Web
	wrappedGrpc := grpcweb.WrapServer(grpcServer)

	otelHandler := otelhttp.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if wrappedGrpc.IsGrpcWebRequest(r) || wrappedGrpc.IsAcceptableGrpcCorsRequest(r) {
			wrappedGrpc.ServeHTTP(w, r)
			return
		}
		http.NotFound(w, r)
	}), "grpc-web-gateway")

	finalHandler := utils.EnableCORS(otelHandler)

	httpServer := &http.Server{
		Addr:    httpPort,
		Handler: finalHandler,
	}

	// Run servers
	go func() {
		log.Println("Starting gRPC server on", grpcPort)
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("gRPC server error: %v", err)
		}
	}()

	log.Println("Starting HTTP server on", httpPort)
	if err := httpServer.ListenAndServe(); err != nil {
		log.Fatalf("HTTP server error: %v", err)
	}
}

// --- OTel Setup ---

func newResource() (*resource.Resource, error) {
	return resource.Merge(resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL,
			semconv.ServiceName("my-service"),
			semconv.ServiceVersion("0.1.0"),
		),
	)
}

func newLoggerProvider(ctx context.Context, res *resource.Resource) (*otellog.LoggerProvider, error) {
	exporter, err := otlploghttp.New(ctx)
	if err != nil {
		return nil, err
	}
	return otellog.NewLoggerProvider(
		otellog.WithResource(res),
		otellog.WithProcessor(otellog.NewBatchProcessor(exporter)),
	), nil
}

func initTracer() (*sdktrace.TracerProvider, error) {
	exporter, err := otlptrace.New(context.Background(), otlptracehttp.NewClient())
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}),
	)
	return tp, nil
}
