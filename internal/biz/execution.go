package biz

import (
	"context"

	"github.com/hobbyGG/Dawnix/internal/biz/model"
)

type ExecutionRepo interface {
	Create(ctx context.Context, exec *model.Execution) error
	CreateBatch(ctx context.Context, execs []model.Execution) error
	GetByID(ctx context.Context, id int64) (*model.Execution, error)
	GetActiveNums(ctx context.Context, instID int64) (int, error)
	GetActiveNumsByParentID(ctx context.Context, parentID int64) (int, error)
	Update(ctx context.Context, exec *model.Execution) error
	UpdateNodeID(ctx context.Context, execID int64, nodeID string) error
	DeleteByID(ctx context.Context, id int64) error
}
