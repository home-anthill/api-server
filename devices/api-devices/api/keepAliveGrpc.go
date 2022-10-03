package api

import (
  "api-devices/api/keepalive"
  "context"
  "go.uber.org/zap"
)

type KeepAliveGrpc struct {
  keepalive.UnimplementedKeepAliveServer
  ctx    context.Context
  logger *zap.SugaredLogger
}

func NewKeepAliveGrpc(ctx context.Context, logger *zap.SugaredLogger) *KeepAliveGrpc {
  return &KeepAliveGrpc{
    ctx:    ctx,
    logger: logger,
  }
}

func (handler *KeepAliveGrpc) GetKeepAlive(ctx context.Context, in *keepalive.StatusRequest) (*keepalive.StatusResponse, error) {
  handler.logger.Info("gRPC - GetKeepAlive - Called")
  return &keepalive.StatusResponse{}, nil
}
