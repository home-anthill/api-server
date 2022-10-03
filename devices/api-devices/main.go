package main

import (
  "api-devices/api"
  pbd "api-devices/api/device"
  pbk "api-devices/api/keepalive"
  pbr "api-devices/api/register"
  mqttClient "api-devices/mqtt-client"
  "context"
  "fmt"
  "github.com/joho/godotenv"
  "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/mongo/options"
  "go.mongodb.org/mongo-driver/mongo/readpref"
  "google.golang.org/grpc"
  "google.golang.org/grpc/credentials"
  "net"
  "os"
)

const DbName = "api-devices"

var registerGrpc *api.RegisterGrpc
var devicesGrpc *api.DevicesGrpc
var keepAliveGrpc *api.KeepAliveGrpc

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
  fmt.Println("MQTT_URL = " + os.Getenv("MQTT_URL"))
  fmt.Println("MQTT_PORT = " + os.Getenv("MQTT_PORT"))
  fmt.Println("MQTT_TLS = " + os.Getenv("MQTT_TLS"))
  fmt.Println("MQTT_CA_FILE = " + os.Getenv("MQTT_CA_FILE"))
  fmt.Println("MQTT_CERT_FILE = " + os.Getenv("MQTT_CERT_FILE"))
  fmt.Println("MQTT_KEY_FILE = " + os.Getenv("MQTT_KEY_FILE"))
  fmt.Println("GRPC_URL = " + os.Getenv("GRPC_URL"))
  fmt.Println("GRPC_TLS = " + os.Getenv("GRPC_TLS"))
  fmt.Println("CERT_FOLDER_PATH = " + os.Getenv("CERT_FOLDER_PATH"))

  logger.Info("ENVIRONMENT = " + os.Getenv("ENV"))
  logger.Info("MONGODB_URL = " + os.Getenv("MONGODB_URL"))
  logger.Info("MQTT_URL = " + os.Getenv("MQTT_URL"))
  logger.Info("MQTT_PORT = " + os.Getenv("MQTT_PORT"))
  logger.Info("MQTT_TLS = " + os.Getenv("MQTT_TLS"))
  logger.Info("MQTT_CA_FILE = " + os.Getenv("MQTT_CA_FILE"))
  logger.Info("MQTT_CERT_FILE = " + os.Getenv("MQTT_CERT_FILE"))
  logger.Info("MQTT_KEY_FILE = " + os.Getenv("MQTT_KEY_FILE"))
  logger.Info("GRPC_URL = " + os.Getenv("GRPC_URL"))
  logger.Info("GRPC_TLS = " + os.Getenv("GRPC_TLS"))
  logger.Info("CERT_FOLDER_PATH = " + os.Getenv("CERT_FOLDER_PATH"))

  ctx := context.Background()

  // 5. Connect to DB
  logger.Info("Connecting to MongoDB...")
  mongoDBUrl := os.Getenv("MONGODB_URL")
  client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoDBUrl))
  if os.Getenv("ENV") != "prod" {
    if err = client.Ping(context.TODO(), readpref.Primary()); err != nil {
      logger.Fatalf("Cannot connect to MongoDB: %s", err)
    }
  }
  logger.Info("Connected to MongoDB")
  // 6. Define DB collections
  collectionACs := client.Database(DbName).Collection("airconditioners")

  // 7. Create gRPC API instances
  registerGrpc = api.NewRegisterGrpc(ctx, logger, collectionACs)
  devicesGrpc = api.NewDevicesGrpc(ctx, logger, collectionACs)
  keepAliveGrpc = api.NewKeepAliveGrpc(ctx, logger)

  // 8. Init MQTT and start it
  mqttClient.InitMqtt()
  fmt.Println("MQTT initialized")

  // 9. Start gRPC listener
  // Create new gRPC server with (blank) options
  var server *grpc.Server
  if os.Getenv("GRPC_TLS") == "true" {
    creds, credErr := credentials.NewServerTLSFromFile(
      os.Getenv("CERT_FOLDER_PATH")+"/server-cert.pem",
      os.Getenv("CERT_FOLDER_PATH")+"/server-key.pem",
    )
    if credErr != nil {
      logger.Fatalf("NewServerTLSFromFile error %v", credErr)
    }
    logger.Info("gRPC TLS security enabled")
    server = grpc.NewServer(grpc.Creds(creds))
  } else {
    logger.Info("gRPC TLS security not enabled")
    server = grpc.NewServer()
  }
  // 10. Register the service with the server
  pbr.RegisterRegistrationServer(server, registerGrpc)
  pbd.RegisterDeviceServer(server, devicesGrpc)
  pbk.RegisterKeepAliveServer(server, keepAliveGrpc)
  lis, errGrpc := net.Listen("tcp", grpcUrl)
  if errGrpc != nil {
    logger.Fatalf("failed to listen: %v", errGrpc)
  }
  fmt.Println("gRPC client listening at " + lis.Addr().String())
  logger.Infof("server listening at %v", lis.Addr())
  if errGrpc := server.Serve(lis); errGrpc != nil {
    logger.Fatalf("failed to serve: %v", errGrpc)
  }
}
