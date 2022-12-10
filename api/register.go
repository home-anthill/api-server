package api

import (
  "api-server/api/gRPC/register"
  "api-server/custom-errors"
  "api-server/models"
  "api-server/utils"
  "bytes"
  "encoding/json"
  "github.com/gin-gonic/gin"
  "github.com/go-playground/validator/v10"
  "github.com/google/uuid"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/bson/primitive"
  "go.mongodb.org/mongo-driver/mongo"
  "go.uber.org/zap"
  "golang.org/x/net/context"
  "google.golang.org/grpc"
  "io"
  "net/http"
  "os"
  "time"
)

type featureReq struct {
  Type     models.Type `json:"type" validate:"required,oneof='controller' 'sensor'"`
  Name     string      `json:"name" validate:"required,min=2,max=20"`
  Enable   bool        `json:"enable" validate:"required,boolean"`
  Order    int         `json:"order" validate:"required,gte=1"`
  Unit     string      `json:"unit" validate:"required,min=1,max=10"`
}

type deviceRegisterReq struct {
  Mac          string       `json:"mac" validate:"required,mac"`
  Manufacturer string       `json:"manufacturer" validate:"required,min=3,max=50"`
  Model        string       `json:"model" validate:"required,min=3,max=20"`
  ApiToken     string       `json:"apiToken" validate:"required,uuid4"`
  Features     []featureReq `json:"features" validate:"required,dive"`
}

type sensorRegisterReq struct {
  Uuid           string `json:"uuid"`
  Mac            string `json:"mac"`
  Manufacturer   string `json:"manufacturer"`
  Model          string `json:"model"`
  ProfileOwnerId string `json:"profileOwnerId"`
  ApiToken       string `json:"apiToken"`
}

type Register struct {
  collection         *mongo.Collection
  collectionProfiles *mongo.Collection
  ctx                context.Context
  logger             *zap.SugaredLogger
  grpcTarget         string
  keepAliveSensorUrl string
  registerSensorUrl  string
  validate           *validator.Validate
}

func ControllerDeviceValidation(sl validator.StructLevel) {
  deviceReq, _ := sl.Current().Interface().(deviceRegisterReq)
  var isController = false
  for _, feature := range deviceReq.Features {
    if feature.Type == models.Controller {
      isController = true
    }
  }
  if isController && len(deviceReq.Features) > 1 {
    sl.ReportError(deviceReq.Features, "features", "Features", "features", "")
  }
}

func NewRegister(ctx context.Context, logger *zap.SugaredLogger, collection *mongo.Collection, collectionProfiles *mongo.Collection, validate *validator.Validate) *Register {
  grpcUrl := os.Getenv("GRPC_URL")
  sensorServerUrl := os.Getenv("HTTP_SENSOR_SERVER") + ":" + os.Getenv("HTTP_SENSOR_PORT")
  keepAliveSensorUrl := sensorServerUrl + os.Getenv("HTTP_SENSOR_KEEPALIVE_API")
  registerSensorUrl := sensorServerUrl + os.Getenv("HTTP_SENSOR_REGISTER_API")

  // add custom validators
  // validator to check that a controller device can have only one feature
  validate.RegisterStructValidation(ControllerDeviceValidation, deviceRegisterReq{})

  return &Register{
    collection:         collection,
    collectionProfiles: collectionProfiles,
    ctx:                ctx,
    logger:             logger,
    grpcTarget:         grpcUrl,
    keepAliveSensorUrl: keepAliveSensorUrl,
    registerSensorUrl:  registerSensorUrl,
    validate:           validate,
  }
}

func (handler *Register) PostRegister(c *gin.Context) {
  handler.logger.Info("REST - PostRegister called")

  var registerBody deviceRegisterReq
  if err := c.ShouldBindJSON(&registerBody); err != nil {
    handler.logger.Errorf("REST - PostRegister - Cannot bind request body. Err = %v\n", err)
    c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
    return
  }

  err := handler.validate.Struct(registerBody)
  if err != nil {
    handler.logger.Errorf("REST - PostRegister - request body is not valid, err %#v", err)
    var errFields = utils.GetErrorMessage(err)
    c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body, these fields are not valid:" + errFields})
    return
  }

  // search if profile token exists and retrieve profile
  var profileFound models.Profile
  errProfile := handler.collectionProfiles.FindOne(handler.ctx, bson.M{
    "apiToken": registerBody.ApiToken,
  }).Decode(&profileFound)
  if errProfile != nil {
    handler.logger.Errorf("REST - PostRegister - Cannot find profile with that apiToken. Err = %v\n", errProfile)
    c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot register, profile token missing or not valid"})
    return
  }

  // search and skip db add if device already exists
  var device models.Device
  err = handler.collection.FindOne(handler.ctx, bson.M{
    "mac": registerBody.Mac,
  }).Decode(&device)
  if err == nil {
    handler.logger.Info("REST - PostRegister - Device already registered")
    // if err == nil => device found in db (already exists)
    // skip register process returning "already registered"
    c.JSON(http.StatusConflict, gin.H{"message": "Already registered"})
    return
  }

  isController := false
  for _, feature := range registerBody.Features {
    if feature.Type == models.Controller {
      isController = true
    }
  }

  handler.logger.Debugf("REST - PostRegister - registering device, isController = %v", isController)

  insertDate := time.Now()
  device = models.Device{}
  device.ID = primitive.NewObjectID()
  device.UUID = uuid.NewString()
  device.Mac = registerBody.Mac
  device.Manufacturer = registerBody.Manufacturer
  device.Model = registerBody.Model
  device.CreatedAt = insertDate
  device.ModifiedAt = insertDate

  device.Features = utils.MapSlice(registerBody.Features, func(fReq featureReq) models.Feature {
    return models.Feature{
      UUID:     uuid.NewString(),
      Type:     fReq.Type,
      Name:     fReq.Name,
      Enable:   fReq.Enable,
      Order:    fReq.Order,
      Unit:     fReq.Unit,
    }
  })

  // if it's a controller device => call gRPC
  // otherwise REST call to sensor service

  if isController {
    status, message, err := handler.registerControllerViaGRPC(&device, &profileFound)
    if err != nil {
      handler.logger.Errorf("REST - PostRegister - cannot register controller device via gRPC. Err %v\n", err)
      if re, ok := err.(*custom_errors.ErrorWrapper); ok {
        handler.logger.Errorf("REST - PostRegister - cannot register device with status = %d, message = %s\n", re.Code, re.Message)
      }
      c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot register controller device"})
      return
    }
    handler.logger.Infof("REST - PostRegister - controller device registered with status = %s, message = %s\n", status, message)
  } else {
    err := handler.registerSensorViaHTTP(&device, &profileFound)
    if err != nil {
      handler.logger.Errorf("REST - PostRegister - cannot register sensor device via HTTP. Err %v\n", err)
      if re, ok := err.(*custom_errors.ErrorWrapper); ok {
        handler.logger.Errorf("REST - PostRegister - cannot register device with status = %d, message = %s\n", re.Code, re.Message)
      }
      c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot register sensor device"})
      return
    }
    handler.logger.Info("REST - PostRegister - sensor device registered successfully")
  }

  handler.logger.Debugf("REST - PostRegister - registered device = %#v", device)
  c.JSON(http.StatusOK, device)
}

func (handler *Register) registerSensorViaHTTP(device *models.Device, profileFound *models.Profile) error {
  // check if service is available calling keep-alive
  // TODO remove this in a production code
  _, _, keepAliveErr := handler.keepAliveSensorService(handler.keepAliveSensorUrl)
  if keepAliveErr != nil {
    return custom_errors.Wrap(http.StatusInternalServerError, keepAliveErr, "Cannot call keepAlive of remote register service")
  }

  // Insert device into api-server database
  errInsDb := handler.insertDevice(device, profileFound)
  if errInsDb != nil {
    return custom_errors.Wrap(http.StatusInternalServerError, errInsDb, "Cannot insert the new device")
  }

  // do the real call to the remote registration service
  payload := sensorRegisterReq{
    Uuid:           device.UUID,
    Mac:            device.Mac,
    Manufacturer:   device.Manufacturer,
    Model:          device.Model,
    ProfileOwnerId: profileFound.ID.Hex(),
    ApiToken:       profileFound.ApiToken,
  }
  payloadJSON, err := json.Marshal(payload)
  if err != nil {
    return custom_errors.Wrap(http.StatusInternalServerError, err, "Cannot create payload to register sensor service")
  }

  for _, feature := range device.Features {
    _, _, err := handler.registerSensor(handler.registerSensorUrl+feature.Name, payloadJSON)
    if err != nil {
      return custom_errors.Wrap(http.StatusInternalServerError, err, "Cannot register sensor device feature "+feature.Name)
    }
    //handler.logger.Debugf("REST - PostRegister - sensor device registered with status= %d, body= %s\n", statusCode, respBody)
  }
  return nil
}

func (handler *Register) registerControllerViaGRPC(device *models.Device, profileFound *models.Profile) (string, string, error) {
  handler.logger.Info("gRPC - registerControllerViaGRPC - Sending register via gRPC...")
  // Set up a connection to the gRPC server.
  securityDialOption, isSecure, err := utils.BuildSecurityDialOption()
  if err != nil {
    return "", "", custom_errors.Wrap(http.StatusInternalServerError, err, "Cannot create securityDialOption to prepare the gRPC connection")
  }
  if isSecure {
    handler.logger.Debug("registerControllerViaGRPC - GRPC secure enabled!")
  } else {
    handler.logger.Info("registerControllerViaGRPC - GRPC secure NOT enabled!")
  }

  contextBg, cancelBg := context.WithTimeout(context.Background(), 5*time.Second)
  defer cancelBg()
  conn, err := grpc.DialContext(contextBg, handler.grpcTarget, securityDialOption, grpc.WithBlock())
  if err != nil {
    handler.logger.Error("gRPC - registerControllerViaGRPC - cannot connect via gRPC", err)
    return "", "", custom_errors.GrpcSendError{
      Status:  custom_errors.ConnectionError,
      Message: "Cannot connect to api-devices",
    }
  }
  defer conn.Close()
  client := register.NewRegistrationClient(conn)

  // -------------------------------------------------------
  // I reach this point only if I can connect to gRPC SERVER
  // -------------------------------------------------------

  // Insert device into api-server database
  errInsDb := handler.insertDevice(device, profileFound)
  if errInsDb != nil {
    return "", "", custom_errors.Wrap(http.StatusInternalServerError, errInsDb, "Cannot insert the new device")
  }

  // Contact the server and print out its response.
  ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
  defer cancel()
  r, err := client.Register(ctx, &register.RegisterRequest{
    Id:             device.ID.Hex(),
    Uuid:           device.UUID,
    Mac:            device.Mac,
    Name:           device.Features[0].Name,
    Manufacturer:   device.Manufacturer,
    Model:          device.Model,
    ProfileOwnerId: profileFound.ID.Hex(),
    ApiToken:       profileFound.ApiToken,
  })
  if err != nil {
    return "", "", custom_errors.Wrap(http.StatusInternalServerError, err, "Cannot invoke Register via gRPC")
  }
  return r.GetStatus(), r.GetMessage(), nil
}

func (handler *Register) keepAliveSensorService(urlKeepAlive string) (int, string, error) {
  response, err := http.Get(urlKeepAlive)
  if err != nil {
    return -1, "", custom_errors.Wrap(http.StatusInternalServerError, err, "Cannot call keepAlive of the remote register service via HTTP")
  }
  defer response.Body.Close()
  body, _ := io.ReadAll(response.Body)
  return response.StatusCode, string(body), nil
}

func (handler *Register) registerSensor(urlRegister string, payloadJSON []byte) (int, string, error) {
  var payloadBody = bytes.NewBuffer(payloadJSON)
  response, err := http.Post(urlRegister, "application/json", payloadBody)
  if err != nil {
    return -1, "", custom_errors.Wrap(http.StatusInternalServerError, err, "Cannot register to sensor service via HTTP")
  }
  defer response.Body.Close()
  body, _ := io.ReadAll(response.Body)
  return response.StatusCode, string(body), nil
}

func (handler *Register) insertDevice(device *models.Device, profile *models.Profile) error {
  // Insert device
  _, errInsert := handler.collection.InsertOne(handler.ctx, device)
  if errInsert != nil {
    return custom_errors.Wrap(http.StatusInternalServerError, errInsert, "Cannot insert the new device")
  }
  // push device.ID to profile.devices into api-server database
  _, errUpd := handler.collectionProfiles.UpdateOne(
    handler.ctx,
    bson.M{"_id": profile.ID},
    bson.M{"$push": bson.M{"devices": device.ID}},
  )
  if errUpd != nil {
    return custom_errors.Wrap(http.StatusInternalServerError, errUpd, "Cannot update profile with the new device")
  }
  return nil
}
