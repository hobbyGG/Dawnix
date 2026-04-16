package biz

import (
	"context"
	"fmt"

	"github.com/hobbyGG/Dawnix/internal/domain"
)

type Node interface {
	ID() string
	Type() string
	AutoAdvance() bool
	Handle(ctx context.Context, exec *domain.Execution, rg *RuntimeGraph) (*domain.ProcessTask, error)
}

type NodeBuilder func(node *domain.NodeModel) (Node, error)

type NodeRegistry map[string]NodeBuilder

type NodeDeps struct {
	TaskRepo      TaskRepo
	InstanceRepo  InstanceRepo
	ExecutionRepo ExecutionRepo
	ServiceTaskMQ ServiceTaskMQ
}

func NewDefaultNodeRegistry(deps NodeDeps) NodeRegistry {
	return NodeRegistry{
		domain.NodeTypeStart: func(node *domain.NodeModel) (Node, error) {
			return newStartNode(node)
		},
		domain.NodeTypeUserTask: func(node *domain.NodeModel) (Node, error) {
			return newUserTaskNode(node, deps.TaskRepo)
		},
		domain.NodeTypeEnd: func(node *domain.NodeModel) (Node, error) {
			return newEndNode(node, deps.ExecutionRepo, deps.InstanceRepo)
		},
		domain.NodeTypeEmailService: func(node *domain.NodeModel) (Node, error) {
			return newEmailServiceNode(node, deps.ServiceTaskMQ)
		},
		domain.NodeTypeForkGateway: func(node *domain.NodeModel) (Node, error) {
			return newForkGatewayNode(node, deps.TaskRepo)
		},
		domain.NodeTypeJoinGateway: func(node *domain.NodeModel) (Node, error) {
			return newJoinGatewayNode(node, deps.TaskRepo)
		},
		domain.NodeTypeXORGateway: func(node *domain.NodeModel) (Node, error) {
			return newXorGatewayNode(node, deps.TaskRepo)
		},
	}
}

func buildNodeFromRegistry(registry NodeRegistry, nodeModel *domain.NodeModel) (Node, error) {
	if registry == nil {
		return nil, fmt.Errorf("node registry is nil")
	}
	if nodeModel == nil {
		return nil, fmt.Errorf("node model is nil")
	}
	builder, ok := registry[nodeModel.Type]
	if !ok {
		return nil, fmt.Errorf("unsupported node type: %s", nodeModel.Type)
	}
	node, err := builder(nodeModel)
	if err != nil {
		return nil, fmt.Errorf("build node %s failed: %w", nodeModel.ID, err)
	}
	return node, nil
}
