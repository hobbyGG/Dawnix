package service

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const currentUserIDGinKey = "current_user_id"

func JWTMiddleware(authService *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}
		if c.Request.URL.Path == "/api/v1/auth/signin" || c.Request.URL.Path == "/api/v1/auth/signup" {
			c.Next()
			return
		}

		authorization := c.GetHeader("Authorization")
		if !strings.HasPrefix(authorization, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}
		token := strings.TrimSpace(strings.TrimPrefix(authorization, "Bearer "))
		claims, err := authService.ParseToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		c.Set(currentUserIDGinKey, claims.UserID)
		ctx := WithUserID(c.Request.Context(), claims.UserID)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func UserIDFromGin(c *gin.Context) (string, bool) {
	value, ok := c.Get(currentUserIDGinKey)
	if !ok {
		return "", false
	}
	userID, ok := value.(string)
	if !ok || userID == "" {
		return "", false
	}
	return userID, true
}
