package service

import (
	"context"

	"github.com/hobbyGG/Dawnix/internal/biz"
	"go.uber.org/zap"
)

type HelloService struct {
	biz.HelloRepo
	logger *zap.Logger
}

func NewHelloService(repo biz.HelloRepo, logger *zap.Logger) *HelloService {
	return &HelloService{
		HelloRepo: repo,
		logger:    logger,
	}
}

func (s *HelloService) Hello(ctx context.Context, data interface{}) error {
	// 假装执行
	s.logger.Info("service hello...")
	return s.HelloRepo.Hello(ctx, data)
}
