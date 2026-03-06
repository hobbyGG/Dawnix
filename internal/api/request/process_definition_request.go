package request

import (
	"github.com/hobbyGG/Dawnix/internal/biz"
	"github.com/hobbyGG/Dawnix/internal/biz/model"
)

type ProcessDefinitionCreateReq struct {
	Code      string           `json:"code"`                         // 流程模板业务号，用于创建流程
	Name      string           `json:"name" binding:"required"`      // 流程模板名称
	Structure ProcessStructure `json:"structure" binding:"required"` // 流程模板图结构
	Config    ProcessConfig    `json:"config"`                       // 流程全局配置，例如该流程结束后处理配置等
}

// ProcessStructure 对应 ReactFlow 的导出对象
type ProcessStructure struct {
	// Nodes: 节点列表 (包含 id, type, data, position)
	Nodes []ProcessNode `json:"nodes" binding:"required"`

	// Edges: 连线列表 (包含 source, target, id)
	Edges []ProcessEdge `json:"edges" binding:"required"`

	// Viewport: 视口状态 (x, y, zoom)，用于用户下次打开时恢复视角
	Viewport map[string]interface{} `json:"viewport"`
}

type ProcessNode struct {
	ID         string           `json:"id"`                   // 节点ID
	Type       string           `json:"type"`                 // 节点类型
	Name       string           `json:"name"`                 // 节点展示的名称
	Candidates model.Candidates `json:"candidates,omitempty"` // 候选人，仅用户任务节点有效
}

type ProcessEdge struct {
	ID         string `json:"id"`     // 边ID
	SourceNode string `json:"source"` // 源节点ID
	TargetNode string `json:"target"` // 目标节点ID
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

func (r *ProcessDefinitionCreateReq) ToBizParams() *biz.ProcessDefinitionCreateParams {
	// 这里需要将 ProcessStructure 转换为 graph结构
	graph := model.GraphModel{
		Nodes: []model.NodeModel{},
		Edges: []model.EdgeModel{},
	}
	// 转换 Nodes
	for _, node := range r.Structure.Nodes {
		workflowNode := model.NodeModel{
			ID:         node.ID,
			Type:       node.Type,
			Name:       node.Name,
			Candidates: node.Candidates,
		}
		graph.Nodes = append(graph.Nodes, workflowNode)
	}
	// 转换 Edges
	for _, edge := range r.Structure.Edges {
		workflowEdge := model.EdgeModel{
			ID:         edge.ID,
			SourceNode: edge.SourceNode,
			TargetNode: edge.TargetNode,
		}
		graph.Edges = append(graph.Edges, workflowEdge)
	}
	return &biz.ProcessDefinitionCreateParams{
		Name:      r.Name,
		Code:      r.Code,
		Structure: &graph,
	}
}

type ProcessDefinitionListReq struct {
	Page int `form:"page" binding:"required,omitempty,min=1"`        // 页码
	Size int `form:"size" binding:"required,omitempty,min=1,max=50"` // 每页数量
}

type ProcessDefinitionListResp struct {
	Total int64                     `json:"total"` // 总记录数
	List  []model.ProcessDefinition `json:"list"`  // 流程模板列表
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
