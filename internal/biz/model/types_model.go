package model

const (
	// 基础控制
	NodeTypeStart    = "start"
	NodeTypeEnd      = "end"
	NodeTypeGateway  = "gateway"
	NodeTypeUserTask = "user_task"

	// 自动化节点
	NodeTypeAITask   = "ai_task"
	NodeTypeRuleTask = "rule_task"

	// 事件节点
	NodeTypeServiceTask = "service_task" // 主动投递 (HTTP/Internal)
	NodeTypeReceiveTask = "receive_task" // 被动等待 (Webhook回调)
)

type WorkflowGraph struct {
	Nodes []NodeConfig `json:"nodes"`
	Edges []EdgeConfig `json:"edges"`
}

type NodeConfig struct {
	ID   string `json:"id"`   // 节点ID
	Type string `json:"type"` // 节点类型
	Name string `json:"name"` // 节点展示的名称

	// 节点属性，不同节点配置不同
	Properties map[string]interface{} `json:"properties"`
}

type EdgeConfig struct {
	ID         string `json:"id"`          // 边ID
	SourceNode string `json:"source_node"` // 源节点ID
	TargetNode string `json:"target_node"` // 目标节点ID

	// 条件表达式，仅当SourceNode为网关时生效
	Condition string `json:"condition"`
}

type SchedulerGraph struct {
	Nodes map[string]NodeConfig // nodeID->NodeConfig的映射
}

// ProcessStructure 对应 ReactFlow 的导出对象
type ProcessStructure struct {
	// Nodes: 节点列表 (包含 id, type, data, position)
	Nodes []map[string]interface{} `json:"nodes" binding:"required"`

	// Edges: 连线列表 (包含 source, target, id)
	Edges []map[string]interface{} `json:"edges" binding:"required"`

	// Viewport: 视口状态 (x, y, zoom)，用于用户下次打开时恢复视角
	Viewport map[string]interface{} `json:"viewport"`
}

// ProcessConfig 流程全局配置
type ProcessConfig struct {
	// AutoApprove: 是否开启自动去重/自动通过
	AutoApprove bool `json:"auto_approve"`

	// Timeout: 全局超时时间 (秒)，0表示不超时
	Timeout int64 `json:"timeout"`

	// CallbackURL: 流程结束后的回调地址 (Webhook)
	CallbackURL string `json:"callback_url"`
}
