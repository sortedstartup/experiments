package api

import (
	context "context"
	"sortedstartup/otel/otel/proto"
)

// Server implements the gRPC service defined in your proto.
type Server struct {
	proto.UnimplementedSortedtestServer
}

// NewServer returns a new Server instance.
func NewServer() *Server {
	return &Server{}
}

// Implement the Test method for SortedtestServer interface
type TestRequest = proto.TestRequest
type TestResponse = proto.TestResponse

func (s *Server) Test(ctx context.Context, req *TestRequest) (*TestResponse, error) {
	// Example implementation
	return &TestResponse{Text: "Hello, " + req.Message}, nil
}
