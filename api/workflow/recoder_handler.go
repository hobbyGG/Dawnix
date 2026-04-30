package workflow

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	//"github.com/hobbyGG/Dawnix/internal/workflow/biz"
	"github.com/hobbyGG/Dawnix/internal/workflow/service"
	"go.uber.org/zap"
)

type RecordHandler struct {
	svc    *service.RecordService
	logger *zap.Logger
}

func NewRecordHandler(svc *service.RecordService, logger *zap.Logger) *RecordHandler {
	return &RecordHandler{
		svc:    svc,
		logger: logger,
	}
}

func (h *RecordHandler) Register(rg *gin.RouterGroup) {
	r := rg.Group("record")
	r.GET("create", h.List)
}

func (h *RecordHandler) List(c *gin.Context) {
	// 支持可选的 instance_id：若未提供则返回全部记录
	instanceIDStr := c.Query("instance_id")
	if instanceIDStr == "" {
		list, err := h.svc.ListAll(c.Request.Context())
		if err != nil {
			writeInternalError(c, h.logger, "failed to list approval records", err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "data": list, "msg": "success"})
		return
	}

	// 转换格式并按 instance_id 查询
	instanceId, err := strconv.ParseInt(instanceIDStr, 10, 64)
	if err != nil {
		h.logger.Error("instance_id must be a valid int64", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "msg": "instance_id must be a valid int64"})
		return
	}

	list, err := h.svc.ListByInstanceID(c.Request.Context(), instanceId)
	if err != nil {
		writeInternalError(c, h.logger, "failed to list approval records", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "data": list, "msg": "success"})

	// req := new(ListInstancesReq)
	// if err := c.ShouldBindQuery(req); err != nil {
	// 	writeBindError(c, h.logger, "failed to bind ListInstancesReq", err)
	// 	return
	// }
	// instances, err := h.svc.ListInstances(c.Request.Context(), req.ToBizParams())
	// if err != nil {
	// 	writeInternalError(c, h.logger, "failed to list instances", err)
	// 	return
	// }
	// c.JSON(http.StatusOK, instances)

}
