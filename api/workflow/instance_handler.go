package workflow

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hobbyGG/Dawnix/internal/workflow/biz"
	"github.com/hobbyGG/Dawnix/internal/workflow/service"
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
	r.GET("list", h.List)
	r.GET(":id", h.Detail)
	r.POST("delete/:id", h.Delete)
}

func (h *InstanceHandler) Create(c *gin.Context) {
	// 处理创建实例的请求
	uid, err := getUIDFromCtx(c)
	if err != nil {
		h.logger.Error("failed to get uid from context", zap.Error(err))
		writeUnauthorized(c)
		return
	}

	req := new(CreateInstanceReq)
	if err := c.ShouldBindJSON(req); err != nil {
		writeBindError(c, h.logger, "failed to bind CreateInstanceReq", err)
		return
	}
	params := req.ToBizParams()
	params.SubmitterID = uid

	id, err := h.svc.CreateInstance(c.Request.Context(), params)
	if err != nil {
		writeInternalError(c, h.logger, "failed to create instance", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id})
}

func (h *InstanceHandler) List(c *gin.Context) {
	// 处理获取实例列表的请求
	req := new(ListInstancesReq)
	if err := c.ShouldBindQuery(req); err != nil {
		writeBindError(c, h.logger, "failed to bind ListInstancesReq", err)
		return
	}
	instances, err := h.svc.ListInstances(c.Request.Context(), req.ToBizParams())
	if err != nil {
		writeInternalError(c, h.logger, "failed to list instances", err)
		return
	}

	listItems := make([]InstanceListItem, 0, len(instances))
	for _, inst := range instances {
		listItems = append(listItems, InstanceListItem{
			ID:          inst.ID,
			ProcessCode: inst.ProcessCode,
			ProcessName: "",
			Status:      inst.Status,
			SubmitterID: inst.SubmitterID,
			CreatedAt:   inst.CreatedAt,
			FinishedAt:  inst.FinishedAt,
		})
	}
	c.JSON(http.StatusOK, InstanceListReply{Total: int64(len(instances)), List: listItems})
}

func (h *InstanceHandler) Detail(c *gin.Context) {
	// 处理获取实例详情的请求
	req := new(GetInstanceDetailReq)
	if err := c.ShouldBindUri(req); err != nil {
		writeBindError(c, h.logger, "failed to bind GetInstanceDetailReq", err)
		return
	}
	instance, err := h.svc.GetInstanceDetail(c.Request.Context(), req.ID)
	if err != nil {
		writeInternalError(c, h.logger, "failed to get instance detail", err)
		return
	}
	var reply *InstanceDetailReply
	if instance != nil && instance.Inst != nil {
		detailItem := InstanceDetailItem{
			ID:                instance.Inst.ID,
			DefinitionID:      instance.Inst.DefinitionID,
			ProcessCode:       instance.Inst.ProcessCode,
			SnapshotStructure: json.RawMessage(instance.Inst.SnapshotStructure),
			ParentID:          instance.Inst.ParentID,
			ParentNodeID:      instance.Inst.ParentNodeID,
			FormData:          json.RawMessage(instance.Inst.FormData),
			Status:            instance.Inst.Status,
			SubmitterID:       instance.Inst.SubmitterID,
			FinishedAt:        instance.Inst.FinishedAt,
			CreatedAt:         instance.Inst.CreatedAt,
			UpdatedAt:         instance.Inst.UpdatedAt,
			CreatedBy:         instance.Inst.CreatedBy,
			UpdatedBy:         instance.Inst.UpdatedBy,
		}
		reply = &InstanceDetailReply{Instance: detailItem}
		if len(instance.Executions) > 0 {
			reply.Executions = make([]ExecutionReply, 0, len(instance.Executions))
			for _, exec := range instance.Executions {
				reply.Executions = append(reply.Executions, ExecutionReply{
					ID:        exec.ID,
					InstID:    exec.InstID,
					ParentID:  exec.ParentID,
					NodeID:    exec.NodeID,
					IsActive:  exec.IsActive,
					CreatedAt: exec.CreatedAt,
					UpdatedAt: exec.UpdatedAt,
					CreatedBy: exec.CreatedBy,
					UpdatedBy: exec.UpdatedBy,
				})
			}
		}
	}
	c.JSON(http.StatusOK, reply)
}

func (h *InstanceHandler) Delete(c *gin.Context) {
	// 处理删除实例的请求
	req := new(DeleteInstanceReq)
	if err := c.ShouldBindUri(req); err != nil {
		writeBindError(c, h.logger, "failed to bind DeleteInstanceReq", err)
		return
	}
	if err := h.svc.DeleteInstance(c.Request.Context(), req.ID); err != nil {
		writeInternalError(c, h.logger, "failed to delete instance", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted success"})
}

type CreateInstanceReq struct {
	// 流程标识 (必填)
	// 前端只传 Code，后端负责查 Definition 表找最新版
	ProcessCode string `json:"process_code" binding:"required"`

	// 业务表单数据 (可选)
	FormData []biz.FormDataItem `json:"form_data"`

	// 父流程相关 (可选，用于子流程场景)
	ParentID     int64  `json:"parent_id"`
	ParentNodeID string `json:"parent_node_id"`
}

func (r *CreateInstanceReq) ToBizParams() *biz.StartProcessInstanceParams {
	return &biz.StartProcessInstanceParams{
		ProcessCode:  r.ProcessCode,
		FormData:     r.FormData,
		ParentID:     r.ParentID,
		ParentNodeID: r.ParentNodeID,
	}
}

type ListInstancesReq struct {
	Page int `form:"page" binding:"omitempty,min=1"`
	Size int `form:"size" binding:"omitempty,min=1,max=100"`
}

func (r *ListInstancesReq) ToBizParams() *biz.ListInstancesParams {
	return &biz.ListInstancesParams{
		Page: r.Page,
		Size: r.Size,
	}
}

type GetInstanceDetailReq struct {
	ID int64 `uri:"id" binding:"required"`
}

type DeleteInstanceReq struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}
