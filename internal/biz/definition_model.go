package biz

import (
	"context"

	"github.com/hobbyGG/Dawnix/internal/biz/model"
)

type ProcessDefinitionCreateParams struct {
	Name      string
	Code      string
	Structure WorkflowGraph
}

type ProcessDefinitionListParams struct {
	Page int
	Size int
}

type ProcessDefinitionRepo interface {
	// 这里定义ProcessDefinition相关的数据操作方法
	Create(ctx context.Context, processDefinition *model.ProcessDefinition) (int64, error)
	List(ctx context.Context, params *ProcessDefinitionListParams) ([]*model.ProcessDefinition, error)
	GetByID(ctx context.Context, id int64) (*model.ProcessDefinition, error)
	GetByCode(ctx context.Context, code string) (*model.ProcessDefinition, error)
	DeleteByID(ctx context.Context, id int64) error
	Update(ctx context.Context, processDefinition *model.ProcessDefinition) error
}
