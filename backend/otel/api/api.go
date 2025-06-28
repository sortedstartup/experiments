package api

import (
	"fmt"
	context "context"
	"database/sql"
	"sortedstartup/otel/dao"
	"sortedstartup/otel/otel/proto"
)

// Server implements the gRPC service defined in your proto.
type Server struct {
	proto.UnimplementedSortedtestServer
	DAO dao.DAO
}

// NewServer returns a new Server instance.
func NewServer(db *sql.DB) *Server {
	return &Server{DAO: dao.NewDAO(db)}
}

type TestRequest = proto.TestRequest
type TestResponse = proto.TestResponse

func (s *Server) Test(ctx context.Context, req *TestRequest) (*TestResponse, error) {
	fmt.Println("Received Test request:", "req")
	err := s.DAO.SaveMessage(req.ChatId, req.Message)
	if err != nil {
		return nil, err
	}
	return &TestResponse{Text: "Saved: " + req.Message}, nil
}
