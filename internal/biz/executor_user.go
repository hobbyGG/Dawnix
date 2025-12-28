package biz

import "github.com/hobbyGG/Dawnix/internal/biz/model"

type UserNodeBehaviour struct {
	taskRepo TaskCommandRepo
}

func NewUserNodeBehaviour(taskRepo TaskCommandRepo) *UserNodeBehaviour {
	return &UserNodeBehaviour{
		taskRepo: taskRepo,
	}
}

func (e *UserNodeBehaviour) OnEnter(nodeCtx *NodeContext, currentNode *model.NodeModel) (bool, error) {
	// 用户节点操作
	// 发起任务
	task := &model.ProcessTask{
		// 填充任务信息
		InstanceID: nodeCtx.inst.ID,
		NodeID:     currentNode.ID,
		Type:       model.TaskTypeUser,
		Status:     model.TaskStatusPending,
	}
	return false, e.taskRepo.Create(nodeCtx.ctx, task)
}
