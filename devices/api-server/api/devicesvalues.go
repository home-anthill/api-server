package api

import (
  device3 "api-server/api/gRPC/device"
  custom_errors "api-server/custom-errors"
  "api-server/models"
  "api-server/utils"
  "encoding/json"
  "github.com/gin-gonic/contrib/sessions"
  "github.com/gin-gonic/gin"
  "github.com/go-playground/validator/v10"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/bson/primitive"
  "go.mongodb.org/mongo-driver/mongo"
  "go.uber.org/zap"
  "golang.org/x/net/context"
  "google.golang.org/grpc"
  "google.golang.org/grpc/credentials/insecure"
  "io"
  "net/http"
  "os"
  "reflect"
  "time"
)

type sensorValue struct {
  UUID  string  `json:"uuid"` // feature uuid
  Value float64 `json:"value"`
}

type DevicesValues struct {
  collection         *mongo.Collection
  collectionProfiles *mongo.Collection
  collectionHomes    *mongo.Collection
  ctx                context.Context
  logger             *zap.SugaredLogger
  grpcTarget         string
  sensorGetValueUrl  string
  validate           *validator.Validate
}

func NewDevicesValues(ctx context.Context, logger *zap.SugaredLogger, collection *mongo.Collection, collectionProfiles *mongo.Collection, collectionHomes *mongo.Collection, validate *validator.Validate) *DevicesValues {
  grpcUrl := os.Getenv("GRPC_URL")
  sensorServerUrl := os.Getenv("HTTP_SENSOR_SERVER") + ":" + os.Getenv("HTTP_SENSOR_PORT")
  sensorGetValueUrl := sensorServerUrl + os.Getenv("HTTP_SENSOR_GETVALUE_API")

  return &DevicesValues{
    collection:         collection,
    collectionProfiles: collectionProfiles,
    collectionHomes:    collectionHomes,
    ctx:                ctx,
    logger:             logger,
    grpcTarget:         grpcUrl,
    sensorGetValueUrl:  sensorGetValueUrl,
    validate:           validate,
  }
}

// ------------------------------ Public methods ------------------------------

func (handler *DevicesValues) GetValuesDevice(c *gin.Context) {
  handler.logger.Info("REST - GET - GetValuesDevice called")

  objectId, errId := primitive.ObjectIDFromHex(c.Param("id"))
  if errId != nil {
    handler.logger.Error("REST - GET - GetValuesDevice - wrong format of the path param 'id'")
    c.JSON(http.StatusBadRequest, gin.H{"error": "wrong format of the path param 'id'"})
    return
  }

  // retrieve current profile object from database (using session profile as input)
  session := sessions.Default(c)
  profile, err := utils.GetLoggedProfile(handler.ctx, &session, handler.collectionProfiles)
  if err != nil {
    handler.logger.Error("REST - GET - GetValuesDevice - cannot find profile in session")
    c.JSON(http.StatusUnauthorized, gin.H{"error": "cannot find profile in session"})
    return
  }

  // check if device is in profile (device owned by profile)
  if !isDeviceInProfile(&profile, objectId) {
    handler.logger.Error("REST - GET - GetValuesDevice - this device is not in your profile")
    c.JSON(http.StatusBadRequest, gin.H{"error": "this device is not in your profile"})
    return
  }
  // get device from db
  device, err := handler.getDevice(objectId)
  if err != nil {
    handler.logger.Error("REST - GET - GetValuesDevice - cannot find device")
    c.JSON(http.StatusBadRequest, gin.H{"error": "cannot find device"})
    return
  }

  isController := utils.HasControllerFeature(device.Features)

  if isController {
    // Set up a connection to the server.
    conn, err := grpc.Dial(handler.grpcTarget, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
    if err != nil {
      handler.logger.Error("REST - GET - GetValuesDevice - cannot establish gRPC connection")
      c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot get values - connection error"})
      return
    }
    defer conn.Close()
    client := device3.NewDeviceClient(conn)
    ctx, cancel := context.WithTimeout(context.Background(), time.Second)
    defer cancel()

    response, errSend := client.GetStatus(ctx, &device3.StatusRequest{
      Id:       device.ID.Hex(),
      Mac:      device.Mac,
      ApiToken: profile.ApiToken,
    })

    if errSend != nil {
      handler.logger.Error("REST - GET - GetValuesDevice - cannot get values via gRPC")
      c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot get values"})
      return
    }

    c.JSON(http.StatusOK, gin.H{
      "on":          response.On,
      "temperature": response.Temperature,
      "mode":        response.Mode,
      "fanSpeed":    response.FanSpeed,
      "swing":       response.Swing,
    })
  } else {
    deviceValues := make([]sensorValue, 0)
    for _, feature := range device.Features {
      path := handler.sensorGetValueUrl + device.UUID + "/" + feature.Name
      _, result, err := handler.getSensorValue(path)
      if err != nil {
        // TODO manage errors
        // return custom_errors.Wrap(http.StatusInternalServerError, err, "Cannot register sensor device feature "+feature.Name)
      }

      sensorFeatureValue := sensorValue{}
      err = json.Unmarshal([]byte(result), &sensorFeatureValue)
      if err != nil {
        // TODO
      }
      // add to the object with the value also other information
      // to associate the value to the specific feature
      sensorFeatureValue.UUID = feature.UUID

      handler.logger.Debugf("REST - GetValuesDevice - sensor value for feature = %s is = %#v\n", feature.Name, sensorFeatureValue)
      deviceValues = append(deviceValues, sensorFeatureValue)
    }

    c.JSON(http.StatusOK, deviceValues)

  }
}

func (handler *DevicesValues) PostValueDevice(c *gin.Context) {
  handler.logger.Info("REST - POST - PostValueDevice called")

  objectId, errId := primitive.ObjectIDFromHex(c.Param("id"))
  if errId != nil {
    handler.logger.Errorf("REST - GET - PostValueDevice - wrong format of the path param 'id', err %#v", errId)
    c.JSON(http.StatusBadRequest, gin.H{"error": "wrong format of the path param 'id'"})
    return
  }

  var value models.DeviceState
  if err := c.ShouldBindJSON(&value); err != nil {
    handler.logger.Errorf("REST - POST - PostValueDevice - invalid request payload, err %#v", err)
    c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
    return
  }

  err := handler.validate.Struct(value)
  if err != nil {
    handler.logger.Errorf("REST - POST - PostValueDevice - request body is not valid, err %#v", err)
    var errFields = utils.GetErrorMessage(err)
    c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body, these fields are not valid:" + errFields})
    return
  }

  // retrieve current profile object from database (using session profile as input)
  session := sessions.Default(c)
  profile, err := utils.GetLoggedProfile(handler.ctx, &session, handler.collectionProfiles)
  if err != nil {
    handler.logger.Errorf("REST - GET - PostValueDevice - cannot find profile in session, err %#v", err)
    c.JSON(http.StatusUnauthorized, gin.H{"error": "cannot find profile in session"})
    return
  }

  // check if device is in profile (device owned by profile)
  if !isDeviceInProfile(&profile, objectId) {
    handler.logger.Error("REST - POST - PostValueDevice - this is not your device")
    c.JSON(http.StatusBadRequest, gin.H{"error": "this device is not in your profile"})
    return
  }
  // get device from db
  device, err := handler.getDevice(objectId)
  if err != nil {
    handler.logger.Errorf("REST - POST - PostValueDevice - cannot find device, err %#v", err)
    c.JSON(http.StatusBadRequest, gin.H{"error": "cannot find device"})
    return
  }
  // send via gRPC
  err = handler.sendViaGrpc(&device, &value, profile.ApiToken)
  if err != nil {
    handler.logger.Errorf("REST - POST - PostValueDevice - cannot set value via gRPC, err %#v", err)
    c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot set value"})
    return
  }

  c.JSON(http.StatusOK, gin.H{"message": "set value success"})
}

// ------------------------------ Private methods ------------------------------

func (handler *DevicesValues) getSensorValue(url string) (int, string, error) {
  response, err := http.Get(url)
  if err != nil {
    return -1, "", custom_errors.Wrap(http.StatusInternalServerError, err, "Cannot get sensor value via HTTP")
  }
  defer response.Body.Close()
  body, _ := io.ReadAll(response.Body)
  return response.StatusCode, string(body), nil
}

func (handler *DevicesValues) sendViaGrpc(device *models.Device, value *models.DeviceState, apiToken string) error {
  handler.logger.Infof("gRPC - sendViaGrpc - Called with value = %#v and apiToken = %s", value, apiToken)

  // Set up a connection to the gRPC server.
  securityDialOption, isSecure, err := utils.BuildSecurityDialOption()
  if err != nil {
    handler.logger.Errorf("gRPC - sendViaGrpc - Cannot build security dial option object!, err %#v", err)
    return custom_errors.Wrap(http.StatusInternalServerError, err, "Cannot create securityDialOption to prepare the gRPC connection")
  }
  if isSecure {
    handler.logger.Info("gRPC - sendViaGrpc - GRPC secure enabled!")
  } else {
    handler.logger.Info("gRPC - sendViaGrpc - GRPC secure NOT enabled!")
  }

  contextBg, cancelBg := context.WithTimeout(context.Background(), 5*time.Second)
  defer cancelBg()
  conn, err := grpc.DialContext(contextBg, handler.grpcTarget, securityDialOption, grpc.WithBlock())
  if err != nil {
    handler.logger.Errorf("gRPC - sendViaGrpc - cannot connect via gRPC, err %#v", err)
    return custom_errors.GrpcSendError{
      Status:  custom_errors.ConnectionError,
      Message: "Cannot connect to api-devices",
    }
  }
  defer conn.Close()
  client := device3.NewDeviceClient(conn)

  // -------------------------------------------------------
  // I reach this point only if I can connect to gRPC SERVER
  // -------------------------------------------------------
  handler.logger.Info("gRPC - sendViaGrpc - gRPC server connected")

  clientDeadline := time.Now().Add(time.Duration(200) * time.Millisecond)
  ctx, cancel := context.WithDeadline(contextBg, clientDeadline)
  defer cancel()
  handler.logger.Infof("gRPC - sendViaGrpc - getType(value) %s", getType(value))

  response, errSend := client.SetValues(ctx, &device3.ValuesRequest{
    Id:          device.ID.Hex(),
    Uuid:        device.UUID,
    Mac:         device.Mac,
    On:          value.On,
    Temperature: int32(value.Temperature),
    Mode:        int32(value.Mode),
    FanSpeed:    int32(value.FanSpeed),
    Swing:       value.Swing,
    ApiToken:    apiToken,
  })
  handler.logger.Debugf("gRPC - sendViaGrpc - Device set value status %s", response.GetStatus())
  handler.logger.Debugf("gRPC - sendViaGrpc - Device set value message %s", response.GetMessage())
  return errSend
}

func (handler *DevicesValues) getDevice(deviceId primitive.ObjectID) (models.Device, error) {
  handler.logger.Info("gRPC - getDevice - searching device with objectId: ", deviceId)
  var device models.Device
  err := handler.collection.FindOne(handler.ctx, bson.M{
    "_id": deviceId,
  }).Decode(&device)
  handler.logger.Info("Device found: ", device)
  return device, err
}

// check if the profile contains that device -> if profile is the owner of that device
func isDeviceInProfile(profile *models.Profile, deviceId primitive.ObjectID) bool {
  return utils.Contains(profile.Devices, deviceId)
}

func getType(value interface{}) string {
  if t := reflect.TypeOf(value); t.Kind() == reflect.Ptr {
    return "*" + t.Elem().Name()
  } else {
    return t.Name()
  }
}
