package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hobbyGG/Dawnix/internal/api/request"
	"github.com/hobbyGG/Dawnix/internal/service"
	"go.uber.org/zap"
)

type TaskHandler struct {
	svc    *service.TaskService
	logger *zap.Logger
}

func NewTaskHandler(svc *service.TaskService, logger *zap.Logger) *TaskHandler {
	return &TaskHandler{svc: svc, logger: logger}
}

func (h *TaskHandler) Register(rg *gin.RouterGroup) {
	// 在这里注册Task相关的路由
	r := rg.Group("tasks")
	r.GET(":id", h.Detail)
	r.POST("complete/:id", h.Complete)
}

func (h *TaskHandler) Detail(c *gin.Context) {
	// 处理获取任务详情的请求
	req := new(request.GetTaskDetailReq)
	if err := c.ShouldBindUri(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	taskDetail, err := h.svc.GetTaskDetail(c.Request.Context(), req.ID)
	if err != nil {
		h.logger.Error("failed to get task detail", zap.Error(err))
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, taskDetail)
}

func (h *TaskHandler) Complete(c *gin.Context) {
	// 处理完成任务的请求
	req := new(request.CompleteTaskReq)
	if idStr, exist := c.Params.Get("id"); exist {
		var err error
		req.ID, err = strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
			return
		}
	}
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.CompleteTask(c.Request.Context(), req.ToBizParams()); err != nil {
		h.logger.Error("failed to complete task", zap.Error(err))
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}
