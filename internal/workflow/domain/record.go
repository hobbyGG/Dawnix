package domain

import (
	"time"
	//"gorm.io/datatypes"
)

// const (  //const的作用是什么？ 业务里固定不变、全系统通用、用来做判断 / 分类的枚举 → 全部放 domain + const

// )

const ( //不知道这个常量是否真的需要？4.26
	HistoryStatusPending   = "PENDING"
	HistoryStatusApproved  = "APPROVED"
	HistoryStatusRejected  = "REJECTED"
	HistoryStatusCanceled  = "CANCELED"
	HistoryStatusSuspended = "SUSPENDED"
)

// type Record struct {
// 	BaseModel

// }

// Record 审批记录领域实体
// 纯数据、无接口、无标签、无逻辑
type Record struct {
	//BaseModel
	ID           int64     `json:"id"`
	InstanceID   int64     `json:"instance_id"`
	TaskID       int64     `json:"task_id"`
	NodeID       string    `json:"node_id"`
	NodeName     string    `json:"node_name"`
	ApproverUID  string    `json:"approver_uid"`
	ApproverName string    `json:"approver_name"`
	Action       string    `json:"action"`
	Comment      string    `json:"comment"`
	CreatedAt    time.Time `json:"created_at"`
}
