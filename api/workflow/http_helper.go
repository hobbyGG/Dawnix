package workflow

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func writeBindError(c *gin.Context, logger *zap.Logger, message string, err error) {
	logger.Error(message, zap.Error(err))
	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
}

func writeInternalError(c *gin.Context, logger *zap.Logger, message string, err error) {
	logger.Error(message, zap.Error(err))
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}

func writeUnauthorized(c *gin.Context) {
	c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
}
