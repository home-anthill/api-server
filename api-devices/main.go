package main

import (
	"air-conditioner/handlers"
	mqttClient "air-conditioner/mqtt-client"
	"context"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
)

var devicesHandler *handlers.DevicesHandler

func init() {
	ctx := context.Background()
	log.Println("Connecting to MongoDB...")
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017/"))
	if err = client.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to MongoDB")
	collectionHomes := client.Database("airConditionerDb").Collection("homes")

	devicesHandler = handlers.NewDevicesHandler(ctx, collectionHomes)
}

func main() {
	mqttClient.InitMqtt()

	router := gin.Default()

	router.POST("/devices/onoff", devicesHandler.PostOnOffDeviceHandler)
	router.POST("/devices/temperature", devicesHandler.PostTemperatureDeviceHandler)
	router.POST("/devices/mode", devicesHandler.PostModeDeviceHandler)
	router.POST("/devices/fan", devicesHandler.PostFanDeviceHandler)
	router.POST("/devices/swing", devicesHandler.PostSwingDeviceHandler)

	err := router.Run(":8081")
	if err != nil {
		panic(err)
	}
}
