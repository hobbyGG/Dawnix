package biz

import (
	"context"

	"github.com/hobbyGG/Dawnix/internal/biz/model"
	"gorm.io/datatypes"
)

type TaskCommandRepo interface {
	Create(ctx context.Context, task *model.ProcessTask) error
	GetByID(ctx context.Context, taskID int64) (*model.ProcessTask, error)
	Update(ctx context.Context, task *model.ProcessTask) error
}

// ==========================================
// 接口 2: TaskQueryRepo
// 给谁用: HTTP API (前端列表)
// 特点: 出参是 Read Model (TaskSummary)
// ==========================================
type TaskQueryRepo interface {
	GetDetailView(ctx context.Context, taskID int64) (*model.TaskView, error)
	ListPending(ctx context.Context, params *ListTasksParams) ([]*model.TaskView, error)
}

type ListTasksParams struct {
	Page int
	Size int
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
