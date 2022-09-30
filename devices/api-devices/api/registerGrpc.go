package api

import (
  "api-devices/api/register"
  "context"
  "fmt"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/mongo/options"
  "go.uber.org/zap"
  "time"
)

type RegisterGrpc struct {
  register.UnimplementedRegistrationServer
  airConditionerCollection *mongo.Collection
  ctx                      context.Context
  logger                   *zap.SugaredLogger
}

func NewRegisterGrpc(ctx context.Context, logger *zap.SugaredLogger, collection *mongo.Collection) *RegisterGrpc {
  return &RegisterGrpc{
    airConditionerCollection: collection,
    ctx:                      ctx,
    logger:                   logger,
  }
}

func (handler *RegisterGrpc) Register(ctx context.Context, in *register.RegisterRequest) (*register.RegisterReply, error) {
  handler.logger.Info("gRPC - Register - Called")
  fmt.Println("Received: ", in)

  // update ac
  upsert := true
  opts := options.UpdateOptions{
    Upsert: &upsert,
  }
  _, err := handler.airConditionerCollection.UpdateOne(handler.ctx, bson.M{
    "mac": in.Mac,
  }, bson.M{
    "$set": bson.M{
      "mac":            in.Mac,
      "uuid":           in.Uuid,
      "name":           in.Name,
      "manufacturer":   in.Manufacturer,
      "model":          in.Model,
      "profileOwnerId": in.ProfileOwnerId,
      "apiToken":       in.ApiToken,
      "createdAt":      time.Now(),
      "modifiedAt":     time.Now(),
    },
  }, &opts)

  if err != nil {
    handler.logger.Error("gRPC - Register - Cannot update db with the registered AC with id " + in.Id)
  }

  return &register.RegisterReply{Status: "200", Message: "Inserted"}, err
}
