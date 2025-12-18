package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hobbyGG/Dawnix/internal/service"
	"go.uber.org/zap"
)

// 对应controller

type HelloHandler struct {
	svc    *service.HelloService
	logger *zap.Logger
}

func NewHelloHandler(svc *service.HelloService, logger *zap.Logger) *HelloHandler {
	return &HelloHandler{
		svc:    svc,
		logger: logger,
	}
}

func (h *HelloHandler) Register(rg *gin.RouterGroup) {
	rg.POST("/hello", h.Hello)
}

func (h *HelloHandler) Hello(c *gin.Context) {
	var helloReq struct{}
	c.ShouldBind(&helloReq)
	if err := h.svc.Hello(c, helloReq); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"err": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "hello!",
	})
}
