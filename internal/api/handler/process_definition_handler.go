package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hobbyGG/Dawnix/internal/api/request"
	"github.com/hobbyGG/Dawnix/internal/service"
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
	r.POST("list", h.List)

	// 获取流程模板列表
	r.GET(":id", h.Detail)

	// 删除流程模板
	r.DELETE(":id", h.Delete)
}

func (h *ProcessDefinitionHandler) Create(c *gin.Context) {
	// 处理创建流程模板的请求
	req := new(request.ProcessDefinitionCreateReq)
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
	req := new(request.ProcessDefinitionListReq)
	if err := c.ShouldBindJSON(req); err != nil {
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
	c.JSON(http.StatusOK, pdList)
}

func (h *ProcessDefinitionHandler) Detail(c *gin.Context) {
	// 处理获取流程模板详情的请求
	req := new(request.ProcessDefinitionDetailReq)
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
	req := new(request.ProcessDefinitionDeleteReq)
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
