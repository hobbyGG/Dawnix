package biz

import (
	"context"

	"github.com/hobbyGG/Dawnix/internal/biz/model"
)

type NodeContext struct {
	ctx context.Context
	// 其他必须的字段
	inst *model.ProcessInstance
}

type NodeBehaviour interface {
	OnEnter(ctx *NodeContext, node *model.NodeModel) (bool, error)
}
