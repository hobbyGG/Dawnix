package biz

import (
	"context"
	"fmt"

	"github.com/hobbyGG/Dawnix/internal/workflow/domain"
)

type nodeBase struct {
	id   string
	kind string
}

func (n nodeBase) ID() string {
	return n.id
}

func (n nodeBase) Type() string {
	return n.kind
}

func (n nodeBase) AutoAdvance() bool {
	return false
}

type startNode struct {
	nodeBase
}

func newStartNode(node *domain.NodeModel) (Node, error) {
	if node == nil {
		return nil, fmt.Errorf("node is nil")
	}
	return &startNode{nodeBase: nodeBase{id: node.ID, kind: domain.NodeTypeStart}}, nil
}

func (n *startNode) Handle(ctx context.Context, exec *domain.Execution, rg *RuntimeGraph) (*domain.ProcessTask, error) {
	return nil, nil
}

type taskNode struct {
	nodeBase
	taskRepo   TaskRepo
	taskType   string
	assignee   string
	candidates []string
}

func newTaskNode(node *domain.NodeModel, taskRepo TaskRepo, kind string, taskType string) (Node, error) {
	if node == nil {
		return nil, fmt.Errorf("node is nil")
	}
	if taskRepo == nil {
		return nil, fmt.Errorf("task repo is nil")
	}
	assignee, candidates := resolveTaskAssignment(node.Candidates.Users)
	return &taskNode{
		nodeBase:   nodeBase{id: node.ID, kind: kind},
		taskRepo:   taskRepo,
		taskType:   taskType,
		assignee:   assignee,
		candidates: candidates,
	}, nil
}

func (n *taskNode) Handle(ctx context.Context, exec *domain.Execution, rg *RuntimeGraph) (*domain.ProcessTask, error) {
	task := &domain.ProcessTask{
		InstanceID:  exec.InstID,
		ExecutionID: exec.ID,
		NodeID:      exec.NodeID,
		Type:        n.taskType,
		Assignee:    n.assignee,
		Candidates:  n.candidates,
		Status:      domain.TaskStatusPending,
	}
	if err := n.taskRepo.Create(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}
	return task, nil
}

func newUserTaskNode(node *domain.NodeModel, taskRepo TaskRepo) (Node, error) {
	return newTaskNode(node, taskRepo, domain.NodeTypeUserTask, domain.TaskTypeUser)
}

type endNode struct {
	nodeBase
	executionRepo ExecutionRepo
	instanceRepo  InstanceRepo
}

func newEndNode(node *domain.NodeModel, executionRepo ExecutionRepo, instanceRepo InstanceRepo) (Node, error) {
	if node == nil {
		return nil, fmt.Errorf("node is nil")
	}
	if executionRepo == nil {
		return nil, fmt.Errorf("execution repo is nil")
	}
	if instanceRepo == nil {
		return nil, fmt.Errorf("instance repo is nil")
	}
	return &endNode{nodeBase: nodeBase{id: node.ID, kind: domain.NodeTypeEnd}, executionRepo: executionRepo, instanceRepo: instanceRepo}, nil
}

func (n *endNode) Handle(ctx context.Context, exec *domain.Execution, rg *RuntimeGraph) (*domain.ProcessTask, error) {
	if err := n.executionRepo.DeleteByID(ctx, exec.ID); err != nil {
		return nil, fmt.Errorf("failed to delete execution: %w", err)
	}
	if err := n.instanceRepo.UpdateStatus(ctx, exec.InstID, domain.InstanceStatusApproved); err != nil {
		return nil, fmt.Errorf("failed to update instance status: %w", err)
	}
	return nil, nil
}

type emailServiceNode struct {
	nodeBase
	payload []byte
	mq      ServiceTaskMQ
}

func newEmailServiceNode(node *domain.NodeModel, mq ServiceTaskMQ) (Node, error) {
	if node == nil {
		return nil, fmt.Errorf("node is nil")
	}
	if mq == nil {
		return nil, fmt.Errorf("service task mq is nil")
	}
	return &emailServiceNode{nodeBase: nodeBase{id: node.ID, kind: domain.NodeTypeEmailService}, payload: node.Properties, mq: mq}, nil
}

func (n *emailServiceNode) AutoAdvance() bool {
	return true
}

func (n *emailServiceNode) Handle(ctx context.Context, exec *domain.Execution, rg *RuntimeGraph) (*domain.ProcessTask, error) {
	if err := n.mq.ProduceEmailTask(ctx, n.payload); err != nil {
		return nil, fmt.Errorf("failed to produce email task: %w", err)
	}
	return nil, nil
}

func newForkGatewayNode(node *domain.NodeModel, taskRepo TaskRepo) (Node, error) {
	return newTaskNode(node, taskRepo, domain.NodeTypeForkGateway, domain.TaskTypeReceive)
}

func newJoinGatewayNode(node *domain.NodeModel, taskRepo TaskRepo) (Node, error) {
	return newTaskNode(node, taskRepo, domain.NodeTypeJoinGateway, domain.TaskTypeReceive)
}

func newXorGatewayNode(node *domain.NodeModel, taskRepo TaskRepo) (Node, error) {
	return newTaskNode(node, taskRepo, domain.NodeTypeXORGateway, domain.TaskTypeReceive)
}

func newInclusiveGatewayNode(node *domain.NodeModel, taskRepo TaskRepo) (Node, error) {
	return newTaskNode(node, taskRepo, domain.NodeTypeInclusiveGateway, domain.TaskTypeReceive)
}

func resolveTaskAssignment(users []string) (string, []string) {
	if len(users) == 0 {
		return "", nil
	}
	if len(users) == 1 {
		return users[0], users
	}
	candidates := make([]string, len(users))
	copy(candidates, users)
	return "", candidates
}
