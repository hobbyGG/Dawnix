package model

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/datatypes"
)

// 实例状态定义 (对齐飞书 Open API)
const (
	// InstanceStatusPending 审批中 (进行中)
	// 含义：流程启动后，尚未结束的状态。
	InstanceStatusPending = "PENDING"

	// InstanceStatusApproved 已通过 (正常结束)
	// 含义：Token 成功流转到了 "End" 节点。
	InstanceStatusApproved = "APPROVED"

	// InstanceStatusRejected 已拒绝 (异常结束)
	// 含义：审批人拒绝，且流程规则决定终止流程。
	InstanceStatusRejected = "REJECTED"

	// InstanceStatusCanceled 已撤回/已取消
	// 含义：发起人主动撤回，或者管理员强制取消。
	InstanceStatusCanceled = "CANCELED"

	// InstanceStatusSuspended 已挂起 (Dawnix 扩展状态)
	// 含义：流程被暂停（例如等待长时间外部回调，或管理员介入），不消耗计算资源。
	InstanceStatusSuspended = "SUSPENDED"
)

type ProcessInstance struct {
	BaseModel

	// 关联模版 ID
	DefinitionID int64 `gorm:"index;not null" json:"definition_id"`

	// [建议新增] 冗余流程编码，方便不 Join 表直接查询所有 "leave_flow"
	ProcessCode string `gorm:"type:varchar(64);index;not null" json:"process_code"`

	// [建议新增] 流程图快照 (核心！保护运行中实例不受模版变更影响)
	SnapshotStructure datatypes.JSON `gorm:"type:jsonb;not null" json:"snapshot_structure"`

	// 父流程 ID (支持嵌套流程)
	ParentID int64 `gorm:"index;default:0" json:"parent_id"`
	// 父流程的哪个节点启动的 (用于子流程回调)
	ParentNodeID string `gorm:"type:varchar(64)" json:"parent_node_id"`

	// 业务上下文 (JSONB)
	Variables datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"variables"`

	// 令牌桶
	ActiveTokens pq.StringArray `gorm:"type:text[];comment:当前令牌位置" json:"active_tokens"`

	// 状态
	Status string `gorm:"type:varchar(32);index;default:'PENDING'" json:"status"`

	// 发起人 ID
	SubmitterID string `gorm:"type:varchar(64);index" json:"submitter_id"`

	// 结束时间
	FinishedAt *time.Time `json:"finished_at"`
}

func (ProcessInstance) TableName() string {
	return "process_instances"
}
