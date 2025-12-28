package model

import (
	"slices"
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

	// 冗余流程编码，方便不 Join 表直接查询所有 "leave_flow"
	ProcessCode string `gorm:"type:varchar(64);index;not null" json:"process_code"`

	// 流程图快照 (核心！保护运行中实例不受模版变更影响)
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

func (inst *ProcessInstance) ConsumeToken(nodeID string) {
	// 1. 查找第一个匹配项的索引
	idx := slices.Index(inst.ActiveTokens, nodeID)
	if idx == -1 {
		return // 没找到令牌，直接返回
	}

	// 2. 使用 slices.Delete 删除
	// 注意：Delete 会处理好底层元素重排
	inst.ActiveTokens = slices.Delete(inst.ActiveTokens, idx, idx+1)
}

// ProduceToken 产生（增加）一个新令牌，并去重
func (i *ProcessInstance) ProduceToken(nodeID string) {
	for _, t := range i.ActiveTokens {
		if t == nodeID {
			return // 已经存在，不再重复添加
		}
	}
	i.ActiveTokens = append(i.ActiveTokens, nodeID)
}

// HasToken 检查是否存在某个令牌
func (i *ProcessInstance) HasToken(nodeID string) bool {
	for _, t := range i.ActiveTokens {
		if t == nodeID {
			return true
		}
	}
	return false
}
