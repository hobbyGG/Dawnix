package biz

import (
	"context"

	"github.com/hobbyGG/Dawnix/internal/domain"
)

type NodeExecutor interface {
	Execute(ctx context.Context, node *domain.NodeModel) error
}
