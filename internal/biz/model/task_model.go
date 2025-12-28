package model

import (
	"time"

	"gorm.io/datatypes"
)

// 状态定义 (完全对齐飞书 Open API)
const (
	// TaskStatusPending 审批中/待处理
	// 初始状态，Token 停留在该节点
	TaskStatusPending = "PENDING"

	// TaskStatusApproved 已通过
	// 动作：同意 (Agree)。Token 流向下一节点。
	TaskStatusApproved = "APPROVED"

	// TaskStatusRejected 已拒绝
	// 动作：拒绝 (Reject)。Token 销毁，流程通常直接结束或标记为失败。
	TaskStatusRejected = "REJECTED"

	// TaskStatusTransferred 已转交
	// 动作：转交 (Transfer)。当前任务结束，Token 不动，但在当前节点生成一个新的 Task 给被转交人。
	TaskStatusTransferred = "TRANSFERRED"

	// TaskStatusRolledBack 已退回
	// 动作：退回 (Rollback)。Token 跳回到之前的某个节点。
	TaskStatusRolledBack = "ROLLED_BACK"

	// TaskStatusCanceled 已取消
	// 场景：
	// 1. 发起人撤回流程，所有进行中任务变为 Canceled
	// 2. 并发分支中，由于"或签"规则，一人通过，其他人任务自动变为 Canceled
	TaskStatusCanceled = "CANCELED"

	// TaskStatusAborted 已终止 (异常)
	// 场景：Service Task 执行失败且重试耗尽
	TaskStatusAborted = "ABORTED"
)

const (
	TaskTypeUser    = "user_task"    // 人工审批
	TaskTypeService = "service_task" // 系统自动执行
	TaskTypeReceive = "receive_task" // 等待外部回调
	TaskTypeCc      = "cc_task"      // 抄送任务
)

type ProcessTask struct {
	BaseModel

	InstanceID int64 `gorm:"index;not null" json:"instance_id"`

	// 节点ID (对应 JSON 图里的 id)
	NodeID string `gorm:"type:varchar(64);not null" json:"node_id"`

	// 任务类型
	Type string `gorm:"type:varchar(32);default:'user_task'" json:"type"`

	// 处理人标识
	// "user:1001", "role:manager", "sys:http"
	Assignee string `gorm:"type:varchar(64);index" json:"assignee"`

	Candidates datatypes.JSON `gorm:"type:jsonb" json:"candidates"`

	// 【核心修改】任务状态
	// 使用上述 TaskStatus* 常量
	Status string `gorm:"type:varchar(32);default:'PENDING';index" json:"status"`

	// 具体的按钮动作 (辅助字段)
	// 虽然 Status 已经是结果了，但 Action 用于记录用户点击了哪个按钮
	// 例如 Status=APPROVED, Action 可能为 "agree" 或 "fast_pass"(一键通过)
	Action string `gorm:"type:varchar(32)" json:"action"`

	// 审批意见 / 备注
	Comment string `gorm:"type:text" json:"comment"`

	// 任务局部变量 (快照)
	Variables datatypes.JSON `gorm:"type:jsonb" json:"variables"`

	// 超时与重试
	RetryCount int        `gorm:"default:0" json:"retry_count"`
	ExpireAt   *time.Time `gorm:"index" json:"expire_at"`

	CostTime   int64      `json:"cost_time"`
	FinishedAt *time.Time `json:"finished_at"`
}

func (ProcessTask) TableName() string {
	return "process_tasks"
}

// NewTask 创建一个新的 ProcessTask 实例
func NewTask(inst *ProcessInstance, nodeID, taskType string) *ProcessTask {
	return &ProcessTask{
		InstanceID: inst.ID,
		NodeID:     nodeID,
		Type:       taskType,
		Status:     TaskStatusPending,
	}
}

// TaskView 这是你的【查询结果】，用于 CQRS 的读操作
// 它不是一张表，它是多张表 Join 后的"投影"
type TaskView struct {
	ID            int64     `json:"id"`
	TaskName      string    `json:"task_name"`
	Status        string    `json:"status"`
	ProcessTitle  string    `json:"process_title"`  // 来自 definition 表
	SubmitterName string    `json:"submitter_name"` // 来自 instance 表
	ArrivedAt     time.Time `json:"arrived_at"`
}
