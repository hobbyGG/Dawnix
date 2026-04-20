package biz

import (
	"context"

	"github.com/hobbyGG/Dawnix/internal/workflow/domain"
)

type InstanceRepo interface {
	// 这里定义Instance相关的数据操作方法
	Create(ctx context.Context, model *domain.ProcessInstance) (int64, error)
	List(ctx context.Context, params *ListInstancesParams) ([]domain.ProcessInstance, error)
	GetByID(ctx context.Context, id int64) (*domain.ProcessInstance, error)
	GetWithExecutionsByID(ctx context.Context, id int64) (*domain.ProcessInstance, []domain.Execution, error)
	Delete(ctx context.Context, id int64) error
	Update(ctx context.Context, model *domain.ProcessInstance) error
	UpdateStatus(ctx context.Context, id int64, status string) error
}

type ListInstancesParams struct {
	Page int
	Size int
}

type InstanceScheduler interface {
	StartProcessInstance(ctx context.Context, params *StartProcessInstanceParams) (int64, error)
}

type StartProcessInstanceParams struct {
	// 定义了调用StartProcessInstance必须提供的所有信息
	ProcessCode  string         // 流程业务code
	SubmitterID  string         // 流程发起人
	FormData     []FormDataItem // 表单数据
	ParentID     int64          // 父流程id
	ParentNodeID string         // 父流程节点id
}
