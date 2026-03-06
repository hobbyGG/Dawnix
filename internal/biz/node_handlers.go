package biz

import (
	"context"
	"fmt"

	"github.com/hobbyGG/Dawnix/internal/biz/model"
)

// NodeHandlerFunc 节点到达时的处理函数签名
type NodeHandlerFunc func(ctx context.Context, node *model.NodeModel, exec *model.Execution) error

// nodeHandlers 持有节点处理所需的依赖，方法作为 NodeHandlerFunc 注册到 Scheduler
type nodeHandlers struct {
	taskCmdRepo   TaskCommandRepo
	executionRepo ExecutionRepo
	instanceRepo  InstanceRepo
}

// userTask 到达用户任务节点：创建待办任务
func (h *nodeHandlers) userTask(ctx context.Context, node *model.NodeModel, exec *model.Execution) error {
	task := &model.ProcessTask{
		InstanceID:  exec.InstID,
		ExecutionID: exec.ID,
		Type:        model.NodeTypeUserTask,
		Status:      model.TaskStatusPending,
	}
	if err := h.taskCmdRepo.Create(ctx, task); err != nil {
		return fmt.Errorf("userTask: failed to create task: %w", err)
	}
	return nil
}

// endNode 到达结束节点：销毁 execution，标记流程实例完成
func (h *nodeHandlers) endNode(ctx context.Context, node *model.NodeModel, exec *model.Execution) error {
	if err := h.executionRepo.DeleteByID(ctx, exec.ID); err != nil {
		return fmt.Errorf("endNode: failed to delete execution: %w", err)
	}
	if err := h.instanceRepo.UpdateStatus(ctx, exec.InstID, model.InstanceStatusApproved); err != nil {
		return fmt.Errorf("endNode: failed to update instance status: %w", err)
	}
	return nil
}
