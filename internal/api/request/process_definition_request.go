package request

import "github.com/hobbyGG/Dawnix/internal/biz"

type ProcessDefinitionCreateReq struct {
	Code      string           `json:"code"`                         // 流程模板业务号，用于创建流程
	Name      string           `json:"name" binding:"required"`      // 流程模板名称
	Structure ProcessStructure `json:"structure" binding:"required"` // 流程模板图结构
	Config    ProcessConfig    `json:"config"`                       // 流程全局配置，例如该流程结束后处理配置等
}

func (r *ProcessDefinitionCreateReq) ToBizParams() *biz.ProcessDefinitionCreateParams {
	// 这里需要将 ProcessStructure 转换为 biz 所需的 WorkflowGraph 结构
	workflowGraph := biz.WorkflowGraph{
		Nodes: []biz.WorkflowNode{},
		Edges: []biz.WorkflowEdge{},
	}
	// 转换 Nodes
	for _, node := range r.Structure.Nodes {
		workflowNode := biz.WorkflowNode{
			ID:   node["id"].(string),
			Type: node["type"].(string),
			Name: node["name"].(string),
		}
		workflowGraph.Nodes = append(workflowGraph.Nodes, workflowNode)
	}
	// 转换 Edges
	for _, edge := range r.Structure.Edges {
		workflowEdge := biz.WorkflowEdge{
			ID:         edge["id"].(string),
			SourceNode: edge["source"].(string),
			TargetNode: edge["target"].(string),
		}
		workflowGraph.Edges = append(workflowGraph.Edges, workflowEdge)
	}
	return &biz.ProcessDefinitionCreateParams{
		Name:      r.Name,
		Code:      r.Code,
		Structure: workflowGraph,
	}
}

type ProcessDefinitionListReq struct {
	Page int `json:"page" binding:"required,min=1"` // 页码，从1开始
	Size int `json:"size" binding:"required,min=1"` // 每页大小
}

func (r *ProcessDefinitionListReq) ToBizParams() *biz.ProcessDefinitionListParams {
	return &biz.ProcessDefinitionListParams{
		Page: r.Page,
		Size: r.Size,
	}
}

type ProcessDefinitionDetailReq struct {
	ID int64 `uri:"id" binding:"required,min=1"` // 流程模板ID
}

type ProcessDefinitionDeleteReq struct {
	ID int64 `uri:"id" binding:"required,min=1"` // 流程模板ID
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
