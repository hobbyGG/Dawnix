package biz

import (
	"context"

	"github.com/hobbyGG/Dawnix/internal/biz/model"
)

type InstanceRepo interface {
	// 这里定义Instance相关的数据操作方法
	Create(ctx context.Context, model *model.ProcessInstance) (int64, error)
	List(ctx context.Context, params *ListInstancesParams) ([]model.ProcessInstance, error)
	GetByID(ctx context.Context, id int64) (*model.ProcessInstance, error)
	Delete(ctx context.Context, id int64) error
	Update(ctx context.Context, model *model.ProcessInstance) error
}

type ListInstancesParams struct {
	Page int
	Size int
}

type InstanceScheduler interface {
	StartProcessInstance(ctx context.Context, cmd *StartProcessInstanceCmd) (int64, error)
}

type StartProcessInstanceCmd struct {
	// 定义了调用StartProcessInstance必须提供的所有信息
	ProcessCode  string                 // 流程业务code
	SubmitterID  string                 // 流程发起人
	Variables    map[string]interface{} // 表单数据
	ParentID     int64                  // 父流程id
	ParentNodeID string                 // 父流程节点id
}

type CompleteTaskCmd struct {
	Task   *model.ProcessTask
	Action string
	UserID int64 // 执行用户ID
}
