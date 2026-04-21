package workflow

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hobbyGG/Dawnix/api/workflow/middleware"
	"github.com/hobbyGG/Dawnix/internal/workflow/domain"
)

type EnumHandler struct {
	emailServiceEnabled bool
}

func NewEnumHandler(emailServiceEnabled bool) *EnumHandler {
	return &EnumHandler{emailServiceEnabled: emailServiceEnabled}
}

func (h *EnumHandler) Register(rg *gin.RouterGroup) {
	r := rg.Group("enum", middleware.InjectUID())
	r.GET("node-types", h.NodeTypes)
	r.GET("form-types", h.FormTypes)
}

func (h *EnumHandler) NodeTypes(c *gin.Context) {
	items := []EnumItem{
		{Label: "开始节点", Value: domain.NodeTypeStart},
		{Label: "结束节点", Value: domain.NodeTypeEnd},
		{Label: "用户任务", Value: domain.NodeTypeUserTask},
		{Label: "并行分支网关", Value: domain.NodeTypeForkGateway},
		{Label: "并行汇聚网关", Value: domain.NodeTypeJoinGateway},
		{Label: "排他网关", Value: domain.NodeTypeXORGateway},
		{Label: "包含网关", Value: domain.NodeTypeInclusiveGateway},
	}
	if h.emailServiceEnabled {
		items = append(items, EnumItem{Label: "邮件服务节点", Value: domain.NodeTypeEmailService})
	}
	c.JSON(http.StatusOK, gin.H{"list": items})
}

func (h *EnumHandler) FormTypes(c *gin.Context) {
	items := []EnumItem{
		{Label: "单行文本", Value: domain.FormTypeTextSingleLine},
		{Label: "数字", Value: domain.FormTypeNumber},
		{Label: "单选/下拉", Value: domain.FormTypeSingleSelect},
		{Label: "日期", Value: domain.FormTypeDate},
	}
	c.JSON(http.StatusOK, gin.H{"list": items})
}

type EnumItem struct {
	Label string `json:"label"`
	Value string `json:"value"`
}
