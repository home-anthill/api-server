package api

import (
	"api-server/models"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"net/http"
)

type KeepAlive struct {
	ctx    context.Context
	logger *zap.SugaredLogger
}

func NewKeepAlive(ctx context.Context, logger *zap.SugaredLogger) *KeepAlive {
	return &KeepAlive{
		ctx:    ctx,
		logger: logger,
	}
}
func (handler *KeepAlive) GetKeepAlive(c *gin.Context) {
	handler.logger.Info("REST - GET - GetKeepAlive called")
	response := models.KeepAlive{}
	response.Message = "ok"
	c.JSON(http.StatusOK, &response)
}
