package biz

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hobbyGG/Dawnix/internal/domain"
)

// 核心流程引擎
type Scheduler struct {
	txManager      TransactionManager
	definitionRepo ProcessDefinitionRepo
	instanceRepo   InstanceRepo
	executionRepo  ExecutionRepo
	taskRepo       TaskRepo
	nodeRegistry   NodeRegistry
}

type completeTaskHandler func(ctx context.Context, exec *domain.Execution, rg *RuntimeGraph) error

func NewScheduler(
	TxManager TransactionManager,
	DefinitionRepo ProcessDefinitionRepo,
	InstanceRepo InstanceRepo,
	ExecutionRepo ExecutionRepo,
	TaskRepo TaskRepo,
	NodeRegistry NodeRegistry,
) *Scheduler {
	if TxManager == nil || DefinitionRepo == nil || InstanceRepo == nil || ExecutionRepo == nil || TaskRepo == nil {
		panic("missing dependencies for Scheduler")
	}
	if NodeRegistry == nil {
		panic("missing node registry for Scheduler")
	}
	return &Scheduler{
		txManager:      TxManager,
		definitionRepo: DefinitionRepo,
		instanceRepo:   InstanceRepo,
		executionRepo:  ExecutionRepo,
		taskRepo:       TaskRepo,
		nodeRegistry:   NodeRegistry,
	}
}

// 接收service创建实例的意图
func (s *Scheduler) StartProcessInstance(ctx context.Context, params *StartProcessInstanceParams) (int64, error) {
	// 由流程引擎负责整个创建实例的流程
	var instID int64
	err := s.txManager.InTx(ctx, func(ctx context.Context) error {
		// 根据code找到对应流程模板
		processDef, err := s.definitionRepo.GetByCode(ctx, params.ProcessCode)
		if err != nil {
			return fmt.Errorf("Get definition failed, %w", err)
		}

		// 从def中拿到SchedulerGraph
		var graph domain.GraphModel
		if err := json.Unmarshal(processDef.Structure, &graph); err != nil {
			return fmt.Errorf("structure unmarshal failed, %w", err)
		}
		runtimeGraph, err := NewSchedulerRuntimeGraph(&graph, s.nodeRegistry)
		if err != nil {
			return fmt.Errorf("build runtime graph failed: %w", err)
		}
		startEdges := runtimeGraph.Next[runtimeGraph.StartNode.ID()]
		if len(startEdges) == 0 {
			return fmt.Errorf("start node %s has no outgoing edge", runtimeGraph.StartNode.ID())
		}

		variables, err := json.Marshal(params.Variables)
		if err != nil {
			return fmt.Errorf("variables marshal failed, %w", err)
		}

		// 创建流程实例
		inst := &domain.ProcessInstance{
			DefinitionID:      processDef.ID,
			ProcessCode:       processDef.Code,
			SnapshotStructure: processDef.Structure,
			ParentID:          params.ParentID,
			ParentNodeID:      params.ParentNodeID,
			Variables:         variables,
			Status:            domain.InstanceStatusPending,
			SubmitterID:       params.SubmitterID,
		}
		instID, err = s.instanceRepo.Create(ctx, inst)
		if err != nil {
			return fmt.Errorf("creatr instance failed, %w", err)
		}

		// 创建 Execution 记录
		// NOTE: 暂时不考虑自动执行节点的情况
		exec := &domain.Execution{
			InstID: inst.ID,
			NodeID: startEdges[0].TargetNode, // 从开始节点的下一跳开始执行
		}
		if err := s.executionRepo.Create(ctx, exec); err != nil {
			return fmt.Errorf("failed to create execution record: %w", err)
		}

		if _, _, err := s.moveToken(ctx, exec.ID, exec.NodeID, runtimeGraph); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return 0, err
	}

	return instID, nil
}

// 接受service完成任务的意图
// 必须是审批类任务完成后才会调用这个接口，其他类型的任务不经过审批，直接由流程引擎自己推进
func (s *Scheduler) CompleteTask(ctx context.Context, task *domain.ProcessTask) error {
	err := s.txManager.InTx(ctx, func(ctx context.Context) error {
		task.Status = domain.TaskStatusApproved
		if err := s.taskRepo.Update(ctx, task); err != nil {
			return fmt.Errorf("failed to update task status: %w", err)
		}

		inst, err := s.instanceRepo.GetByID(ctx, task.InstanceID)
		if err != nil {
			return fmt.Errorf("failed to get instance by id: %w", err)
		}
		// 解析流程，构建运行时图
		var graph domain.GraphModel
		if err := json.Unmarshal(inst.SnapshotStructure, &graph); err != nil {
			return fmt.Errorf("structure unmarshal failed, %w", err)
		}

		rg, err := NewSchedulerRuntimeGraph(&graph, s.nodeRegistry)
		if err != nil {
			return fmt.Errorf("build runtime graph failed: %w", err)
		}
		exec, err := s.executionRepo.GetByID(ctx, task.ExecutionID)
		if err != nil {
			return fmt.Errorf("failed to get execution by id: %w", err)
		}

		currentNode, ok := rg.Nodes[exec.NodeID]
		if !ok {
			return fmt.Errorf("node not found: %s", exec.NodeID)
		}

		handler := s.resolveCompleteTaskHandler(currentNode.Type())
		if err := handler(ctx, exec, rg); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *Scheduler) resolveCompleteTaskHandler(nodeType string) completeTaskHandler {
	handlers := map[string]completeTaskHandler{
		domain.NodeTypeForkGateway: s.handleForkGatewayTaskCompletion,
		domain.NodeTypeJoinGateway: s.handleJoinGatewayTaskCompletion,
	}
	if handler, ok := handlers[nodeType]; ok {
		return handler
	}
	return s.handleDefaultTaskCompletion
}

func (s *Scheduler) handleForkGatewayTaskCompletion(ctx context.Context, exec *domain.Execution, rg *RuntimeGraph) error {
	exec.IsActive = false
	if err := s.executionRepo.Update(ctx, exec); err != nil {
		return fmt.Errorf("failed to update execution: %w", err)
	}

	edges := rg.Next[exec.NodeID]
	execs := make([]domain.Execution, 0, len(edges))
	for _, edge := range edges {
		execs = append(execs, domain.Execution{
			InstID:   exec.InstID,
			ParentID: exec.ID,
			NodeID:   edge.TargetNode,
			IsActive: true,
		})
	}
	if err := s.executionRepo.CreateBatch(ctx, execs); err != nil {
		return fmt.Errorf("failed to create branch executions: %w", err)
	}
	for i := range execs {
		if _, _, err := s.moveToken(ctx, execs[i].ID, execs[i].NodeID, rg); err != nil {
			return err
		}
	}
	return nil
}

func (s *Scheduler) handleJoinGatewayTaskCompletion(ctx context.Context, exec *domain.Execution, rg *RuntimeGraph) error {
	parentID := exec.ParentID
	nodeID := exec.NodeID

	if err := s.executionRepo.DeleteByID(ctx, exec.ID); err != nil {
		return fmt.Errorf("failed to delete join execution: %w", err)
	}

	count, err := s.executionRepo.GetActiveNumsByParentID(ctx, parentID)
	if err != nil {
		return fmt.Errorf("failed to count active executions by parent id: %w", err)
	}
	if count > 0 {
		return nil
	}

	if parentID == 0 {
		return fmt.Errorf("join gateway has no parent execution, nodeID: %s", nodeID)
	}

	nextEdges := rg.Next[nodeID]
	if len(nextEdges) == 0 {
		return fmt.Errorf("node %s has no outgoing edge", nodeID)
	}
	if _, _, err := s.moveToken(ctx, parentID, nextEdges[0].TargetNode, rg); err != nil {
		return err
	}
	return nil
}

func (s *Scheduler) handleDefaultTaskCompletion(ctx context.Context, exec *domain.Execution, rg *RuntimeGraph) error {
	nextEdges := rg.Next[exec.NodeID]
	if len(nextEdges) == 0 {
		return fmt.Errorf("node %s has no outgoing edge", exec.NodeID)
	}
	if _, _, err := s.moveToken(ctx, exec.ID, nextEdges[0].TargetNode, rg); err != nil {
		return err
	}
	return nil
}

// moveToken 根据执行流id跳转到指定节点，并调用节点处理器
func (s *Scheduler) moveToken(ctx context.Context, executionID int64, nextNodeID string, rg *RuntimeGraph) (*domain.Execution, *domain.ProcessTask, error) {
	exec, err := s.executionRepo.GetByID(ctx, executionID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get execution by id: %w", err)
	}
	node, ok := rg.Nodes[nextNodeID]
	if !ok {
		return nil, nil, fmt.Errorf("node not found: %s", nextNodeID)
	}
	exec.NodeID = nextNodeID
	exec.IsActive = true
	if err := s.executionRepo.Update(ctx, exec); err != nil {
		return nil, nil, fmt.Errorf("failed to update execution: %w", err)
	}

	task, err := node.Handle(ctx, exec, rg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to handle node %s: %w", nextNodeID, err)
	}
	if task == nil && node.AutoAdvance() {
		nextEdges := rg.Next[nextNodeID]
		if len(nextEdges) == 0 {
			return nil, nil, fmt.Errorf("node %s has no outgoing edge", nextNodeID)
		}
		return s.moveToken(ctx, exec.ID, nextEdges[0].TargetNode, rg)
	}

	return exec, task, nil
}
