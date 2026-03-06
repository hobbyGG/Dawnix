package biz

import (
	"context"

	"github.com/hobbyGG/Dawnix/internal/biz/model"
)

type ExecutionRepo interface {
	Create(ctx context.Context, exec *model.Execution) error
	GetByID(ctx context.Context, id int64) (*model.Execution, error)
	Update(ctx context.Context, exec *model.Execution) error
	DeleteByID(ctx context.Context, id int64) error
}
