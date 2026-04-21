package workflow

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hobbyGG/Dawnix/internal/workflow/biz"
	"github.com/hobbyGG/Dawnix/internal/workflow/service"
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
	req := new(GetTaskDetailReq)
	if err := c.ShouldBindUri(req); err != nil {
		writeBindError(c, h.logger, "failed to bind GetTaskDetailReq", err)
		return
	}
	taskDetailView, err := h.svc.GetTaskDetailView(c.Request.Context(), req.ID)
	if err != nil {
		writeInternalError(c, h.logger, "failed to get task detail", err)
		return
	}
	c.JSON(http.StatusOK, taskDetailView)
}

func (h *TaskHandler) List(c *gin.Context) {
	// 处理获取任务列表的请求
	req := new(ListTasksReq)
	if err := c.ShouldBindQuery(req); err != nil {
		writeBindError(c, h.logger, "failed to bind ListTasksReq", err)
		return
	}
	// 默认分页参数
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Size == 0 {
		req.Size = 10
	}

	taskListView, total, err := h.svc.ListTasksView(c.Request.Context(), req.ToBizParams())
	if err != nil {
		writeInternalError(c, h.logger, "failed to list tasks", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"total": total, "tasks": taskListView})
}

func (h *TaskHandler) Complete(c *gin.Context) {
	// 处理完成任务的请求
	req := new(CompleteTaskReq)
	if idStr, exist := c.Params.Get("id"); exist {
		var err error
		req.ID, err = strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
			return
		}
	}
	if err := c.ShouldBindJSON(req); err != nil {
		writeBindError(c, h.logger, "failed to bind CompleteTaskReq", err)
		return
	}

	if err := h.svc.CompleteTask(c.Request.Context(), req.ToBizParams()); err != nil {
		writeInternalError(c, h.logger, "failed to complete task", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

type GetTaskDetailReq struct {
	ID int64 `uri:"id" binding:"required"`
}

type ListTasksReq struct {
	// 分页参数
	Page  int    `form:"page" binding:"omitempty,min=1"`
	Size  int    `form:"size" binding:"omitempty,min=1,max=100"`
	Scope string `form:"scope" binding:"omitempty"` // 列表页范围：my_pending, my_completed, all_pending, all_completed...
}

func (req *ListTasksReq) ToBizParams() *biz.ListTasksParams {
	return &biz.ListTasksParams{
		Page:  req.Page,
		Size:  req.Size,
		Scope: req.Scope,
	}
}

type CompleteTaskReq struct {
	// 任务 ID (路径参数)
	ID int64 `uri:"id"`

	// 动作: "agree", "reject"
	Action string `json:"action" binding:"required,oneof=agree reject"`

	// 审批意见
	Comment string `json:"comment"`

	// 表单数据: 比如请假表单里的实际数据，或者审批人填写的新字段
	FormData []biz.FormDataItem `json:"form_data"`
}

func (req *CompleteTaskReq) ToBizParams() *biz.CompleteTaskParams {
	return &biz.CompleteTaskParams{
		TaskID:   req.ID,
		Action:   req.Action,
		Comment:  req.Comment,
		FormData: req.FormData,
	}
}
