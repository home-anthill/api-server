package main

import (
  "api-devices/api"
  pbd "api-devices/api/device"
  pbr "api-devices/api/register"
  mqttClient "api-devices/mqtt-client"
  "context"
  "fmt"
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
  var envFile string
  if os.Getenv("ENV") == "prod" {
    envFile = ".env_prod"
  } else {
    envFile = ".env"
  }
  err := godotenv.Load(envFile)
  if err != nil {
    logger.Error("failed to load the env file")
  }

  // 3. Read ENV property from .env
  grpcUrl := os.Getenv("GRPC_URL")

  // 4. Print .env vars
  fmt.Println("ENVIRONMENT = " + os.Getenv("ENV"))
  fmt.Println("MONGODB_URL = " + os.Getenv("MONGODB_URL"))
  fmt.Println("RABBITMQ_URL = " + os.Getenv("RABBITMQ_URL"))
  fmt.Println("MQTT_URL = " + os.Getenv("MQTT_URL"))
  fmt.Println("GRPC_URL = " + os.Getenv("GRPC_URL"))

  logger.Info("ENVIRONMENT = " + os.Getenv("ENV"))
  logger.Info("MONGODB_URL = " + os.Getenv("MONGODB_URL"))
  logger.Info("RABBITMQ_URL = " + os.Getenv("RABBITMQ_URL"))
  logger.Info("MQTT_URL = " + os.Getenv("MQTT_URL"))
  logger.Info("GRPC_URL = " + os.Getenv("GRPC_URL"))

  ctx := context.Background()

  // 5. Connect to DB
  logger.Info("Connecting to MongoDB...")
  mongoDBUrl := os.Getenv("MONGODB_URL")
  // mongoDBUrl := "mongodb+srv://ks89:XuF3Zw2omd9cUy7b6A4oVg@cluster0.4wies.mongodb.net"
  client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoDBUrl))
  if err = client.Ping(context.TODO(), readpref.Primary()); err != nil {
    logger.Fatalf("Cannot connect to MongoDB: %s", err)
  }
  logger.Info("Connected to MongoDB")
  // 6. Define DB collections
  collectionACs := client.Database(DbName).Collection("airconditioners")

  // 7. Create gRPC API instances
  registerGrpc = api.NewRegisterGrpc(ctx, logger, collectionACs)
  devicesGrpc = api.NewDevicesGrpc(ctx, logger, collectionACs)

  // 8. Init AMQP and open connection
  // amqpPublisher.InitAmqpPublisher()

  // 9. Init MQTT and start it
  mqttClient.InitMqtt()
  fmt.Println("MQTT initialized")

  // 10. Start gRPC listener
  lis, errGrpc := net.Listen("tcp", grpcUrl)
  if errGrpc != nil {
    logger.Fatalf("failed to listen: %v", errGrpc)
  }
  // Create new gRPC server with (blank) options
  s := grpc.NewServer()
  // Register the service with the server
  pbr.RegisterRegistrationServer(s, registerGrpc)
  pbd.RegisterDeviceServer(s, devicesGrpc)
  fmt.Println("gRPC client listening at " + lis.Addr().String())
  logger.Infof("server listening at %v", lis.Addr())
  if errGrpc := s.Serve(lis); errGrpc != nil {
    logger.Fatalf("failed to serve: %v", errGrpc)
  }
}
