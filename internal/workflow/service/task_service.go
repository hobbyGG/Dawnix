package service

import (
	"context"
	"encoding/json"
	"fmt"

	authService "github.com/hobbyGG/Dawnix/internal/auth/service"
	"github.com/hobbyGG/Dawnix/internal/workflow/biz"
	"github.com/hobbyGG/Dawnix/internal/workflow/domain"
	"go.uber.org/zap"
)

type TaskService struct {
	repo         biz.TaskRepo
	scheduler    biz.TaskScheduler
	logger       *zap.Logger
	defaultScope string
}

func NewTaskService(repo biz.TaskRepo, scheduler biz.TaskScheduler, logger *zap.Logger, defaultScope string) *TaskService {
	if defaultScope == "" {
		defaultScope = "my_todo"
	}
	return &TaskService{repo: repo, scheduler: scheduler, logger: logger, defaultScope: defaultScope}
}

func (s *TaskService) GetTaskDetailView(ctx context.Context, taskID int64) (*domain.TaskView, error) {
	detailView, err := s.repo.GetDetailView(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task detail: %w", err)
	}
	return detailView, nil
}

func (s *TaskService) ListTasksView(ctx context.Context, params *biz.ListTasksParams) ([]*domain.TaskView, int64, error) {
	if params == nil {
		return nil, 0, fmt.Errorf("list tasks params is nil")
	}
	if params.UserID == "" {
		uid, ok := authService.UserIDFromContext(ctx)
		if !ok {
			return nil, 0, fmt.Errorf("user id is required")
		}
		params.UserID = uid
	}
	if params.Scope == "" {
		params.Scope = s.defaultScope
	}
	// 根据不同scope做处理
	switch params.Scope {
	case "my_todo", "my_pending", "all_pending":
		params.Status = domain.TaskStatusPending
	case "my_completed", "all_completed":
		params.Status = domain.TaskStatusApproved
	default:
		return nil, 0, fmt.Errorf("unsupported task scope: %s", params.Scope)
	}
	taskViews, total, err := s.repo.ListWithFilter(ctx, params)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list tasks: %w", err)
	}
	return taskViews, total, nil
}

func (s *TaskService) CompleteTask(ctx context.Context, params *biz.CompleteTaskParams) error {
	if params == nil {
		return fmt.Errorf("complete task params is nil")
	}
	if params.UserID == "" {
		uid, ok := authService.UserIDFromContext(ctx)
		if !ok {
			return fmt.Errorf("user id is required")
		}
		params.UserID = uid
	}
	// 根据id拿到任务实例
	task, err := s.repo.GetByID(ctx, params.TaskID)
	if err != nil {
		return fmt.Errorf("failed to get task detail: %w", err)
	}
	// 2. [业务校验] 只有 PENDING 状态的任务才能完成
	if task.Status != domain.TaskStatusPending {
		return fmt.Errorf("task %d is not pending (current: %s)", task.ID, task.Status)
	}

	action := params.Action
	if action == "" {
		return fmt.Errorf("action is required")
	}

	// 业务处理
	switch action {
	case "agree":
		task.Status = domain.TaskStatusApproved
	case "reject":
		task.Status = domain.TaskStatusRejected
	default:
		return fmt.Errorf("unsupported action: %s", params.Action)
	}
	task.Action = action
	task.Comment = params.Comment // 保存审批意见
	if params.FormData != nil {
		payload, err := json.Marshal(params.FormData)
		if err != nil {
			return fmt.Errorf("marshal form_data failed: %w", err)
		}
		task.FormData = payload
	}

	// 3. 通知调度器完成任务
	if err := s.scheduler.CompleteTask(ctx, task); err != nil {
		return fmt.Errorf("failed to complete task in scheduler: %w", err)
	}
	return nil
}
