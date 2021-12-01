package handlers

import (
	pb "api-devices/register"
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"time"
)

type RegisterGrpcHandler struct {
	pb.UnimplementedRegistrationServer
	airConditionerCollection *mongo.Collection
	ctx               context.Context
	logger     *zap.SugaredLogger
}

func NewRegisterGrpcHandler(ctx context.Context, logger *zap.SugaredLogger, collection *mongo.Collection) *RegisterGrpcHandler {
	return &RegisterGrpcHandler{
		airConditionerCollection: collection,
		ctx:               ctx,
		logger:     logger,
	}
}

func (handler *RegisterGrpcHandler) Register(ctx context.Context, in *pb.RegisterRequest) (*pb.RegisterReply, error) {
	handler.logger.Info("gRPC Register called")
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
		fmt.Println("Cannot update db with the registered AC with id " + in.Id)
	}

	return &pb.RegisterReply{Status: "200", Message: "Inserted"}, err
}
