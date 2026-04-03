package api

import (
	"api-server/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// KeepAlive handles health-check requests.
type KeepAlive struct {
	logger *zap.SugaredLogger
}

// NewKeepAlive constructs a KeepAlive handler.
func NewKeepAlive(logger *zap.SugaredLogger) *KeepAlive {
	return &KeepAlive{
		logger: logger,
	}
}

// GetKeepAlive function
func (ka *KeepAlive) GetKeepAlive(c *gin.Context) {
	response := models.KeepAlive{}
	response.Message = "ok"
	c.JSON(http.StatusOK, &response)
}
