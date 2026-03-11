package model

// 节点类型
const (
	// 普通节点
	NodeTypeStart    = "start"
	NodeTypeEnd      = "end"
	NodeTypeUserTask = "user_task"

	// 服务节点
	NodeTypeEmailService = "email_service"

	// 网关节点
	NodeTypeForkGateway = "fork_gateway"
	NodeTypeJoinGateway = "join_gateway"
	NodeTypeXORGateway  = "xor_gateway"
)

// 数据库中的流程图结构
type GraphModel struct {
	Nodes []NodeModel `json:"nodes"`
	Edges []EdgeModel `json:"edges"`
}

type NodeModel struct {
	ID         string     `json:"id"`                   // 节点ID
	Type       string     `json:"type"`                 // 节点类型
	Name       string     `json:"name"`                 // 节点展示的名称
	Candidates Candidates `json:"candidates,omitempty"` // 候选人，仅用户任务节点有效

	// 额外参数
	Properties []byte `json:"properties,omitempty"` // 预留字段，存储节点的额外参数，格式为JSON
}

type EdgeModel struct {
	ID         string `json:"id"`          // 边ID
	SourceNode string `json:"source_node"` // 源节点ID
	TargetNode string `json:"target_node"` // 目标节点ID

	// 条件表达式，仅当SourceNode为网关时生效
	Condition string `json:"condition"`
}

type Candidates struct {
	Users []string `json:"users"` // 用户列表
}

// 定义 email service节点的参数结构
type EmailNodeParmas struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}
