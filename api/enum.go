package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hobbyGG/Dawnix/internal/domain"
)

type EnumHandler struct {
	emailServiceEnabled bool
}

func NewEnumHandler(emailServiceEnabled bool) *EnumHandler {
	return &EnumHandler{emailServiceEnabled: emailServiceEnabled}
}

func (h *EnumHandler) Register(rg *gin.RouterGroup) {
	r := rg.Group("enum")
	r.GET("node-types", h.NodeTypes)
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

type EnumItem struct {
	Label string `json:"label"`
	Value string `json:"value"`
}
