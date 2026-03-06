package biz

import (
	"context"

	"github.com/hobbyGG/Dawnix/internal/biz/model"
)

type NodeExecutor interface {
	Execute(ctx context.Context, node *model.NodeModel) error
}
