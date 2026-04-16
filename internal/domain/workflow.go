package domain

const (
	NodeTypeStart        = "start"
	NodeTypeEnd          = "end"
	NodeTypeUserTask     = "user_task"
	NodeTypeEmailService = "email_service"
	NodeTypeForkGateway  = "fork_gateway"
	NodeTypeJoinGateway  = "join_gateway"
	NodeTypeXORGateway   = "xor_gateway"
)

type GraphModel struct {
	Nodes []NodeModel `json:"nodes"`
	Edges []EdgeModel `json:"edges"`
}

type NodeModel struct {
	ID         string     `json:"id"`
	Type       string     `json:"type"`
	Name       string     `json:"name"`
	Candidates Candidates `json:"candidates,omitempty"`
	Properties []byte     `json:"properties,omitempty"`
}

type EdgeModel struct {
	ID         string `json:"id"`
	SourceNode string `json:"source_node"`
	TargetNode string `json:"target_node"`
	Condition  string `json:"condition"`
}

type Candidates struct {
	Users []string `json:"users"`
}

type EmailNodeParams struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}
