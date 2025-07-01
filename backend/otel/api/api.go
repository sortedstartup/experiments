package api

import (
	"context"
	"database/sql"
	"log/slog"
	"sortedstartup/otel/dao"
	"sortedstartup/otel/otel/proto"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	otellog "go.opentelemetry.io/otel/sdk/log"
)

var meter = otel.Meter("sortedstartup/otel/otel/api")
var apiCounter metric.Int64Counter

// Server implements the gRPC service defined in your proto.
type Server struct {
	proto.UnimplementedSortedtestServer
	DAO     dao.DAO
	Log     *slog.Logger
	Service *Service
}

// NewServer returns a new Server instance.
func NewServer(db *sql.DB, loggerProvider *otellog.LoggerProvider) *Server {
	log := otelslog.NewLogger("my/pkg/name", otelslog.WithLoggerProvider(loggerProvider)) //slog-otelbridge
	myDao := dao.NewDAO(db)
	service := NewService(myDao)
	return &Server{DAO: myDao, Log: log, Service: service}
}

func init() {
	var err error
	apiCounter, err = meter.Int64Counter(
		"api.counter",
		metric.WithDescription("Number of API calls."),
		metric.WithUnit("{call}"),
	)
	if err != nil {
		panic(err)
	}
}

type TestRequest = proto.TestRequest
type TestResponse = proto.TestResponse

func (s *Server) Test(ctx context.Context, req *TestRequest) (*TestResponse, error) {
	//Logging
	s.Log.Info("In api")

	//metrics
	apiCounter.Add(ctx, 1)

	//Tracing
	ctx, span := otel.Tracer("go_manual").Start(ctx, "api layer")
	defer span.End()

	// time.Sleep(500 * time.Millisecond)
	err := s.Service.Test(ctx, req.ChatId, req.Message)
	if err != nil {
		return nil, err
	}
	return &TestResponse{Text: "Saved: " + req.Message}, nil
}
