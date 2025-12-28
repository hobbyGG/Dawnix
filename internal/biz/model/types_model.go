package model

// 节点类型
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

// 数据库中的流程图结构
type GraphModel struct {
	Nodes []NodeModel `json:"nodes"`
	Edges []EdgeModel `json:"edges"`
}

type NodeModel struct {
	ID   string `json:"id"`   // 节点ID
	Type string `json:"type"` // 节点类型
	Name string `json:"name"` // 节点展示的名称

	// 节点属性，不同节点配置不同
	Properties map[string]interface{} `json:"properties"`
}

func (n *NodeModel) IsAutoType() bool {
	return n.Type == NodeTypeAITask || n.Type == NodeTypeRuleTask || n.Type == NodeTypeServiceTask
}

type EdgeModel struct {
	ID         string `json:"id"`          // 边ID
	SourceNode string `json:"source_node"` // 源节点ID
	TargetNode string `json:"target_node"` // 目标节点ID

	// 条件表达式，仅当SourceNode为网关时生效
	Condition string `json:"condition"`
}
