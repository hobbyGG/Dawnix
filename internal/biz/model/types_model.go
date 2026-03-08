package model

// 节点类型
const (
	NodeTypeStart    = "start"
	NodeTypeEnd      = "end"
	NodeTypeUserTask = "user_task"

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
