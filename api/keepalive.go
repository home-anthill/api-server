package api

import (
	"api-server/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

// KeepAlive struct
type KeepAlive struct {
	ctx    context.Context
	logger *zap.SugaredLogger
}

// NewKeepAlive function
func NewKeepAlive(ctx context.Context, logger *zap.SugaredLogger) *KeepAlive {
	return &KeepAlive{
		ctx:    ctx,
		logger: logger,
	}
}

// GetKeepAlive function
func (handler *KeepAlive) GetKeepAlive(c *gin.Context) {
	response := models.KeepAlive{}
	response.Message = "ok"
	c.JSON(http.StatusOK, &response)
}
