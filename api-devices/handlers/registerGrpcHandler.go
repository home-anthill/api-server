package handlers

import (
	pb "api-devices/register"
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type RegisterGrpcHandler struct {
	pb.UnimplementedRegistrationServer
	airConditionerCollection *mongo.Collection
	contextRef               context.Context
}

func NewRegisterGrpcHandler(ctx context.Context, collection *mongo.Collection) *RegisterGrpcHandler {
	return &RegisterGrpcHandler{
		airConditionerCollection: collection,
		contextRef:               ctx,
	}
}

func (handler *RegisterGrpcHandler) Register(ctx context.Context, in *pb.RegisterRequest) (*pb.RegisterReply, error) {
	fmt.Println("Received: ", in)

	// update ac
	upsert := true
	opts := options.UpdateOptions{
		Upsert: &upsert,
	}
	_, err := handler.airConditionerCollection.UpdateOne(handler.contextRef, bson.M{
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
