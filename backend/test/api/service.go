package api

import (
	"context"
	"sortedstartup/load-test/dao"
	"time"
)

type Service struct {
	DAO dao.DAO
}

func NewService(dao dao.DAO) *Service {
	return &Service{DAO: dao}
}

func (s *Service) Test(ctx context.Context, chatId string, message string) error {
	time.Sleep(500 * time.Millisecond)
	return s.DAO.SaveMessage(ctx, chatId, message)
}
