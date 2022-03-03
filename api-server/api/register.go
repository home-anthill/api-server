package api

import (
  "api-server/api/gRPC/register"
  "api-server/models"
  "crypto/tls"
  "crypto/x509"
  "fmt"
  "github.com/gin-gonic/gin"
  "github.com/google/uuid"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/bson/primitive"
  "go.mongodb.org/mongo-driver/mongo"
  "go.uber.org/zap"
  "golang.org/x/net/context"
  "google.golang.org/grpc"
  "google.golang.org/grpc/credentials"
  "io/ioutil"
  "net/http"
  "os"
  "time"
)

type DeviceRequest struct {
  //swagger:ignore
  Mac          string `json:"mac"`
  Name         string `json:"name"`
  Manufacturer string `json:"manufacturer"`
  Model        string `json:"model"`
  Type         string `json:"type"`
  APIToken     string `json:"apiToken"`
}

type Register struct {
  collection         *mongo.Collection
  collectionProfiles *mongo.Collection
  ctx                context.Context
  logger             *zap.SugaredLogger
  grpcTarget         string
}

func NewRegister(ctx context.Context, logger *zap.SugaredLogger, collection *mongo.Collection, collectionProfiles *mongo.Collection) *Register {
  grpcUrl := os.Getenv("GRPC_URL")
  return &Register{
    collection:         collection,
    collectionProfiles: collectionProfiles,
    ctx:                ctx,
    logger:             logger,
    grpcTarget:         grpcUrl,
  }
}

func loadTLSCredentials() (credentials.TransportCredentials, error) {
  // Load certificate of the CA who signed server's certificate
  pemServerCA, err := ioutil.ReadFile("cert/ca-cert.pem")
  if err != nil {
    return nil, err
  }

  certPool := x509.NewCertPool()
  if !certPool.AppendCertsFromPEM(pemServerCA) {
    return nil, fmt.Errorf("failed to add server CA's certificate")
  }

  // Create the credentials and return it
  config := &tls.Config{
    RootCAs: certPool,
  }

  return credentials.NewTLS(config), nil
}

func (handler *Register) PostRegister(c *gin.Context) {
  handler.logger.Debug("REST - POST - PostRegister called")

  // receive a payload from devices with
  var registerBody DeviceRequest
  if err := c.ShouldBindJSON(&registerBody); err != nil {
    handler.logger.Error("REST - POST - PostRegister - Cannot bind request body", err)
    c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
    return
  }

  // search if profile token exists and retrieve profile
  var profileFound models.Profile
  errProfile := handler.collectionProfiles.FindOne(handler.ctx, bson.M{
    "apiToken": registerBody.APIToken,
  }).Decode(&profileFound)
  if errProfile != nil {
    handler.logger.Error("REST - POST - PostRegister - Cannot find profile with that apiToken: ", errProfile)
    c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot register, profile token missing or not valid"})
    return
  }

  // search and skip db add if already exists
  var device models.Device
  err := handler.collection.FindOne(handler.ctx, bson.M{
    "mac": registerBody.Mac,
  }).Decode(&device)
  if err == nil {
    handler.logger.Debug("REST - POST - PostRegister - Device already registered")
    // if err == nil => ac found in db (already exists)
    // skip register process returning "already registered"
    c.JSON(http.StatusConflict, gin.H{"message": "Already registered"})
    return
  }

  device = models.Device{}
  device.ID = primitive.NewObjectID()
  device.UUID = uuid.NewString()
  device.Mac = registerBody.Mac
  device.Name = registerBody.Name
  device.Manufacturer = registerBody.Manufacturer
  device.Model = registerBody.Model
  device.CreatedAt = time.Now()
  device.ModifiedAt = time.Now()

  _, errInsert := handler.collection.InsertOne(handler.ctx, device)
  if errInsert != nil {
    handler.logger.Error("REST - POST - PostRegister - Cannot insert the new device")
    c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot insert the new device"})
    return
  }

  // push AC.ID to profile.devices
  _, errUpd := handler.collectionProfiles.UpdateOne(
    handler.ctx,
    bson.M{"_id": profileFound.ID},
    bson.M{"$push": bson.M{"devices": device.ID}},
  )
  if errUpd != nil {
    handler.logger.Error("REST - POST - PostRegister - Cannot update profile with the new device")
    c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot update your profile with the new device"})
    return
  }

  // TODO TODO TODO TODO If here it fails, I should remove the paired device, otherwise I won't be able to register it again
  // Set up a connection to the server.
  tlsCredentials, err := loadTLSCredentials()
  if err != nil {
    handler.logger.Fatal("cannot load TLS credentials: ", err)
  }
  contextBg, cancelBg := context.WithTimeout(context.Background(), 5*time.Second)
  defer cancelBg()
  conn, err := grpc.DialContext(contextBg, handler.grpcTarget, grpc.WithTransportCredentials(tlsCredentials), grpc.WithBlock())
  if err != nil {
    handler.logger.Errorf("Cannot connect via gRPC: %v", err)
    c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot connect to remote service"})
  }
  defer conn.Close()
  client := register.NewRegistrationClient(conn)

  // Contact the server and print out its response.
  ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
  defer cancel()
  r, err := client.Register(ctx, &register.RegisterRequest{
    Id:             device.ID.Hex(),
    Uuid:           device.UUID,
    Mac:            device.Mac,
    Name:           device.Name,
    Manufacturer:   device.Manufacturer,
    Model:          device.Model,
    ProfileOwnerId: profileFound.ID.Hex(),
    ApiToken:       profileFound.ApiToken,
  })
  if err != nil {
    handler.logger.Fatalf("Could not execute gRPC register: %v", err)
    c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot call remote method 'register'"})
  }
  handler.logger.Debug("Register status: ", r.GetStatus())
  handler.logger.Debug("Register message: ", r.GetMessage())

  c.JSON(http.StatusOK, device)
}
