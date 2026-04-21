package biz

import (
	"context"
	"fmt"

	"github.com/hobbyGG/Dawnix/internal/workflow/domain"
)

type Node interface {
	ID() string
	Type() string
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

type NodeFeatures struct {
	EmailServiceEnabled bool
}

func NewDefaultNodeRegistry(deps NodeDeps, features NodeFeatures) NodeRegistry {
	registry := NodeRegistry{
		domain.NodeTypeStart: func(node *domain.NodeModel) (Node, error) {
			return newStartNode(node)
		},
		domain.NodeTypeUserTask: func(node *domain.NodeModel) (Node, error) {
			return newUserTaskNode(node, deps.TaskRepo)
		},
		domain.NodeTypeEnd: func(node *domain.NodeModel) (Node, error) {
			return newEndNode(node, deps.ExecutionRepo, deps.InstanceRepo)
		},
		domain.NodeTypeForkGateway: func(node *domain.NodeModel) (Node, error) {
			return newForkGatewayNode(node)
		},
		domain.NodeTypeJoinGateway: func(node *domain.NodeModel) (Node, error) {
			return newJoinGatewayNode(node)
		},
		domain.NodeTypeXORGateway: func(node *domain.NodeModel) (Node, error) {
			return newXorGatewayNode(node)
		},
		domain.NodeTypeInclusiveGateway: func(node *domain.NodeModel) (Node, error) {
			return newInclusiveGatewayNode(node)
		},
	}
	if features.EmailServiceEnabled {
		registry[domain.NodeTypeEmailService] = func(node *domain.NodeModel) (Node, error) {
			return newEmailServiceNode(node, deps.ServiceTaskMQ)
		}
	}
	return registry
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
