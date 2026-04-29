package service

import (
	"context"
	//"fmt"
	"github.com/hobbyGG/Dawnix/internal/workflow/biz"
	"github.com/hobbyGG/Dawnix/internal/workflow/domain"
	"go.uber.org/zap"
)

type RecordService struct {
	RecordRepo biz.RecordRepo
	logger     *zap.Logger
}

func NewRecordService(RecordRepo biz.RecordRepo, logger *zap.Logger) *RecordService {
	return &RecordService{
		RecordRepo: RecordRepo,
		logger:     logger,
	}
}

func (s *RecordService) ListByInstanceID(ctx context.Context, instanceID int64) ([]*domain.Record, error) {
	return s.RecordRepo.List(ctx, instanceID)
}

func (s *RecordService) ListAll(ctx context.Context) ([]*domain.Record, error) {
	return s.RecordRepo.ListAll(ctx)
}
