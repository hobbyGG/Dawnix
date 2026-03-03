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
	r := rg.Group("task")
	r.GET(":id", h.Detail)
	r.GET("list", h.List)
	r.POST("complete/:id", h.Complete)
}

func (h *TaskHandler) Detail(c *gin.Context) {
	// 处理获取任务详情的请求
	req := new(request.GetTaskDetailReq)
	if err := c.ShouldBindUri(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	taskDetailView, err := h.svc.GetTaskDetailView(c.Request.Context(), req.ID)
	if err != nil {
		h.logger.Error("failed to get task detail", zap.Error(err))
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, taskDetailView)
}

func (h *TaskHandler) List(c *gin.Context) {
	// 处理获取任务列表的请求
	req := new(request.ListTasksReq)
	if err := c.ShouldBindQuery(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// 默认分页参数
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Size == 0 {
		req.Size = 10
	}
	ListTasksParams := req.ToBizParams()
	ListTasksParams.UserID = "umep123" // TODO: 从中间件获取当前用户ID

	taskListView, total, err := h.svc.ListTasksView(c.Request.Context(), ListTasksParams)
	if err != nil {
		h.logger.Error("failed to list tasks", zap.Error(err))
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"total": total, "tasks": taskListView})
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
