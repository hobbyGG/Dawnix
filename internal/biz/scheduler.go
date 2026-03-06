package biz

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hobbyGG/Dawnix/internal/biz/model"
)

// 核心流程引擎
type Scheduler struct {
	txManager      TransactionManager
	definitionRepo ProcessDefinitionRepo
	instanceRepo   InstanceRepo
	executionRepo  ExecutionRepo
	taskCmdRepo    TaskCommandRepo
	navigator      *Navigator
	nodeHandlers   map[string]NodeHandlerFunc // nodeType -> 到达该节点时的处理逻辑
}

type SchedulerDependencies struct {
	TxManager      TransactionManager
	DefinitionRepo ProcessDefinitionRepo
	InstanceRepo   InstanceRepo
	ExecutionRepo  ExecutionRepo
	TaskCmdRepo    TaskCommandRepo
	Navigator      *Navigator
}

func NewScheduler(dependencies *SchedulerDependencies) *Scheduler {
	nh := &nodeHandlers{
		taskCmdRepo:   dependencies.TaskCmdRepo,
		executionRepo: dependencies.ExecutionRepo,
		instanceRepo:  dependencies.InstanceRepo,
	}
	return &Scheduler{
		txManager:      dependencies.TxManager,
		definitionRepo: dependencies.DefinitionRepo,
		instanceRepo:   dependencies.InstanceRepo,
		executionRepo:  dependencies.ExecutionRepo,
		taskCmdRepo:    dependencies.TaskCmdRepo,
		navigator:      dependencies.Navigator,
		nodeHandlers: map[string]NodeHandlerFunc{
			model.NodeTypeUserTask: nh.userTask,
			model.NodeTypeEnd:      nh.endNode,
		},
	}
}

// 接收service创建实例的意图
func (s *Scheduler) StartProcessInstance(ctx context.Context, cmd *StartProcessInstanceCmd) (int64, error) {
	// 由流程引擎负责整个创建实例的流程
	var instID int64
	err := s.txManager.InTx(ctx, func(ctx context.Context) error {
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

		variables, err := json.Marshal(cmd.Variables)
		if err != nil {
			return fmt.Errorf("variables marshal failed, %w", err)
		}

		// 创建流程实例
		inst := &model.ProcessInstance{
			DefinitionID:      processDef.ID,
			ProcessCode:       processDef.Code,
			SnapshotStructure: processDef.Structure,
			ParentID:          cmd.ParentID,
			ParentNodeID:      cmd.ParentNodeID,
			Variables:         variables,
			Status:            model.InstanceStatusPending,
			SubmitterID:       cmd.SubmitterID,
		}
		instID, err = s.instanceRepo.Create(ctx, inst)
		if err != nil {
			return fmt.Errorf("creatr instance failed, %w", err)
		}

		// 创建 Execution 记录
		// NOTE: 暂时不考虑自动执行节点的情况
		execParams := &model.Execution{
			InstID: inst.ID,
			NodeID: runtimeGraph.Next[runtimeGraph.StartNode.ID][0].TargetNode, // 从开始节点的下一跳开始执行
		}
		if err := s.executionRepo.Create(ctx, execParams); err != nil {
			return fmt.Errorf("failed to create execution record: %w", err)
		}

		// 根据节点类型分派处理逻辑
		firstNode := runtimeGraph.Nodes[execParams.NodeID]
		handler, ok := s.nodeHandlers[firstNode.Type]
		if !ok {
			return fmt.Errorf("unknown node type: %s", firstNode.Type)
		}
		if err := handler(ctx, firstNode, execParams); err != nil {
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
func (s *Scheduler) CompleteTask(ctx context.Context, task *model.ProcessTask) error {
	err := s.txManager.InTx(ctx, func(ctx context.Context) error {
		inst, err := s.instanceRepo.GetByID(ctx, task.InstanceID)
		if err != nil {
			return fmt.Errorf("failed to get instance by id: %w", err)
		}
		// 2. 解析流程，构建运行时图
		var graph model.GraphModel
		if err := json.Unmarshal(inst.SnapshotStructure, &graph); err != nil {
			return fmt.Errorf("structure unmarshal failed, %w", err)
		}
		runtimeGraph := NewSchedulerRuntimeGraph(&graph)
		// 3. 驱动流程流转
		if err := s.moveToken(ctx, task.ExecutionID, runtimeGraph); err != nil {
			return fmt.Errorf("failed to move token: %w", err)
		}
		// 4. 更新任务状态为已完成
		task.Status = model.TaskStatusApproved
		if err := s.taskCmdRepo.Update(ctx, task); err != nil {
			return fmt.Errorf("failed to update task status: %w", err)
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// TODO: 如果有多条边，先根据条件算出下一跳，再流转
func (s *Scheduler) moveToken(ctx context.Context, executionID int64, rg *RuntimeGraph) error {
	return s.txManager.InTx(ctx, func(ctx context.Context) error {
		exec, err := s.executionRepo.GetByID(ctx, executionID)
		if err != nil {
			return fmt.Errorf("failed to get execution record: %w", err)
		}

		// 找到所有可流转的边
		edges := rg.Next[exec.NodeID]
		if len(edges) == 0 {
			return fmt.Errorf("no outgoing edges from node %s", exec.NodeID)
		}
		if len(edges) > 1 {
			return fmt.Errorf("multiple outgoing edges not supported yet")
		}

		// 推进 token
		exec.NodeID = edges[0].TargetNode
		if err := s.executionRepo.Update(ctx, exec); err != nil {
			return fmt.Errorf("failed to update execution record: %w", err)
		}

		// 根据节点类型分派处理逻辑
		node := rg.Nodes[exec.NodeID]
		handler, ok := s.nodeHandlers[node.Type]
		if !ok {
			return fmt.Errorf("unknown node type: %s", node.Type)
		}
		return handler(ctx, node, exec)
	})
}
