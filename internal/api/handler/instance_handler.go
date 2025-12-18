package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hobbyGG/Dawnix/internal/api/request"
	"github.com/hobbyGG/Dawnix/internal/service"
	"go.uber.org/zap"
)

type InstanceHandler struct {
	svc    *service.InstanceService
	logger *zap.Logger
}

func NewInstanceHandler(svc *service.InstanceService, logger *zap.Logger) *InstanceHandler {
	return &InstanceHandler{svc: svc, logger: logger}
}

func (h *InstanceHandler) Register(rg *gin.RouterGroup) {
	// 在这里注册Instance相关的路由
	r := rg.Group("instance")
	r.POST("create", h.Create)
	r.POST("list", h.List)
	r.GET(":id", h.Detail)
	r.DELETE(":id", h.Delete)
}

func (h *InstanceHandler) Create(c *gin.Context) {
	// 处理创建实例的请求
	req := new(request.CreateInstanceReq)
	if err := c.ShouldBindJSON(req); err != nil {
		h.logger.Error("failed to bind CreateInstanceReq", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// 调用服务层创建实例
	id, err := h.svc.CreateInstance(c, req.ToBizCmd())
	if err != nil {
		h.logger.Error("failed to create instance", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id})
}

func (h *InstanceHandler) List(c *gin.Context) {
	// 处理获取实例列表的请求
	req := new(request.ListInstancesReq)
	if err := c.ShouldBindJSON(req); err != nil {
		h.logger.Error("failed to bind ListInstancesReq", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	instances, err := h.svc.ListInstances(c, req.ToBizParams())
	if err != nil {
		h.logger.Error("failed to list instances", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, instances)
}

func (h *InstanceHandler) Detail(c *gin.Context) {
	// 处理获取实例详情的请求
	req := new(request.GetInstanceDetailReq)
	if err := c.ShouldBindUri(req); err != nil {
		h.logger.Error("failed to bind GetInstanceDetailReq", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	instance, err := h.svc.GetInstanceDetail(c, req.ID)
	if err != nil {
		h.logger.Error("failed to get instance detail", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, instance)
}

func (h *InstanceHandler) Delete(c *gin.Context) {
	// 处理删除实例的请求
}
