package biz

import (
	"context"
	"encoding/json"

	"github.com/hobbyGG/Dawnix/internal/domain"
)

type FormDefinitionItem struct {
	Key   string          `json:"key"`
	Value json.RawMessage `json:"value"`
	Type  string          `json:"type"`
}

type ProcessDefinitionCreateParams struct {
	Name           string
	Code           string
	Structure      *domain.GraphModel
	FormDefinition []FormDefinitionItem
}

type ProcessDefinitionListParams struct {
	Page int
	Size int
}

type ProcessDefinitionRepo interface {
	// 这里定义ProcessDefinition相关的数据操作方法
	Create(ctx context.Context, processDefinition *domain.ProcessDefinition) (int64, error)
	List(ctx context.Context, params *ProcessDefinitionListParams) ([]domain.ProcessDefinition, error)
	GetByID(ctx context.Context, id int64) (*domain.ProcessDefinition, error)
	GetByCode(ctx context.Context, code string) (*domain.ProcessDefinition, error)
	DeleteByID(ctx context.Context, id int64) error
	Update(ctx context.Context, processDefinition *domain.ProcessDefinition) error
}
