package biz

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Knetic/govaluate"
	"github.com/hobbyGG/Dawnix/internal/workflow/domain"
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

		definitionItems, err := DecodeFormDataItems(processDef.FormDefinition)
		if err != nil {
			return fmt.Errorf("decode form_definition failed: %w", err)
		}
		if err := ValidateRuntimeFormDataAgainstDefinition(definitionItems, params.FormData); err != nil {
			return fmt.Errorf("invalid form_data: %w", err)
		}

		formData, err := json.Marshal(params.FormData)
		if err != nil {
			return fmt.Errorf("form_data marshal failed: %w", err)
		}

		// 创建流程实例
		inst := &domain.ProcessInstance{
			DefinitionID:      processDef.ID,
			ProcessCode:       processDef.Code,
			SnapshotStructure: processDef.Structure,
			ParentID:          params.ParentID,
			ParentNodeID:      params.ParentNodeID,
			FormData:          formData,
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
		if task == nil {
			return fmt.Errorf("task is nil")
		}
		action := task.Action
		if action == "" {
			action = "agree"
			task.Action = action
		}

		if err := s.taskRepo.Update(ctx, task); err != nil {
			return fmt.Errorf("failed to update task status: %w", err)
		}

		if action == "reject" {
			if err := s.executionRepo.DeleteByID(ctx, task.ExecutionID); err != nil {
				return fmt.Errorf("failed to delete execution on reject: %w", err)
			}
			if err := s.instanceRepo.UpdateStatus(ctx, task.InstanceID, domain.InstanceStatusRejected); err != nil {
				return fmt.Errorf("failed to update instance status on reject: %w", err)
			}
			return nil
		}
		if action != "agree" {
			return fmt.Errorf("unsupported action in scheduler: %s", task.Action)
		}

		inst, err := s.instanceRepo.GetByID(ctx, task.InstanceID)
		if err != nil {
			return fmt.Errorf("failed to get instance by id: %w", err)
		}
		processDef, err := s.definitionRepo.GetByID(ctx, inst.DefinitionID)
		if err != nil {
			return fmt.Errorf("failed to get definition by id: %w", err)
		}
		definitionItems, err := DecodeFormDataItems(processDef.FormDefinition)
		if err != nil {
			return fmt.Errorf("decode form_definition failed: %w", err)
		}
		taskItems, err := DecodeFormDataItems(task.FormData)
		if err != nil {
			return fmt.Errorf("decode task form_data failed: %w", err)
		}
		if err := ValidateRuntimeFormDataAgainstDefinition(definitionItems, taskItems); err != nil {
			return fmt.Errorf("invalid task form_data: %w", err)
		}
		if err := s.mergeInstanceFormData(ctx, inst, task.FormData); err != nil {
			return err
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
	switch nodeType {
	case domain.NodeTypeForkGateway:
		return s.handleForkGatewayTaskCompletion
	case domain.NodeTypeJoinGateway:
		return s.handleJoinGatewayTaskCompletion
	case domain.NodeTypeXORGateway:
		return s.handleXORGatewayTaskCompletion
	case domain.NodeTypeInclusiveGateway:
		return s.handleInclusiveGatewayTaskCompletion
	default:
		return s.handleDefaultTaskCompletion
	}
}

func (s *Scheduler) handleXORGatewayTaskCompletion(ctx context.Context, exec *domain.Execution, rg *RuntimeGraph) error {
	nextEdges := rg.Next[exec.NodeID]
	if len(nextEdges) == 0 {
		return fmt.Errorf("xor node %s has no outgoing edge", exec.NodeID)
	}

	inst, err := s.instanceRepo.GetByID(ctx, exec.InstID)
	if err != nil {
		return fmt.Errorf("failed to get instance by id: %w", err)
	}

	targetEdge, err := selectXORTargetEdge(nextEdges, inst.FormData, exec.NodeID)
	if err != nil {
		return err
	}

	if _, _, err := s.moveToken(ctx, exec.ID, targetEdge.TargetNode, rg); err != nil {
		return err
	}
	return nil
}

func (s *Scheduler) handleForkGatewayTaskCompletion(ctx context.Context, exec *domain.Execution, rg *RuntimeGraph) error {
	edges := rg.Next[exec.NodeID]
	if len(edges) == 0 {
		return fmt.Errorf("node %s has no outgoing edge", exec.NodeID)
	}
	return s.fanOutExecutions(ctx, exec, edges, rg)
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

func (s *Scheduler) handleInclusiveGatewayTaskCompletion(ctx context.Context, exec *domain.Execution, rg *RuntimeGraph) error {
	nextEdges := rg.Next[exec.NodeID]
	if len(nextEdges) == 0 {
		return fmt.Errorf("inclusive node %s has no outgoing edge", exec.NodeID)
	}

	inst, err := s.instanceRepo.GetByID(ctx, exec.InstID)
	if err != nil {
		return fmt.Errorf("failed to get instance by id: %w", err)
	}

	formData, err := formDataMapFromPayload(inst.FormData, exec.NodeID, "inclusive")
	if err != nil {
		return err
	}

	var defaultEdge *domain.EdgeModel
	matched := make([]*domain.EdgeModel, 0, len(nextEdges))

	for _, edge := range nextEdges {
		if edge == nil {
			continue
		}

		if edge.IsDefault {
			if defaultEdge != nil {
				return fmt.Errorf("inclusive node %s has more than one default edge", exec.NodeID)
			}
			defaultEdge = edge
			continue
		}

		condition := edge.Condition
		if condition == "" {
			// no condition means always match for inclusive gateway
			matched = append(matched, edge)
			continue
		}

		hit, err := evaluateEdgeCondition(exec.NodeID, edge, condition, formData, "inclusive")
		if err != nil {
			return err
		}
		if hit {
			matched = append(matched, edge)
		}
	}

	// if no conditions matched, use default edge
	if len(matched) == 0 {
		if defaultEdge == nil {
			return fmt.Errorf("inclusive node %s has no matched edge and no default edge", exec.NodeID)
		}
		matched = append(matched, defaultEdge)
	}
	return s.fanOutExecutions(ctx, exec, matched, rg)
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

func selectXORTargetEdge(edges []*domain.EdgeModel, formPayload []byte, nodeID string) (*domain.EdgeModel, error) {
	formData, err := formDataMapFromPayload(formPayload, nodeID, "xor")
	if err != nil {
		return nil, err
	}

	var defaultEdge *domain.EdgeModel
	matched := make([]*domain.EdgeModel, 0, len(edges))
	for _, edge := range edges {
		if edge == nil {
			continue
		}

		if edge.IsDefault {
			if defaultEdge != nil {
				return nil, fmt.Errorf("xor node %s has more than one default edge", nodeID)
			}
			if edge.Condition != "" {
				return nil, fmt.Errorf("xor node %s edge %s: default edge cannot define condition", nodeID, edge.ID)
			}
			defaultEdge = edge
			continue
		}

		condition := edge.Condition
		if condition == "" {
			return nil, fmt.Errorf("xor node %s edge %s: condition is required", nodeID, edge.ID)
		}

		hit, err := evaluateEdgeCondition(nodeID, edge, condition, formData, "xor")
		if err != nil {
			return nil, err
		}
		if hit {
			matched = append(matched, edge)
		}
	}

	if len(matched) == 1 {
		return matched[0], nil
	}
	if len(matched) > 1 {
		return nil, fmt.Errorf("xor node %s matched %d edges, expected exactly one", nodeID, len(matched))
	}
	if defaultEdge != nil {
		return defaultEdge, nil
	}

	return nil, fmt.Errorf("xor node %s has no matched edge and no default edge", nodeID)
}

func (s *Scheduler) fanOutExecutions(ctx context.Context, exec *domain.Execution, edges []*domain.EdgeModel, rg *RuntimeGraph) error {
	exec.IsActive = false
	if err := s.executionRepo.Update(ctx, exec); err != nil {
		return fmt.Errorf("failed to update execution: %w", err)
	}

	nextExecutions := make([]domain.Execution, 0, len(edges))
	for _, edge := range edges {
		nextExecutions = append(nextExecutions, domain.Execution{
			InstID:   exec.InstID,
			ParentID: exec.ID,
			NodeID:   edge.TargetNode,
			IsActive: true,
		})
	}

	if err := s.executionRepo.CreateBatch(ctx, nextExecutions); err != nil {
		return fmt.Errorf("failed to create branch executions: %w", err)
	}
	for i := range nextExecutions {
		if _, _, err := s.moveToken(ctx, nextExecutions[i].ID, nextExecutions[i].NodeID, rg); err != nil {
			return err
		}
	}
	return nil
}

func formDataMapFromPayload(formPayload []byte, nodeID, gateway string) (map[string]interface{}, error) {
	items, err := DecodeFormDataItems(formPayload)
	if err != nil {
		return nil, fmt.Errorf("%s node %s decode instance form_data failed: %w", gateway, nodeID, err)
	}
	formData, err := FormDataItemsToMap(items)
	if err != nil {
		return nil, fmt.Errorf("%s node %s convert form_data failed: %w", gateway, nodeID, err)
	}
	return formData, nil
}

func evaluateEdgeCondition(nodeID string, edge *domain.EdgeModel, condition string, formData map[string]interface{}, gateway string) (bool, error) {
	expression, err := govaluate.NewEvaluableExpression(condition)
	if err != nil {
		return false, fmt.Errorf("%s node %s edge %s parse condition failed: %w", gateway, nodeID, edge.ID, err)
	}

	result, err := expression.Evaluate(formData)
	if err != nil {
		return false, fmt.Errorf("%s node %s edge %s evaluate condition failed: %w", gateway, nodeID, edge.ID, err)
	}

	hit, ok := result.(bool)
	if !ok {
		return false, fmt.Errorf("%s node %s edge %s condition result must be bool", gateway, nodeID, edge.ID)
	}
	return hit, nil
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

func (s *Scheduler) mergeInstanceFormData(ctx context.Context, inst *domain.ProcessInstance, taskFormData []byte) error {
	if inst == nil || len(taskFormData) == 0 {
		return nil
	}

	baseItems, err := DecodeFormDataItems(inst.FormData)
	if err != nil {
		return fmt.Errorf("decode instance form data failed: %w", err)
	}

	incomingItems, err := DecodeFormDataItems(taskFormData)
	if err != nil {
		return fmt.Errorf("decode task form data failed: %w", err)
	}

	mergedItems := MergeFormDataItems(baseItems, incomingItems)

	payload, err := json.Marshal(mergedItems)
	if err != nil {
		return fmt.Errorf("marshal merged form data failed: %w", err)
	}
	inst.FormData = payload
	if err := s.instanceRepo.Update(ctx, inst); err != nil {
		return fmt.Errorf("update instance form data failed: %w", err)
	}

	return nil
}
