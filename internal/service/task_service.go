package service

import (
	"context"
	"fmt"
	"time"

	"github.com/hobbyGG/Dawnix/internal/biz"
	"github.com/hobbyGG/Dawnix/internal/biz/model"
	"go.uber.org/zap"
)

type TaskService struct {
	cmd       biz.TaskCommandRepo
	query     biz.TaskQueryRepo
	scheduler biz.TaskScheduler
	logger    *zap.Logger
}

func NewTaskService(cmd biz.TaskCommandRepo, query biz.TaskQueryRepo, scheduler biz.TaskScheduler, logger *zap.Logger) *TaskService {
	return &TaskService{cmd: cmd, query: query, scheduler: scheduler, logger: logger}
}

func (s *TaskService) GetTaskDetailView(ctx context.Context, taskID int64) (*model.TaskView, error) {
	detailView, err := s.query.GetDetailView(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task detail: %w", err)
	}
	return detailView, nil
}

func (s *TaskService) ListTasksView(ctx context.Context, params *biz.ListTasksParams) ([]*model.TaskView, int64, error) {
	// 根据不同scope做处理
	switch params.Scope {
	case "my_todo":
		params.Status = model.TaskStatusPending
	case "my_completed":
		params.Status = model.TaskStatusApproved
	default:
		return nil, 0, fmt.Errorf("unsupported task scope: %s", params.Scope)
	}
	taskViews, total, err := s.query.ListWithFilter(ctx, params)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list tasks: %w", err)
	}
	return taskViews, total, nil
}

func (s *TaskService) CompleteTask(ctx context.Context, params *biz.CompleteTaskParams) error {
	// 根据id拿到任务实例
	task, err := s.cmd.GetByID(ctx, params.TaskID)
	if err != nil {
		return fmt.Errorf("failed to get task detail: %w", err)
	}
	// 2. [业务校验] 只有 PENDING 状态的任务才能完成
	if task.Status != model.TaskStatusPending {
		return fmt.Errorf("task %d is not pending (current: %s)", task.ID, task.Status)
	}

	// 业务处理
	task.Status = model.TaskStatusApproved
	task.Comment = params.Comment // 保存审批意见
	now := time.Now()
	task.FinishedAt = &now

	// 3. 通知调度器完成任务
	if err := s.scheduler.CompleteTask(ctx, task); err != nil {
		return fmt.Errorf("failed to complete task in scheduler: %w", err)
	}
	if err := s.cmd.Update(ctx, task); err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}
	return nil
}
