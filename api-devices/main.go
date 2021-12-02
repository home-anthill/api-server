package main

import (
	amqpPublisher "api-devices/amqp-publisher"
	"api-devices/api"
	pbd "api-devices/api/device"
	pbr "api-devices/api/register"
	mqttClient "api-devices/mqtt-client"
	"context"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
)

var registerGrpc *api.RegisterGrpc
var devicesGrpc *api.DevicesGrpc

const DbName = "api-devices"

func main() {
	logger := InitLogger()
	defer logger.Sync()

	logger.Info("Starting application...)")

	//Load the .env file
	err := godotenv.Load(".env")
	if err != nil {
		logger.Error("failed to load the env file")
	}

	logger.Info("GRPC PORT = " + os.Getenv("GRPC_PORT"))

	ctx := context.Background()

	logger.Info("Connecting to MongoDB...")
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017/"))
	if err = client.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Fatal(err)
	}
	logger.Info("Connected to MongoDB")

	collectionACs := client.Database(DbName).Collection("airconditioners")

	registerGrpc = api.NewRegisterGrpc(ctx, logger, collectionACs)
	devicesGrpc = api.NewDevicesGrpc(ctx, logger, collectionACs)

	amqpPublisher.InitAmqpPublisher()
	mqttClient.InitMqtt()

	port := os.Getenv("GRPC_PORT")

	// Start listener, 50051 is the default gRPC port
	lis, errGrpc := net.Listen("tcp", ":"+port)
	if errGrpc != nil {
		logger.Fatalf("failed to listen: %v", errGrpc)
	}
	// Create new gRPC server with (blank) options
	s := grpc.NewServer()
	// Register the service with the server
	pbr.RegisterRegistrationServer(s, registerGrpc)
	pbd.RegisterDeviceServer(s, devicesGrpc)

	logger.Infof("server listening at %v", lis.Addr())
	if errGrpc := s.Serve(lis); errGrpc != nil {
		logger.Fatalf("failed to serve: %v", errGrpc)
	}
}
