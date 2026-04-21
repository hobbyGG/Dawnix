package workflow

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	authService "github.com/hobbyGG/Dawnix/internal/auth/service"
	"github.com/hobbyGG/Dawnix/internal/workflow/domain"
)

type EnumHandler struct {
	emailServiceEnabled bool
	userLookup          UserLookupService
}

type UserLookupService interface {
	SearchActiveUsers(ctx context.Context, keyword string, limit int) ([]*authService.UserOption, error)
}

func NewEnumHandler(emailServiceEnabled bool, userLookup UserLookupService) *EnumHandler {
	return &EnumHandler{
		emailServiceEnabled: emailServiceEnabled,
		userLookup:          userLookup,
	}
}

func (h *EnumHandler) Register(rg *gin.RouterGroup) {
	r := rg.Group("enum")
	r.GET("node-types", h.NodeTypes)
	r.GET("form-types", h.FormTypes)
	r.GET("approvers", h.Approvers)
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

func (h *EnumHandler) Approvers(c *gin.Context) {
	if h.userLookup == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "approver search service is not configured"})
		return
	}

	keyword := c.Query("keyword")
	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil || parsedLimit <= 0 || parsedLimit > 100 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be an integer between 1 and 100"})
			return
		}
		limit = parsedLimit
	}

	users, err := h.userLookup.SearchActiveUsers(c.Request.Context(), keyword, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	items := make([]EnumItem, 0, len(users))
	for _, user := range users {
		label := user.DisplayName
		if label == "" {
			label = user.UserID
		}
		items = append(items, EnumItem{
			Label: label,
			Value: user.UserID,
		})
	}

	c.JSON(http.StatusOK, gin.H{"list": items})
}

type EnumItem struct {
	Label string `json:"label"`
	Value string `json:"value"`
}
