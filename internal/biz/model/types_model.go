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
