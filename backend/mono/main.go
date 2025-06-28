package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"sortedstartup/otel/dao"
	"sortedstartup/otel/mono/utils"
	"sortedstartup/otel/otel/api"
	"sortedstartup/otel/otel/proto"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/joho/godotenv"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	otellog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.34.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

const (
	grpcPort = ":8000"
	httpPort = ":8080" // Only gRPC-Web and static UI, no extra upload server
)

// var staticUIFS embed.FS

func main() {
	ctx := context.Background()

	res, err := newResource()
	if err != nil {
		log.Fatalf("Failed to create OTel resource: %v", err)
	}
	loggerProvider, err := newLoggerProvider(ctx, res)
	if err != nil {
		log.Fatalf("Failed to create OTel logger provider: %v", err)
	}
	defer func() {
		if err := loggerProvider.Shutdown(ctx); err != nil {
			fmt.Println("OTel logger shutdown error:", err)
		}
	}()
	global.SetLoggerProvider(loggerProvider)

	tp, err := initTracer()
	if err != nil {
		log.Fatal(err)
	}
	defer tp.Shutdown(context.Background())

	meterProvider, err := newMeterProvider(res)
	if err != nil {
		panic(err)
	}

	// Handle shutdown properly so nothing leaks.
	defer func() {
		if err := meterProvider.Shutdown(context.Background()); err != nil {
			log.Println(err)
		}
	}()

	otel.SetMeterProvider(meterProvider)

	err = godotenv.Load()
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
	apiServer := api.NewServer(db, loggerProvider)
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
	processor := otellog.NewBatchProcessor(exporter)
	provider := otellog.NewLoggerProvider(
		otellog.WithResource(res),
		otellog.WithProcessor(processor),
	)
	return provider, nil
}

func newMeterProvider(res *resource.Resource) (*metric.MeterProvider, error) {
	metricExporter, err := stdoutmetric.New()
	if err != nil {
		return nil, err
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(metric.NewPeriodicReader(metricExporter,
			// Default is 1m. Set to 3s for demonstrative purposes.
			metric.WithInterval(3*time.Second))),
	)
	return meterProvider, nil
}

func initTracer() (*sdktrace.TracerProvider, error) {
	// Create stdout exporter to be able to retrieve
	// the collected spans.
	exporter, err := otlptrace.New(context.Background(), otlptracehttp.NewClient())
	if err != nil {
		return nil, err
	}

	// For the demonstration, use sdktrace.AlwaysSample sampler to sample all traces.
	// In a production application, use sdktrace.ProbabilitySampler with a desired probability.
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return tp, err
}
