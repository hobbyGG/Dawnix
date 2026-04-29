package biz

import (
	"context"
	"time"

	"github.com/hobbyGG/Dawnix/internal/workflow/domain" //作用是什么
)

//定义审批记录相关的方法 结构体
//Record 审批记录结构体

type RecordRepo interface {
	Create(ctx context.Context, record *domain.Record) (int64, error)
	List(ctx context.Context, instanceID int64) ([]*domain.Record, error)
	ListAll(ctx context.Context) ([]*domain.Record, error)
	//[]*Record 这个应该怎么写 现在还没有学会？
}

// biz 里的结构体 = 全系统通用的 “数据模板”
// data层（数据库)
// service层（业务逻辑）
// api层（前端接口）
type Record struct {
	InstanceID   int64     //`json:"instance_id"`  //这边对应的格式是jason的格式和之前的格式不一样
	TaskID       int64     //`json:"task_id"`
	NodeID       string    //`json:"node_id"`
	NodeName     string    //`json:"node_name"`
	ApproverUID  string    //`json:"approver_uid"`
	ApproverName string    //`json:"approver_name"`
	Action       string    //`json:"action"`
	Comment      string    //`json:"comment"`
	CreatedAt    time.Time //`json:"created_at"`
}

func newRecordFromTask(task *domain.ProcessTask) *domain.Record {
	if task == nil {
		return nil
	}

	return &domain.Record{
		InstanceID:   task.InstanceID,
		TaskID:       task.ID,
		NodeID:       task.NodeID,
		NodeName:     "",
		ApproverUID:  task.Assignee,
		ApproverName: "",
		Action:       task.Action,
		Comment:      task.Comment,
		CreatedAt:    time.Now(),
	}
}
