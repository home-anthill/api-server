package api

import (
	"api-devices/api/device"
	"api-devices/models"
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"time"
)

type DevicesGrpc struct {
	device.UnimplementedDeviceServer
	airConditionerCollection *mongo.Collection
	contextRef               context.Context
	logger                   *zap.SugaredLogger
}

func NewDevicesGrpc(ctx context.Context, logger *zap.SugaredLogger, collection *mongo.Collection) *DevicesGrpc {
	return &DevicesGrpc{
		airConditionerCollection: collection,
		contextRef:               ctx,
		logger:                   logger,
	}
}

func (handler *DevicesGrpc) GetStatus(ctx context.Context, in *device.StatusRequest) (*device.StatusResponse, error) {
	handler.logger.Info("gRPC GetStatus called")
	fmt.Println("Received: ", in)

	var ac models.AirConditioner
	err := handler.airConditionerCollection.FindOne(handler.contextRef, bson.M{
		"mac": in.Mac,
	}).Decode(&ac)
	if err != nil {
		fmt.Println("Cannot get AC with specified mac ", in.Mac)
	}
	return &device.StatusResponse{
		On:          ac.Status.On,
		Temperature: int32(ac.Status.Temperature),
		Mode:        int32(ac.Status.Mode),
		FanMode:     int32(ac.Status.Fan.Mode),
		FanSpeed:    int32(ac.Status.Fan.Speed),
	}, err
}
func (handler *DevicesGrpc) SetOnOff(ctx context.Context, in *device.OnOffValueRequest) (*device.OnOffValueResponse, error) {
	handler.logger.Info("gRPC SetOnOff called")
	fmt.Println("Received: ", in)

	_, err := handler.airConditionerCollection.UpdateOne(handler.contextRef, bson.M{
		"mac": in.Mac,
	}, bson.M{
		"$set": bson.M{
			"status.on":  in.On,
			"modifiedAt": time.Now(),
		},
	})
	if err != nil {
		fmt.Println("Cannot update db with the registered AC with mac ", in.Mac)
	}
	return &device.OnOffValueResponse{Status: "200", Message: "Updated"}, err
}
func (handler *DevicesGrpc) SetTemperature(ctx context.Context, in *device.TemperatureValueRequest) (*device.TemperatureValueResponse, error) {
	handler.logger.Info("gRPC SetTemperature called")
	fmt.Println("Received: ", in)

	_, err := handler.airConditionerCollection.UpdateOne(handler.contextRef, bson.M{
		"mac": in.Mac,
	}, bson.M{
		"$set": bson.M{
			"status.temperature": in.Temperature,
			"modifiedAt":         time.Now(),
		},
	})
	if err != nil {
		fmt.Println("Cannot update db with the registered AC with mac ", in.Mac)
	}
	return &device.TemperatureValueResponse{Status: "200", Message: "Updated"}, err
}
func (handler *DevicesGrpc) SetMode(ctx context.Context, in *device.ModeValueRequest) (*device.ModeValueResponse, error) {
	handler.logger.Info("gRPC SetMode called")
	fmt.Println("Received: ", in)

	_, err := handler.airConditionerCollection.UpdateOne(handler.contextRef, bson.M{
		"mac": in.Mac,
	}, bson.M{
		"$set": bson.M{
			"status.mode": in.Mode,
			"modifiedAt":  time.Now(),
		},
	})
	if err != nil {
		fmt.Println("Cannot update db with the registered AC with mac ", in.Mac)
	}
	return &device.ModeValueResponse{Status: "200", Message: "Updated"}, err
}
func (handler *DevicesGrpc) SetFanMode(ctx context.Context, in *device.FanModeValueRequest) (*device.FanModeValueResponse, error) {
	handler.logger.Info("gRPC SetFanMode called")
	fmt.Println("Received: ", in)

	_, err := handler.airConditionerCollection.UpdateOne(handler.contextRef, bson.M{
		"mac": in.Mac,
	}, bson.M{
		"$set": bson.M{
			"status.fan.mode": in.FanMode,
			"modifiedAt":      time.Now(),
		},
	})
	if err != nil {
		fmt.Println("Cannot update db with the registered AC with mac ", in.Mac)
	}
	return &device.FanModeValueResponse{Status: "200", Message: "Updated"}, err
}
func (handler *DevicesGrpc) SetFanSpeed(ctx context.Context, in *device.FanSpeedValueRequest) (*device.FanSpeedValueResponse, error) {
	handler.logger.Info("gRPC SetFanSpeed called")
	fmt.Println("Received: ", in)

	_, err := handler.airConditionerCollection.UpdateOne(handler.contextRef, bson.M{
		"mac": in.Mac,
	}, bson.M{
		"$set": bson.M{
			"status.fan.speed": in.FanSpeed,
			"modifiedAt":       time.Now(),
		},
	})
	if err != nil {
		fmt.Println("Cannot update db with the registered AC with mac ", in.Mac)
	}
	return &device.FanSpeedValueResponse{Status: "200", Message: "Updated"}, err
}
