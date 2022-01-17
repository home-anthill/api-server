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
	"net"
	"os"
)

const DbName = "api-devices"

var registerGrpc *api.RegisterGrpc
var devicesGrpc *api.DevicesGrpc

func main() {
	// 1. Init logger
	logger := InitLogger()
	defer logger.Sync()
	logger.Info("Starting application...)")

	// 2. Load the .env file
	err := godotenv.Load(".env")
	if err != nil {
		logger.Error("failed to load the env file")
	}

	// 3. Read ENV property from .env
	port := os.Getenv("GRPC_PORT")

	// 4. Print .env vars
	logger.Info("GRPC PORT = " + port)

	ctx := context.Background()

	// 5. Connect to DB
	logger.Info("Connecting to MongoDB...")
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017/"))
	if err = client.Ping(context.TODO(), readpref.Primary()); err != nil {
		logger.Fatalf("Cannot connect to MongoDB: %s", err)
		return
	}
	logger.Info("Connected to MongoDB")
	// 6. Define DB collections
	collectionACs := client.Database(DbName).Collection("airconditioners")

	// 7. Create gRPC API instances
	registerGrpc = api.NewRegisterGrpc(ctx, logger, collectionACs)
	devicesGrpc = api.NewDevicesGrpc(ctx, logger, collectionACs)

	// 8. Init AMQP and open connection
	amqpPublisher.InitAmqpPublisher()

	// 9. Init MQTT and start it
	mqttClient.InitMqtt()

	// 10. Start gRPC listener
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
