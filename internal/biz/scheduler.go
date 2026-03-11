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
	nodeHandlers   map[string]NodeHandlerFunc // nodeType -> 到达该节点时的处理逻辑
	srvTaskMQ      ServiceTaskMQ
}

func NewScheduler(
	TxManager TransactionManager,
	DefinitionRepo ProcessDefinitionRepo,
	InstanceRepo InstanceRepo,
	ExecutionRepo ExecutionRepo,
	TaskCmdRepo TaskCommandRepo,
	ServiceTaskMQ ServiceTaskMQ,
) *Scheduler {
	if TxManager == nil || DefinitionRepo == nil || InstanceRepo == nil || ExecutionRepo == nil || TaskCmdRepo == nil {
		panic("missing dependencies for Scheduler")
	}
	nh := &nodeHandlers{
		taskCmdRepo:   TaskCmdRepo,
		executionRepo: ExecutionRepo,
		instanceRepo:  InstanceRepo,
	}
	return &Scheduler{
		txManager:      TxManager,
		definitionRepo: DefinitionRepo,
		instanceRepo:   InstanceRepo,
		executionRepo:  ExecutionRepo,
		taskCmdRepo:    TaskCmdRepo,
		nodeHandlers: map[string]NodeHandlerFunc{
			model.NodeTypeUserTask: nh.userTask,
			model.NodeTypeEnd:      nh.endNode,
		},
		srvTaskMQ: ServiceTaskMQ,
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
		exec := &model.Execution{
			InstID: inst.ID,
			NodeID: runtimeGraph.Next[runtimeGraph.StartNode.ID][0].TargetNode, // 从开始节点的下一跳开始执行
		}
		if err := s.executionRepo.Create(ctx, exec); err != nil {
			return fmt.Errorf("failed to create execution record: %w", err)
		}

		return s.moveToken(ctx, exec.ID, exec.NodeID, runtimeGraph)
	})
	if err != nil {
		return 0, err
	}

	return instID, nil
}

// 接受service完成任务的意图
// 必须是审批类任务完成后才会调用这个接口，其他类型的任务不经过审批，直接由流程引擎自己推进
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

		rg := NewSchedulerRuntimeGraph(&graph)
		// 3. 驱动流程流转
		nextNodeID := rg.Next[task.NodeID][0].TargetNode
		if err := s.moveToken(ctx, task.ExecutionID, nextNodeID, rg); err != nil {
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

// moveToken 根据执行流id跳转到指定节点，移动的 execution 一定会变成 active 状态
func (s *Scheduler) moveToken(ctx context.Context, executionID int64, nextNodeID string, rg *RuntimeGraph) error {
	exec, err := s.executionRepo.GetByID(ctx, executionID)
	if err != nil {
		return fmt.Errorf("failed to get execution by id: %w", err)
	}
	exec.NodeID = nextNodeID
	exec.IsActive = true

	// 执行到达节点的处理逻辑
	// 早期版本不做设计，全部用switch case
	switch rg.Nodes[nextNodeID].Type {
	case model.NodeTypeUserTask:
		if err := s.nodeHanlderUserTask(ctx, exec); err != nil {
			return fmt.Errorf("failed to handle user task: %w", err)
		}
	case model.NodeTypeEnd:
		if err := s.nodeHandlerEndNode(ctx, exec); err != nil {
			return fmt.Errorf("failed to handle end node: %w", err)
		}
	case model.NodeTypeEmailService:
		if err := s.nodeHandlerEmailService(ctx, exec, rg); err != nil {
			return fmt.Errorf("failed to handle email service node: %w", err)
		}
	case model.NodeTypeForkGateway:
		return s.gatewayHandlerFork(ctx, exec, rg)
	case model.NodeTypeJoinGateway:
		return s.gatewayHandlerJoin(ctx, exec, rg)
	}

	// 更新Execution的当前节点
	return s.executionRepo.Update(ctx, exec)
}

// NodeHanlderUserTask 处理到达用户任务节点的逻辑
func (s *Scheduler) nodeHanlderUserTask(ctx context.Context, exec *model.Execution) error {
	// 创建用户审批任务
	task := &model.ProcessTask{
		InstanceID:  exec.InstID,
		ExecutionID: exec.ID,
		NodeID:      exec.NodeID,
		Status:      model.TaskStatusPending,
	}
	if err := s.taskCmdRepo.Create(ctx, task); err != nil {
		return err
	}

	return nil
}

// NodeHandlerEndNode 处理到达END节点的逻辑
func (s *Scheduler) nodeHandlerEndNode(ctx context.Context, exec *model.Execution) error {
	// 先删除 Execution，再更新流程实例状态
	if err := s.executionRepo.DeleteByID(ctx, exec.ID); err != nil {
		return err
	}
	if err := s.instanceRepo.UpdateStatus(ctx, exec.InstID, model.InstanceStatusApproved); err != nil {
		return err
	}

	return nil
}

func (s *Scheduler) nodeHandlerEmailService(ctx context.Context, exec *model.Execution, rg *RuntimeGraph) error {
	if s.srvTaskMQ == nil {
		return fmt.Errorf("service task mq is not initialized")
	}

	// 将邮件发送请求放到消息队列中，由独立的邮件服务消费者处理
	if err := s.srvTaskMQ.ProduceEmailTask(ctx, rg.Nodes[exec.NodeID].Properties); err != nil {
		return fmt.Errorf("failed to produce email task: %w", err)
	}
	// 流转
	nextNodeID := rg.Next[exec.NodeID][0].TargetNode
	if err := s.moveToken(ctx, exec.ID, nextNodeID, rg); err != nil {
		return fmt.Errorf("failed to move token after enqueueing email task: %w", err)
	}
	return nil
}

// gatewayHandlerFork 处理到达并行网关的节点
// TODO: 事务优化
func (s *Scheduler) gatewayHandlerFork(ctx context.Context, execution *model.Execution, rg *RuntimeGraph) error {
	// 先将父 execution 设置为 inactive
	execution.IsActive = false
	if err := s.executionRepo.Update(ctx, execution); err != nil {
		return err
	}

	// 根据 Edges 的数量创建对应的子 Execution
	edges := rg.Next[execution.NodeID]
	execs := make([]model.Execution, 0, len(edges))
	for _, edge := range edges {
		execs = append(execs, model.Execution{
			InstID:   execution.InstID,
			ParentID: execution.ID,
			NodeID:   edge.TargetNode, // 直接设置为目标节点
			IsActive: true,
		})
	}
	if err := s.executionRepo.CreateBatch(ctx, execs); err != nil {
		return err
	}

	// 对每个子 execution 调用 moveToken
	for i := range execs {
		if err := s.moveToken(ctx, execs[i].ID, execs[i].NodeID, rg); err != nil {
			return err
		}
	}

	return nil
}

// gatewayHandlerJoin 处理到达合并网关的节点
// TODO: 事务优化
func (s *Scheduler) gatewayHandlerJoin(ctx context.Context, execution *model.Execution, rg *RuntimeGraph) error {
	// 保存必要信息，因为 execution 即将被删除
	parentID := execution.ParentID
	nodeID := execution.NodeID

	// 删除当前 execution
	if err := s.executionRepo.DeleteByID(ctx, execution.ID); err != nil {
		return err
	}

	// 检查同一个 fork 下是否还有其他 active 的 execution
	count, err := s.executionRepo.GetActiveNumsByParentID(ctx, parentID)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	// count == 0，所有分支都已到达，唤醒父 execution
	if parentID == 0 {
		return fmt.Errorf("join gateway has no parent execution, nodeID: %s", nodeID)
	}

	nextNodeID := rg.Next[nodeID][0].TargetNode
	return s.moveToken(ctx, parentID, nextNodeID, rg)
}
