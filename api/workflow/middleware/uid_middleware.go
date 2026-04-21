package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	authService "github.com/hobbyGG/Dawnix/internal/auth/service"
)

func InjectUID() gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, ok := authService.UserIDFromContext(c.Request.Context())
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		ctx := authService.WithUserID(c.Request.Context(), uid)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
