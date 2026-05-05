package domain

// type NodeType string

// func (n *NodeType) IsGateway() bool {

// }

const (
	NodeTypeStart            = "start"
	NodeTypeEnd              = "end"
	NodeTypeUserTask         = "user_task"
	NodeTypeEmailService     = "email_service"
	NodeTypeForkGateway      = "fork_gateway"
	NodeTypeJoinGateway      = "join_gateway"
	NodeTypeXORGateway       = "xor_gateway"
	NodeTypeInclusiveGateway = "inclusive_gateway"
)

type GraphModel struct {
	Nodes    []NodeModel            `json:"nodes"`
	Edges    []EdgeModel            `json:"edges"`
	Viewport map[string]interface{} `json:"viewport,omitempty"`
}

func (g *GraphModel) EdgesBySource() map[string][]EdgeModel {
	grouped := make(map[string][]EdgeModel)
	if g == nil {
		return grouped
	}
	for _, edge := range g.Edges {
		grouped[edge.SourceNode] = append(grouped[edge.SourceNode], edge)
	}
	return grouped
}

type NodeModel struct {
	ID         string     `json:"id"`
	Type       string     `json:"type"`
	Name       string     `json:"name"`
	Candidates Candidates `json:"candidates,omitempty"`
	Properties []byte     `json:"properties,omitempty"`
	Position   struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
	} `json:"position,omitempty"`
}

type EdgeModel struct {
	ID         string `json:"id"`
	SourceNode string `json:"source_node"`
	TargetNode string `json:"target_node"`
	Condition  string `json:"condition"`
	IsDefault  bool   `json:"is_default,omitempty"`
}

type Candidates struct {
	Users []string `json:"users"`
}

type EmailNodeParams struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

func IsGateway(nodeType string) bool {
	if nodeType != NodeTypeForkGateway && nodeType != NodeTypeInclusiveGateway && nodeType != NodeTypeXORGateway {
		return false
	}
	return true
}
