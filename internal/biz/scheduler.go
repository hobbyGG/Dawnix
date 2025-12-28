package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hobbyGG/Dawnix/internal/biz/model"
)

type Scheduler struct {
	txManager      TransactionManager
	definitionRepo ProcessDefinitionRepo
	instanceRepo   InstanceRepo
	taskCmdRepo    TaskCommandRepo
	navigator      *Navigator
	nodeExecutor   map[string]NodeBehaviour
}

type SchedulerDependencies struct {
	TxManager      TransactionManager
	DefinitionRepo ProcessDefinitionRepo
	InstanceRepo   InstanceRepo
	TaskCmdRepo    TaskCommandRepo
	Navigator      *Navigator
	NodeExecutor   map[string]NodeBehaviour
}

func NewScheduler(dependencies *SchedulerDependencies) *Scheduler {
	return &Scheduler{
		txManager:      dependencies.TxManager,
		definitionRepo: dependencies.DefinitionRepo,
		instanceRepo:   dependencies.InstanceRepo,
		taskCmdRepo:    dependencies.TaskCmdRepo,
		navigator:      dependencies.Navigator,
		nodeExecutor:   dependencies.NodeExecutor,
	}
}

// 接收service创建实例的意图
func (s *Scheduler) StartProcessInstance(ctx context.Context, cmd *StartProcessInstanceCmd) (int64, error) {
	// 由流程引擎负责整个创建实例的流程
	// 根据cmd的信息创建流程实例
	var instID int64
	err := s.txManager.InTx(ctx, func(ctx context.Context) error {
		// 开启事务操作
		// 根据code找到对应流程模板
		processDef, err := s.definitionRepo.GetByCode(ctx, cmd.ProcessCode)
		if err != nil {
			return fmt.Errorf("Get definition failed, %w", err)
		}
		// 从def中拿到SchedulerGraph
		var graph model.GraphModel
		if err := json.Unmarshal(processDef.Structure, &graph); err != nil {
			return fmt.Errorf("structure unmarshal failed, %w", err)
		}
		runtimeGraph := NewSchedulerRuntimeGraph(&graph)
		// 创建流程实例
		variables, err := json.Marshal(cmd.Variables)
		if err != nil {
			return fmt.Errorf("variables marshal failed, %w", err)
		}
		inst := &model.ProcessInstance{
			DefinitionID:      processDef.ID,
			ProcessCode:       processDef.Code,
			SnapshotStructure: processDef.Structure,
			ParentID:          cmd.ParentID,
			ParentNodeID:      cmd.ParentNodeID,
			Variables:         variables,
			ActiveTokens:      []string{runtimeGraph.StartNode.ID}, // 初始令牌放在开始节点
			Status:            model.InstanceStatusPending,
			SubmitterID:       cmd.SubmitterID,
		}
		instID, err = s.instanceRepo.Create(ctx, inst)
		if err != nil {
			return fmt.Errorf("creatr instance failed, %w", err)
		}
		// 初始化后消费startToken，触发后续节点流转
		if err := s.moveToken(ctx, inst, runtimeGraph, runtimeGraph.StartNode.ID); err != nil {
			return fmt.Errorf("failed to move token: %w", err)
		}
		return nil
	})
	if err != nil {
		return 0, err
	}

	return instID, nil
}

// 接受service完成任务的意图
func (s *Scheduler) CompleteTask(ctx context.Context, cmd *CompleteTaskCmd) error {
	err := s.txManager.InTx(ctx, func(ctx context.Context) error {
		inst, err := s.instanceRepo.GetByID(ctx, cmd.Task.InstanceID)
		if err != nil {
			return fmt.Errorf("failed to get instance by id: %w", err)
		}
		// 2. 解析流程定义，构建运行时图
		var graph model.GraphModel
		if err := json.Unmarshal(inst.SnapshotStructure, &graph); err != nil {
			return fmt.Errorf("structure unmarshal failed, %w", err)
		}
		runtimeGraph := NewSchedulerRuntimeGraph(&graph)
		// 3. 驱动流程流转
		if err := s.moveToken(ctx, inst, runtimeGraph, cmd.Task.NodeID); err != nil {
			return fmt.Errorf("failed to move token: %w", err)
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *Scheduler) moveToken(ctx context.Context, inst *model.ProcessInstance, rg *RuntimeGraph, nodeID string) error {
	// 1. 消耗当前节点的 Token (原子操作，内存中移除)
	inst.ConsumeToken(nodeID)

	// 2. 寻找所有出路
	initialPaths := s.navigator.FindPaths(rg, nodeID)
	tokenQueue := tokenQueue{}
	for _, edge := range initialPaths {
		tokenQueue.Enqueue(edge.TargetNode)
	}

	burstVisited := make(map[string]bool)
	// 3. BFS 推进 Token
	for !tokenQueue.IsEmpty() {
		currentTargetID, _ := tokenQueue.Dequeue()

		targetNode, exists := rg.Nodes[currentTargetID]
		if !exists {
			return fmt.Errorf("target node %s not found", currentTargetID)
		}

		if targetNode.Type == model.NodeTypeEnd {
			// 遇到结束节点：不执行任何逻辑，不产生新 Token
			// 直接跳过，让 Token 消失 (Consumed)
			continue
		}

		executor, exists := s.nodeExecutor[targetNode.Type]
		if !exists {
			return fmt.Errorf("unknown executor for type %s", targetNode.Type)
		}

		// 自动节点防死循环检测
		if targetNode.IsAutoType() {
			if burstVisited[currentTargetID] {
				return fmt.Errorf("infinite loop detected at %s", currentTargetID)
			}
			burstVisited[currentTargetID] = true
		}

		// 执行节点逻辑
		nodeContext := &NodeContext{ctx: ctx, inst: inst}
		shouldContinue, err := executor.OnEnter(nodeContext, targetNode)
		if err != nil {
			return err
		}

		if shouldContinue {
			// 情况一：自动节点 (如抄送、或者你定义的 RuleTask) -> 继续往下找
			nextPaths := s.navigator.FindPaths(rg, currentTargetID)
			for _, edge := range nextPaths {
				tokenQueue.Enqueue(edge.TargetNode)
			}
		} else {
			// 情况二：审批节点 (UserTask) -> 停下来，生成 Token
			inst.ProduceToken(currentTargetID)
		}
	}

	// 4. 最终状态判定
	// 只有当所有 Token 都被消耗掉（走到 End 节点被吞掉），且没有新 Token 生成时，流程才算结束
	if len(inst.ActiveTokens) == 0 {
		inst.Status = model.InstanceStatusApproved
		now := time.Now()
		inst.FinishedAt = &now
	}

	// 5. 持久化
	return s.instanceRepo.Update(ctx, inst)
}
