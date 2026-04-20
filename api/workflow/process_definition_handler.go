package workflow

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hobbyGG/Dawnix/internal/workflow/biz"
	"github.com/hobbyGG/Dawnix/internal/workflow/domain"
	"github.com/hobbyGG/Dawnix/internal/workflow/service"
	"go.uber.org/zap"
)

type ProcessDefinitionHandler struct {
	svc    *service.ProcessDefinitionService
	logger *zap.Logger
}

func NewProcessDefinitionHandler(svc *service.ProcessDefinitionService, logger *zap.Logger) *ProcessDefinitionHandler {
	return &ProcessDefinitionHandler{svc: svc, logger: logger}
}

func (h *ProcessDefinitionHandler) Register(rg *gin.RouterGroup) {
	// 这里注册ProcessDefinition相关的路由
	r := rg.Group("definition")
	// 创建流程模板
	r.POST("create", h.Create)

	// 获取流程模板详情
	r.GET("list", h.List)

	// 获取流程模板列表
	r.GET(":id", h.Detail)

	// 编辑流程模板
	r.PUT(":id", h.Update)

	// 删除流程模板
	r.DELETE(":id", h.Delete)
}

func (h *ProcessDefinitionHandler) Create(c *gin.Context) {
	// 处理创建流程模板的请求
	req := new(ProcessDefinitionCreateReq)
	if err := c.ShouldBindJSON(req); err != nil {
		h.logger.Error("failed to bind ProcessDefinitionCreateReq", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := h.svc.CreateProcessDefinition(c, req.ToBizParams())
	if err != nil {
		h.logger.Error("failt to CreateProcessDefinition", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// 对响应进行封装
	c.JSON(http.StatusOK, gin.H{"id": id})
}

func (h *ProcessDefinitionHandler) List(c *gin.Context) {
	// 处理获取流程模板列表的请求
	req := new(ProcessDefinitionListReq)
	if err := c.ShouldBind(req); err != nil {
		h.logger.Error("failed to bind ProcessDefinitionListReq", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pdList, err := h.svc.ListProcessDefinitions(c, req.ToBizParams())
	if err != nil {
		h.logger.Error("failt to ListProcessDefinitions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// 对响应进行封装
	resp := &ProcessDefinitionListResp{
		Total: int64(len(pdList)),
		List:  pdList,
	}
	c.JSON(http.StatusOK, resp)
}

func (h *ProcessDefinitionHandler) Detail(c *gin.Context) {
	// 处理获取流程模板详情的请求
	req := new(ProcessDefinitionDetailReq)
	if err := c.ShouldBindUri(req); err != nil {
		h.logger.Error("failed to bind ProcessDefinitionDetailReq", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	pdDetail, err := h.svc.GetProcessDefinitionDetail(c, req.ID)
	if err != nil {
		h.logger.Error("fail to GetProcessDefinitionDetail", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// 对响应进行封装
	c.JSON(http.StatusOK, pdDetail)
}

func (h *ProcessDefinitionHandler) Delete(c *gin.Context) {
	req := new(ProcessDefinitionDeleteReq)
	if err := c.ShouldBindUri(req); err != nil {
		h.logger.Error("failed to bind ProcessDefinitionDeleteReq", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// 处理删除流程模板的请求
	if err := h.svc.DeleteProcessDefinition(c, req.ID); err != nil {
		h.logger.Error("fail to DeleteProcessDefinition", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted success"})
}

func (h *ProcessDefinitionHandler) Update(c *gin.Context) {
	req := new(ProcessDefinitionUpdateReq)
	if err := c.ShouldBindUri(req); err != nil {
		h.logger.Error("failed to bind ProcessDefinitionUpdateReq uri", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := c.ShouldBindJSON(req); err != nil {
		h.logger.Error("failed to bind ProcessDefinitionUpdateReq body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.UpdateProcessDefinition(c, req.ID, req.ProcessDefinitionCreateReq.ToBizParams()); err != nil {
		h.logger.Error("fail to UpdateProcessDefinition", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "updated success"})
}

type ProcessDefinitionCreateReq struct {
	Code           string             `json:"code"`                         // 流程模板业务号，用于创建流程
	Name           string             `json:"name" binding:"required"`      // 流程模板名称
	Structure      ProcessStructure   `json:"structure" binding:"required"` // 流程模板图结构
	FormDefinition []biz.FormDataItem `json:"form_definition"`              // 表单定义项列表
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
	ID         string            `json:"id"`                   // 节点ID
	Type       string            `json:"type"`                 // 节点类型
	Name       string            `json:"name"`                 // 节点展示的名称
	Candidates domain.Candidates `json:"candidates,omitempty"` // 候选人，仅用户任务节点有效

	Properties json.RawMessage `json:"properties,omitempty"` // 其他属性，针对不同类型节点的特有属性，例如邮件服务节点的邮件参数等
}

type ProcessEdge struct {
	ID         string `json:"id"`     // 边ID
	SourceNode string `json:"source"` // 源节点ID
	TargetNode string `json:"target"` // 目标节点ID
	Condition  string `json:"condition"`
	IsDefault  bool   `json:"is_default"`
}

func (r *ProcessDefinitionCreateReq) ToBizParams() *biz.ProcessDefinitionCreateParams {
	// 这里需要将 ProcessStructure 转换为 graph结构
	graph := domain.GraphModel{
		Nodes: []domain.NodeModel{},
		Edges: []domain.EdgeModel{},
	}
	// 转换 Nodes
	for _, node := range r.Structure.Nodes {
		workflowNode := domain.NodeModel{
			ID:         node.ID,
			Type:       node.Type,
			Name:       node.Name,
			Candidates: node.Candidates,
			Properties: node.Properties,
		}
		graph.Nodes = append(graph.Nodes, workflowNode)
	}
	// 转换 Edges
	for _, edge := range r.Structure.Edges {
		workflowEdge := domain.EdgeModel{
			ID:         edge.ID,
			SourceNode: edge.SourceNode,
			TargetNode: edge.TargetNode,
			Condition:  edge.Condition,
			IsDefault:  edge.IsDefault,
		}
		graph.Edges = append(graph.Edges, workflowEdge)
	}
	return &biz.ProcessDefinitionCreateParams{
		Name:           r.Name,
		Code:           r.Code,
		Structure:      &graph,
		FormDefinition: r.FormDefinition,
	}
}

type ProcessDefinitionListReq struct {
	Page int `form:"page" binding:"required,omitempty,min=1"`        // 页码
	Size int `form:"size" binding:"required,omitempty,min=1,max=50"` // 每页数量
}

type ProcessDefinitionListResp struct {
	Total int64                      `json:"total"` // 总记录数
	List  []domain.ProcessDefinition `json:"list"`  // 流程模板列表
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

type ProcessDefinitionUpdateReq struct {
	ID int64 `uri:"id" binding:"required,min=1"` // 流程模板ID
	ProcessDefinitionCreateReq
}
