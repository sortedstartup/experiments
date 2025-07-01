package api

import (
	"context"
	"sortedstartup/load-test/dao"
	"time"

	"go.opentelemetry.io/otel"
)

type Service struct {
	DAO dao.DAO
}

func NewService(dao dao.DAO) *Service {
	return &Service{DAO: dao}
}

func (s *Service) Test(ctx context.Context, chatId string, message string) error {
	ctx, span := otel.Tracer("go_manual").Start(ctx, "service layer")
	defer span.End()
	time.Sleep(500 * time.Millisecond)
	return s.DAO.SaveMessage(ctx, chatId, message)
}
