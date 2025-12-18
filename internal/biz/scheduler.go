package biz

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hobbyGG/Dawnix/internal/biz/model"
)

type Scheduler struct {
	txManager      TransactionManager
	definitionRepo ProcessDefinitionRepo
	instanceRepo   InstanceRepo
	taskCmdRepo    TaskCommandRepo
}

func NewScheduler(txManager TransactionManager, definitionRepo ProcessDefinitionRepo, instanceRepo InstanceRepo, taskCmdRepo TaskCommandRepo) *Scheduler {
	return &Scheduler{
		txManager:      txManager,
		definitionRepo: definitionRepo,
		instanceRepo:   instanceRepo,
		taskCmdRepo:    taskCmdRepo,
	}
}

func (s *Scheduler) StartProcessInstance(ctx context.Context, cmd StartProcessInstanceCmd) (int64, error) {
	// 由流程引擎负责整个创建实例的流程
	// 根据cmd的信息创建流程实例
	var instID int64
	s.txManager.InTx(ctx, func(ctx context.Context) error {
		// 开启事务操作
		// 根据code找到对应流程模板
		processDef, err := s.definitionRepo.GetByCode(ctx, cmd.ProcessCode)
		if err != nil {
			return fmt.Errorf("Get definition failed, %w", err)
		}
		// 从def中拿到SchedulerGraph
		var graph model.WorkflowGraph
		if err := json.Unmarshal(processDef.Structure, &graph); err != nil {
			return fmt.Errorf("structure unmarshal failed, %w", err)
		}
		runtimeGraph := NewSchedulerGraph(&graph)
		startEdges := runtimeGraph.Next[runtimeGraph.StartNode.ID]
		if len(startEdges) != 1 {
			return fmt.Errorf("start node should have only one outgoing edge")
		}
		firstProcessNodeID := startEdges[0].TargetNode
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
			ActiveTokens:      []string{firstProcessNodeID},
			Status:            model.InstanceStatusPending,
			SubmitterID:       cmd.SubmitterID,
		}
		instID, err = s.instanceRepo.Create(ctx, inst)
		if err != nil {
			return fmt.Errorf("creatr instance failed, %w", err)
		}

		// 创建对应的任务
		task := &model.ProcessTask{
			InstanceID: instID,
			NodeID:     firstProcessNodeID,
			Type:       model.TaskTypeUser,
			Candidates: []byte("umep"),
		}
		if err = s.taskCmdRepo.Create(ctx, task); err != nil {
			return fmt.Errorf("task create failed, %w", err)
		}

		return nil
	})

	return instID, nil
}

// 核心流转逻辑
func (s *Scheduler) MoveToken(ctx context.Context, instance *model.ProcessInstance) error {
	// 【重点】使用 s.tx.InTx 包裹业务逻辑
	// 这里的 ctx 是普通的 context
	return s.txManager.InTx(ctx, func(ctx context.Context) error {
		// 【重点】这里的 ctx (闭包参数) 是“带事务”的 context
		// 把它传给 Repo，Repo 就会自动使用同一个事务

		// 1. 更新实例状态 (使用事务 ctx)
		if err := s.instanceRepo.Update(ctx, instance); err != nil {
			return err // 返回 error 会自动回滚
		}

		// 2. 创建新任务 (使用事务 ctx)
		newTask := &model.ProcessTask{ /*...*/ }
		if err := s.taskCmdRepo.Create(ctx, newTask); err != nil {
			return err // 返回 error 会自动回滚
		}

		return nil // 返回 nil 会自动提交
	})
}
