package handlers

import (
  pb "api-devices/device"
  "api-devices/models"
  "context"
  "fmt"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/mongo"
  "time"
)

type DevicesGrpcHandler struct {
  pb.UnimplementedDeviceServer
  airConditionerCollection *mongo.Collection
  contextRef               context.Context
}

func NewDevicesGrpcHandler(ctx context.Context, collection *mongo.Collection) *DevicesGrpcHandler {
  return &DevicesGrpcHandler{
    airConditionerCollection: collection,
    contextRef:               ctx,
  }
}

func (handler *DevicesGrpcHandler) GetStatus(ctx context.Context, in *pb.StatusRequest) (*pb.StatusResponse, error) {
  fmt.Println("Received: ", in)
  var ac models.AirConditioner
  err := handler.airConditionerCollection.FindOne(handler.contextRef, bson.M{
    "mac": in.Mac,
  }).Decode(&ac)
  if err != nil {
    fmt.Println("Cannot get AC with specified mac ", in.Mac)
  }
  return &pb.StatusResponse{
    On:          ac.Status.On,
    Temperature: int32(ac.Status.Temperature),
    Mode:        int32(ac.Status.Mode),
    FanMode:     int32(ac.Status.Fan.Mode),
    FanSpeed:    int32(ac.Status.Fan.Speed),
  }, err
}
func (handler *DevicesGrpcHandler) SetOnOff(ctx context.Context, in *pb.OnOffValueRequest) (*pb.OnOffValueResponse, error) {
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
  return &pb.OnOffValueResponse{Status: "200", Message: "Updated"}, err
}
func (handler *DevicesGrpcHandler) SetTemperature(ctx context.Context, in *pb.TemperatureValueRequest) (*pb.TemperatureValueResponse, error) {
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
  return &pb.TemperatureValueResponse{Status: "200", Message: "Updated"}, err
}
func (handler *DevicesGrpcHandler) SetMode(ctx context.Context, in *pb.ModeValueRequest) (*pb.ModeValueResponse, error) {
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
  return &pb.ModeValueResponse{Status: "200", Message: "Updated"}, err
}
func (handler *DevicesGrpcHandler) SetFanMode(ctx context.Context, in *pb.FanModeValueRequest) (*pb.FanModeValueResponse, error) {
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
  return &pb.FanModeValueResponse{Status: "200", Message: "Updated"}, err
}
func (handler *DevicesGrpcHandler) SetFanSpeed(ctx context.Context, in *pb.FanSpeedValueRequest) (*pb.FanSpeedValueResponse, error) {
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
  return &pb.FanSpeedValueResponse{Status: "200", Message: "Updated"}, err
}
