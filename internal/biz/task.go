package biz

import (
	"context"

	"github.com/hobbyGG/Dawnix/internal/domain"
	"gorm.io/datatypes"
)

type TaskRepo interface {
	Create(ctx context.Context, task *domain.ProcessTask) error
	GetByID(ctx context.Context, taskID int64) (*domain.ProcessTask, error)
	Update(ctx context.Context, task *domain.ProcessTask) error
	GetDetailView(ctx context.Context, taskID int64) (*domain.TaskView, error)
	ListWithFilter(ctx context.Context, params *ListTasksParams) ([]*domain.TaskView, int64, error)
}

type TaskScheduler interface {
	CompleteTask(ctx context.Context, task *domain.ProcessTask) error
}

type ListTasksParams struct {
	Page      int
	Size      int
	Scope     string
	Status    string
	UserID    string // 当前用户ID (中间件注入)
	Submitter string // 提交人ID
}

type CreateTaskParams struct {
	// 归属的流程实例
	InstanceID int64

	// 对应的流程节点 ID (如 "UserTask_01")
	NodeID string

	// 任务类型 (通常是 "user_task"，但也可能是 "cc_task" 抄送)
	TaskType string

	// 指派给谁 (核心！Scheduler 解析规则后算出来的)
	// 格式: "user:1001" 或 "role:admin"
	Assignee string

	// 支持或签
	Candidates []string

	// 截止时间 (可选，Scheduler 根据配置算出)
	ExpireAt *int64

	// 上下文变量快照 (可选)
	Variables datatypes.JSON
}

type CompleteTaskParams struct {
	TaskID  int64
	Action  string
	UserID  int64
	Comment string
}
