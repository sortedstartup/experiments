package api

import (
	"context"
	"database/sql"
	"sortedstartup/load-test/dao"
	"sortedstartup/load-test/test/proto"
)

// Server implements the gRPC service defined in your proto.
type Server struct {
	proto.UnimplementedSortedtestServer
	DAO     dao.DAO
	Service *Service
}

// NewServer returns a new Server instance.
func NewServer(db *sql.DB) *Server {
	myDao := dao.NewDAO(db)
	service := NewService(myDao)
	return &Server{DAO: myDao, Service: service}
}

type TestRequest = proto.TestRequest
type TestResponse = proto.TestResponse

func (s *Server) Test(ctx context.Context, req *TestRequest) (*TestResponse, error) {
	err := s.Service.Test(ctx, req.ChatId, req.Message)
	if err != nil {
		return nil, err
	}
	return &TestResponse{Text: "Saved: " + req.Message}, nil
}
