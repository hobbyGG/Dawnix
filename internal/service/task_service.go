package service

import (
	"context"

	"github.com/hobbyGG/Dawnix/internal/biz"
	"github.com/hobbyGG/Dawnix/internal/biz/model"
	"go.uber.org/zap"
)

type TaskService struct {
	cmd    biz.TaskCommandRepo
	query  biz.TaskQueryRepo
	logger *zap.Logger
}

func NewTaskService(cmd biz.TaskCommandRepo, query biz.TaskQueryRepo, logger *zap.Logger) *TaskService {
	return &TaskService{cmd: cmd, query: query, logger: logger}
}

func (s *TaskService) GetTaskDetail(ctx context.Context, taskID int64) (*model.TaskView, error) {
	detailView, err := s.query.GetDetailView(ctx, taskID)
	if err != nil {
		s.logger.Error("failed to get task detail view", zap.Int64("taskID", taskID), zap.Error(err))
		return nil, err
	}
	return detailView, nil
}
