package main

import (
	amqpPublisher "api-devices/amqp-publisher"
	pbd "api-devices/device"
	"api-devices/handlers"
	mqttClient "api-devices/mqtt-client"
	pbr "api-devices/register"
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

//var devicesHandler *handlers.DevicesHandler
var registerGrpcHandler *handlers.RegisterGrpcHandler
var devicesGrpcHandler *handlers.DevicesGrpcHandler

const DbName = "api-devices"

func init() {
	//Load the .env file
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("error: failed to load the env file")
	}

	ctx := context.Background()
	log.Println("Connecting to MongoDB...")
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017/"))
	if err = client.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to MongoDB")
	collectionACs := client.Database(DbName).Collection("airconditioners")

	//devicesHandler = handlers.NewDevicesHandler(ctx, collectionHomes)
	registerGrpcHandler = handlers.NewRegisterGrpcHandler(ctx, collectionACs)
	devicesGrpcHandler = handlers.NewDevicesGrpcHandler(ctx, collectionACs)
}

func main() {
	amqpPublisher.InitAmqpPublisher()
	mqttClient.InitMqtt()

	port := os.Getenv("GRPC_PORT")

	// Start listener, 50051 is the default gRPC port
	lis, errGrpc := net.Listen("tcp", ":" + port)
	if errGrpc != nil {
		log.Fatalf("failed to listen: %v", errGrpc)
	}
	// Create new gRPC server with (blank) options
	s := grpc.NewServer()
	// Register the service with the server
	pbr.RegisterRegistrationServer(s, registerGrpcHandler)
	pbd.RegisterDeviceServer(s, devicesGrpcHandler)

	log.Printf("server listening at %v", lis.Addr())
	if errGrpc := s.Serve(lis); errGrpc != nil {
		log.Fatalf("failed to serve: %v", errGrpc)
	}
}
