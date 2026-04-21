package service

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func JWTMiddleware(authService *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		if shouldSkipJWT(c) {
			c.Next()
			return
		}

		authorization := c.GetHeader("Authorization")
		if !strings.HasPrefix(authorization, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}
		token := strings.TrimPrefix(authorization, "Bearer ")
		claims, err := authService.ParseToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		// 将 uid 注入到 request context 和 gin context 中
		ctx := WithUserID(c.Request.Context(), claims.UserID)
		c.Request = c.Request.WithContext(ctx)
		// 同时注入到 Gin Context 中作为备份，方便直接通过 c.Get() 获取
		c.Set("uid", claims.UserID)
		c.Next()
	}
}

func shouldSkipJWT(c *gin.Context) bool {
	if c.Request.Method == http.MethodOptions {
		return true
	}
	fullPath := c.FullPath()
	if fullPath == "" {
		fullPath = c.Request.URL.Path
	}
	return strings.HasPrefix(fullPath, "/api/v1/auth/")
}
