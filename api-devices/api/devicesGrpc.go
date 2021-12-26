package api

import (
	"api-devices/api/device"
	"api-devices/models"
	mqtt_client "api-devices/mqtt-client"
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"time"
)

const TIMEOUT = 3 * time.Second

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
	handler.logger.Info("gRPC - SetOnOff - method called")
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
		fmt.Println("gRPC - SetOnOff - Cannot update db with the registered AC with mac ", in.Mac)
		return nil, err
	}

	onOffValue := models.OnOffValue{
		Uuid:         in.Uuid,
		ProfileToken: in.ProfileToken,
		On:           in.On,
	}
	messageJSON, err := json.Marshal(onOffValue)
	if err != nil {
		fmt.Println("gRPC - SetOnOff - cannot create mqtt payload")
		return nil, err
	}
	t := mqtt_client.SendOnOff(onOffValue.Uuid, messageJSON)
	timeoutResult := t.WaitTimeout(TIMEOUT)
	if t.Error() != nil || !timeoutResult {
		fmt.Println(t.Error())
		fmt.Println("gRPC - SetOnOff - cannot send data via mqtt")
		return nil, t.Error()
	} else {
		fmt.Println("sending response")
		return &device.OnOffValueResponse{Status: "200", Message: "Updated"}, err
	}
}
func (handler *DevicesGrpc) SetTemperature(ctx context.Context, in *device.TemperatureValueRequest) (*device.TemperatureValueResponse, error) {
	handler.logger.Info("gRPC - SetTemperature - method called")
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
		fmt.Println("gRPC - SetTemperature - Cannot update db with the registered AC with mac ", in.Mac)
		return nil, err
	}

	temperatureValue := models.TemperatureValue{
		Uuid:         in.Uuid,
		ProfileToken: in.ProfileToken,
		Temperature:  int(in.Temperature),
	}
	messageJSON, err := json.Marshal(temperatureValue)
	if err != nil {
		fmt.Println("gRPC - SetTemperature - cannot create mqtt payload")
		return nil, err
	}
	t := mqtt_client.SendTemperature(temperatureValue.Uuid, messageJSON)
	timeoutResult := t.WaitTimeout(TIMEOUT)
	if t.Error() != nil || !timeoutResult {
		fmt.Println(t.Error())
		fmt.Println("gRPC - SetTemperature - cannot send data via mqtt")
		return nil, t.Error()
	} else {
		fmt.Println("sending response")
		return &device.TemperatureValueResponse{Status: "200", Message: "Updated"}, err
	}
}
func (handler *DevicesGrpc) SetMode(ctx context.Context, in *device.ModeValueRequest) (*device.ModeValueResponse, error) {
	handler.logger.Info("gRPC - SetMode - method called")
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
		fmt.Println("gRPC - SetMode - Cannot update db with the registered AC with mac ", in.Mac)
		return nil, err
	}

	modeValue := models.ModeValue{
		Uuid:         in.Uuid,
		ProfileToken: in.ProfileToken,
		Mode:         int(in.Mode),
	}
	messageJSON, err := json.Marshal(modeValue)
	if err != nil {
		fmt.Println("gRPC - SetMode - cannot create mqtt payload")
		return nil, err
	}
	t := mqtt_client.SendMode(modeValue.Uuid, messageJSON)
	timeoutResult := t.WaitTimeout(TIMEOUT)
	if t.Error() != nil || !timeoutResult {
		fmt.Println(t.Error())
		fmt.Println("gRPC - SetMode - cannot send data via mqtt")
		return nil, t.Error()
	} else {
		fmt.Println("sending response")
		return &device.ModeValueResponse{Status: "200", Message: "Updated"}, err
	}
}
func (handler *DevicesGrpc) SetFanMode(ctx context.Context, in *device.FanModeValueRequest) (*device.FanModeValueResponse, error) {
	handler.logger.Info("gRPC - SetFanMode - method called")
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
		fmt.Println("gRPC - SetFanMode - Cannot update db with the registered AC with mac ", in.Mac)
		return nil, err
	}

	fanModeValue := models.FanModeValue{
		Uuid:         in.Uuid,
		ProfileToken: in.ProfileToken,
		FanMode:      int(in.FanMode),
	}
	messageJSON, err := json.Marshal(fanModeValue)
	if err != nil {
		fmt.Println("gRPC - SetFanMode - cannot create mqtt payload")
		return nil, err
	}
	t := mqtt_client.SendFanMode(fanModeValue.Uuid, messageJSON)
	timeoutResult := t.WaitTimeout(TIMEOUT)
	if t.Error() != nil || !timeoutResult {
		fmt.Println(t.Error())
		fmt.Println("gRPC - SetFanMode - cannot send data via mqtt")
		return nil, t.Error()
	} else {
		fmt.Println("sending response")
		return &device.FanModeValueResponse{Status: "200", Message: "Updated"}, err
	}
}
func (handler *DevicesGrpc) SetFanSpeed(ctx context.Context, in *device.FanSpeedValueRequest) (*device.FanSpeedValueResponse, error) {
	handler.logger.Info("gRPC - SetFanSpeed - method called")
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
		fmt.Println("gRPC - SetFanSpeed - Cannot update db with the registered AC with mac ", in.Mac)
		return nil, err
	}

	fanSpeedValue := models.FanSpeedValue{
		Uuid:         in.Uuid,
		ProfileToken: in.ProfileToken,
		FanSpeed:     int(in.FanSpeed),
	}
	messageJSON, err := json.Marshal(fanSpeedValue)
	if err != nil {
		fmt.Println("gRPC - SetFanSpeed - cannot create mqtt payload")
		return nil, err
	}
	t := mqtt_client.SendFanSpeed(fanSpeedValue.Uuid, messageJSON)
	timeoutResult := t.WaitTimeout(TIMEOUT)
	if t.Error() != nil || !timeoutResult {
		fmt.Println(t.Error())
		fmt.Println("gRPC - SetFanSpeed - cannot send data via mqtt")
		return nil, t.Error()
	} else {
		fmt.Println("sending response")
		return &device.FanSpeedValueResponse{Status: "200", Message: "Updated"}, err
	}
}
