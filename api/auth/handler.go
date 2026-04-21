package auth

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	authService "github.com/hobbyGG/Dawnix/internal/auth/service"
	"go.uber.org/zap"
)

type Handler struct {
	svc    *authService.Service
	logger *zap.Logger
}

func NewHandler(svc *authService.Service, logger *zap.Logger) *Handler {
	return &Handler{svc: svc, logger: logger}
}

func (h *Handler) Register(rg *gin.RouterGroup) {
	r := rg.Group("auth")
	r.POST("signup", h.Signup)
	r.POST("signin", h.Signin)
	r.POST("logout", h.Logout)
}

type RegisterReq struct {
	Username    string `json:"username" binding:"required"`
	Password    string `json:"password" binding:"required"`
	DisplayName string `json:"display_name"`
}

type LoginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// 注册
func (h *Handler) Signup(c *gin.Context) {
	var req RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBindError(c, h.logger, "failed to bind signup request", err)
		return
	}

	result, err := h.svc.Register(c.Request.Context(), &authService.RegisterParams{
		Username:    req.Username,
		Password:    req.Password,
		DisplayName: req.DisplayName,
	})
	if err != nil {
		if errors.Is(err, authService.ErrUsernameAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		writeInternalError(c, h.logger, "signup failed", err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// 登录
func (h *Handler) Signin(c *gin.Context) {
	var req LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBindError(c, h.logger, "failed to bind signin request", err)
		return
	}

	result, err := h.svc.Login(c.Request.Context(), &authService.LoginParams{Username: req.Username, Password: req.Password})
	if err != nil {
		if errors.Is(err, authService.ErrInvalidCredentials) {
			writeUnauthorizedError(c, authService.ErrInvalidCredentials.Error())
			return
		}
		writeInternalError(c, h.logger, "signin failed", err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) Logout(c *gin.Context) {
	// JWT is stateless in current implementation, so logout is a client-side token discard.
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}
