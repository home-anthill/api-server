package api

import (
  device3 "api-server/api/gRPC/device"
  custom_errors "api-server/custom-errors"
  "api-server/models"
  "fmt"
  "github.com/gin-gonic/contrib/sessions"
  "github.com/gin-gonic/gin"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/bson/primitive"
  "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/mongo/options"
  "go.uber.org/zap"
  "golang.org/x/net/context"
  "google.golang.org/grpc"
  "google.golang.org/grpc/credentials/insecure"
  "net/http"
  "os"
  "reflect"
  "time"
)

type Devices struct {
  collection         *mongo.Collection
  collectionProfiles *mongo.Collection
  collectionHomes    *mongo.Collection
  ctx                context.Context
  logger             *zap.SugaredLogger
  grpcTarget         string
}

func NewDevices(ctx context.Context,
  logger *zap.SugaredLogger,
  collection *mongo.Collection,
  collectionProfiles *mongo.Collection,
  collectionHomes *mongo.Collection) *Devices {

  grpcUrl := os.Getenv("GRPC_URL")
  return &Devices{
    collection:         collection,
    collectionProfiles: collectionProfiles,
    collectionHomes:    collectionHomes,
    ctx:                ctx,
    logger:             logger,
    grpcTarget:         grpcUrl,
  }
}

// swagger:operation GET /devices devices getDevices
// Returns list of devices
// ---
// produces:
// - application/json
// responses:
//     '200':
//         description: Successful operation
func (handler *Devices) GetDevices(c *gin.Context) {
  handler.logger.Info("REST - GET - GetDevices called")

  // retrieve current profile ID from session
  session := sessions.Default(c)
  profileSession := session.Get("profile").(models.Profile)

  // search profile in DB
  // This is required to get fresh data from db, because data in session could be outdated
  var profile models.Profile
  err := handler.collectionProfiles.FindOne(handler.ctx, bson.M{
    "_id": profileSession.ID,
  }).Decode(&profile)
  if err != nil {
    handler.logger.Error("REST - GET - GetDevices - cannot find profile")
    c.JSON(http.StatusBadRequest, gin.H{"error": "cannot find profile"})
    return
  }
  // extract Devices from db
  cur, errDevices := handler.collection.Find(handler.ctx, bson.M{
    "_id": bson.M{"$in": profile.Devices},
  })
  if errDevices != nil {
    handler.logger.Error("REST - GET - GetDevices - cannot find device in profile")
    c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot find device in profile"})
    return
  }
  defer cur.Close(handler.ctx)

  devices := make([]models.Device, 0)
  for cur.Next(handler.ctx) {
    var device models.Device
    cur.Decode(&device)
    devices = append(devices, device)
  }
  c.JSON(http.StatusOK, devices)
}

// swagger:operation DELETE /devices/{id} devices deleteDevice
// Delete an existing device
// ---
// produces:
// - application/json
// responses:
//     '200':
//         description: Successful operation
//     '404':
//         description: Invalid home ID
func (handler *Devices) DeleteDevice(c *gin.Context) {
  handler.logger.Info("REST - DELETE - DeleteDevice called")

  id := c.Param("id")
  objectId, _ := primitive.ObjectIDFromHex(id)
  homeId := c.Query("homeId")
  objectHomeId, _ := primitive.ObjectIDFromHex(homeId)
  roomId := c.Query("roomId")
  objectRoomId, _ := primitive.ObjectIDFromHex(roomId)

  // retrieve current profile ID from session
  session := sessions.Default(c)
  profileSession := session.Get("profile").(models.Profile)

  // search profile in DB
  // This is required to get fresh data from db, because data in session could be outdated
  var profile models.Profile
  err := handler.collectionProfiles.FindOne(handler.ctx, bson.M{
    "_id": profileSession.ID,
  }).Decode(&profile)
  if err != nil {
    handler.logger.Error("REST - DELETE - DeleteDevices - cannot find profile")
    c.JSON(http.StatusBadRequest, gin.H{"error": "cannot find profile"})
    return
  }
  // check if the profile contains that device -> if profile is the owner of that device
  found := contains(profile.Devices, objectId)
  if !found {
    handler.logger.Error("REST - DELETE - DeleteDevices - cannot delete device, because it is not in your profile")
    c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete device, because it is not in your profile"})
    return
  }

  // update rooms removing the device
  filter := bson.D{primitive.E{Key: "_id", Value: objectHomeId}}
  arrayFilters := options.ArrayFilters{Filters: bson.A{bson.M{"x._id": objectRoomId}}}
  opts := options.UpdateOptions{
    ArrayFilters: &arrayFilters,
  }
  update := bson.M{
    "$pull": bson.M{
      "rooms.$[x].devices": objectId,
    },
  }
  _, err2 := handler.collectionHomes.UpdateOne(handler.ctx, filter, update, &opts)
  if err2 != nil {
    handler.logger.Error("REST - DELETE - DeleteDevices - cannot update room")
    c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot update room"})
    return
  }

  // update profile removing the device from devices
  _, errUpd := handler.collectionProfiles.UpdateOne(
    handler.ctx,
    bson.M{"_id": profileSession.ID},
    bson.M{"$pull": bson.M{"devices": objectId}},
  )
  if errUpd != nil {
    handler.logger.Error("REST - DELETE - DeleteDevices - cannot remove device from profile")
    c.JSON(http.StatusBadRequest, gin.H{"error": "cannot remove device from profile"})
    return
  }

  // remove device
  _, errDel := handler.collection.DeleteOne(handler.ctx, bson.M{
    "_id": objectId,
  })
  if errDel != nil {
    handler.logger.Error("REST - DELETE - DeleteDevices - cannot remove device")
    c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot remove device"})
    return
  }
  c.JSON(http.StatusOK, gin.H{"message": "device has been deleted"})
}

func (handler *Devices) GetValuesDevice(c *gin.Context) {
  handler.logger.Info("REST - GET - GetValuesDevice called")

  id := c.Param("id")
  objectId, _ := primitive.ObjectIDFromHex(id)

  session := sessions.Default(c)
  // get profile from db by user id from session
  profile, err := handler.getProfile(&session)
  if err != nil {
    handler.logger.Error("REST - GET - GetValuesDevice - cannot find profile")
    c.JSON(http.StatusBadRequest, gin.H{"error": "cannot find profile"})
    return
  }
  // check if device is in profile (device owned by profile)
  if !handler.isDeviceInProfile(&profile, objectId) {
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
    "fanMode":     response.FanMode,
    "fanSpeed":    response.FanSpeed,
  })
}

func (handler *Devices) PostOnOffDevice(c *gin.Context) {
  handler.logger.Info("REST - POST - PostOnOffDevice called")

  id := c.Param("id")
  objectId, _ := primitive.ObjectIDFromHex(id)

  var value models.OnOffValue
  if err := c.ShouldBindJSON(&value); err != nil {
    handler.logger.Error("REST - POST - PostOnOffDevice - invalid request payload")
    c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
    return
  }
  session := sessions.Default(c)
  // get profile from db by user id from session
  profile, err := handler.getProfile(&session)
  if err != nil {
    handler.logger.Error("REST - POST - PostOnOffDevice - cannot find profile")
    c.JSON(http.StatusBadRequest, gin.H{"error": "cannot find profile"})
    return
  }
  // check if device is in profile (device owned by profile)
  if !handler.isDeviceInProfile(&profile, objectId) {
    handler.logger.Error("REST - POST - PostOnOffDevice - this is not your device")
    c.JSON(http.StatusBadRequest, gin.H{"error": "this device is not in your profile"})
    return
  }
  // get device from db
  device, err := handler.getDevice(objectId)
  if err != nil {
    handler.logger.Error("REST - POST - PostOnOffDevice - cannot find device")
    c.JSON(http.StatusBadRequest, gin.H{"error": "cannot find device"})
    return
  }
  fmt.Println("prepare to send via gRPC")
  // send via gRPC
  err = handler.sendViaGrpc(&device, &value, profile.ApiToken)
  if err != nil {
    fmt.Println("gRPC cannot send")
    handler.logger.Error("REST - POST - PostOnOffDevice - cannot set value via gRPC")
    c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot set value"})
    return
  }
  fmt.Println("gRPC sent")

  c.JSON(http.StatusOK, gin.H{"message": "set value success"})
}

func (handler *Devices) PostTemperatureDevice(c *gin.Context) {
  handler.logger.Info("REST - POST - PostTemperatureDevice called")

  id := c.Param("id")
  objectId, _ := primitive.ObjectIDFromHex(id)

  var value models.TemperatureValue
  if err := c.ShouldBindJSON(&value); err != nil {
    handler.logger.Error("REST - POST - PostTemperatureDevice - invalid request payload")
    c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
    return
  }
  session := sessions.Default(c)
  // get profile from db by user id from session
  profile, err := handler.getProfile(&session)
  if err != nil {
    handler.logger.Error("REST - POST - PostTemperatureDevice - cannot find profile")
    c.JSON(http.StatusBadRequest, gin.H{"error": "cannot find profile"})
    return
  }
  // check if device is in profile (device owned by profile)
  if !handler.isDeviceInProfile(&profile, objectId) {
    handler.logger.Error("REST - POST - PostTemperatureDevice - this is not your device")
    c.JSON(http.StatusBadRequest, gin.H{"error": "this device is not in your profile"})
    return
  }
  // get device from db
  device, err := handler.getDevice(objectId)
  if err != nil {
    handler.logger.Error("REST - POST - PostTemperatureDevice - cannot find device")
    c.JSON(http.StatusBadRequest, gin.H{"error": "cannot find device"})
    return
  }
  // send via gRPC
  err = handler.sendViaGrpc(&device, &value, profile.ApiToken)
  if err != nil {
    handler.logger.Error("REST - POST - PostTemperatureDevice - cannot set value via gRPC")
    c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot set value"})
    return
  }

  c.JSON(http.StatusOK, gin.H{"message": "set value success"})
}
func (handler *Devices) PostModeDevice(c *gin.Context) {
  handler.logger.Info("REST - POST - PostModeDevice called")

  id := c.Param("id")
  objectId, _ := primitive.ObjectIDFromHex(id)

  var value models.ModeValue
  if err := c.ShouldBindJSON(&value); err != nil {
    handler.logger.Error("REST - POST - PostModeDevice - invalid request payload")
    c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
    return
  }
  session := sessions.Default(c)
  // get profile from db by user id from session
  profile, err := handler.getProfile(&session)
  if err != nil {
    handler.logger.Error("REST - POST - PostModeDevice - cannot find profile")
    c.JSON(http.StatusBadRequest, gin.H{"error": "cannot find profile"})
    return
  }
  // check if device is in profile (device owned by profile)
  if !handler.isDeviceInProfile(&profile, objectId) {
    handler.logger.Error("REST - POST - PostModeDevice - this is not your device")
    c.JSON(http.StatusBadRequest, gin.H{"error": "this device is not in your profile"})
    return
  }
  // get device from db
  device, err := handler.getDevice(objectId)
  if err != nil {
    handler.logger.Error("REST - POST - PostModeDevice - cannot find device")
    c.JSON(http.StatusBadRequest, gin.H{"error": "cannot find device"})
    return
  }
  // send via gRPC
  err = handler.sendViaGrpc(&device, &value, profile.ApiToken)
  if err != nil {
    handler.logger.Error("REST - POST - PostModeDevice - cannot set value via gRPC")
    c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot set value"})
    return
  }

  c.JSON(http.StatusOK, gin.H{"message": "set value success"})
}
func (handler *Devices) PostFanModeDevice(c *gin.Context) {
  handler.logger.Info("REST - POST - PostFanModeDevice called")

  id := c.Param("id")
  objectId, _ := primitive.ObjectIDFromHex(id)

  var value models.FanModeValue
  if err := c.ShouldBindJSON(&value); err != nil {
    handler.logger.Error("REST - POST - PostFanModeDevice - invalid request payload")
    c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
    return
  }
  session := sessions.Default(c)
  // get profile from db by user id from session
  profile, err := handler.getProfile(&session)
  if err != nil {
    handler.logger.Error("REST - POST - PostFanModeDevice - cannot find profile")
    c.JSON(http.StatusBadRequest, gin.H{"error": "cannot find profile"})
    return
  }
  // check if device is in profile (device owned by profile)
  if !handler.isDeviceInProfile(&profile, objectId) {
    handler.logger.Error("REST - POST - PostFanModeDevice - this is not your device")
    c.JSON(http.StatusBadRequest, gin.H{"error": "this device is not in your profile"})
    return
  }
  // get device from db
  device, err := handler.getDevice(objectId)
  if err != nil {
    handler.logger.Error("REST - POST - PostFanModeDevice - cannot find device")
    c.JSON(http.StatusBadRequest, gin.H{"error": "cannot find device"})
    return
  }
  // send via gRPC
  err = handler.sendViaGrpc(&device, &value, profile.ApiToken)
  if err != nil {
    fmt.Println("Cannot set value via GRPC", err)
    handler.logger.Error("REST - POST - PostFanModeDevice - cannot set value via gRPC")
    c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot set value"})
    return
  }

  c.JSON(http.StatusOK, gin.H{"message": "set value success"})
}
func (handler *Devices) PostFanSpeedDevice(c *gin.Context) {
  handler.logger.Info("REST - POST - PostFanSpeedDevice called")

  id := c.Param("id")
  objectId, _ := primitive.ObjectIDFromHex(id)

  var value models.FanSpeedValue
  if err := c.ShouldBindJSON(&value); err != nil {
    handler.logger.Error("REST - POST - PostFanSpeedDevice - invalid request payload")
    c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
    return
  }
  session := sessions.Default(c)
  // get profile from db by user id from session
  profile, err := handler.getProfile(&session)
  if err != nil {
    handler.logger.Error("REST - POST - PostFanSpeedDevice - cannot find profile")
    c.JSON(http.StatusBadRequest, gin.H{"error": "cannot find profile"})
    return
  }
  // check if device is in profile (device owned by profile)
  if !handler.isDeviceInProfile(&profile, objectId) {
    handler.logger.Error("REST - POST - PostFanSpeedDevice - this is not your device")
    c.JSON(http.StatusBadRequest, gin.H{"error": "this device is not in your profile"})
    return
  }
  // get device from db
  device, err := handler.getDevice(objectId)
  if err != nil {
    handler.logger.Error("REST - POST - PostFanSpeedDevice - cannot find device")
    c.JSON(http.StatusBadRequest, gin.H{"error": "cannot find device"})
    return
  }
  // send via gRPC
  err = handler.sendViaGrpc(&device, &value, profile.ApiToken)
  if err != nil {
    handler.logger.Error("REST - POST - PostFanSpeedDevice - cannot set value via gRPC")
    c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot set value"})
    return
  }

  c.JSON(http.StatusOK, gin.H{"message": "set value success"})
}

func (handler *Devices) sendViaGrpc(device *models.Device, value interface{}, apiToken string) error {
  handler.logger.Info("gRPC - sendViaGrpc - Sending device via gRPC...")
  contextBg, cancelBg := context.WithTimeout(context.Background(), 10*time.Second)
  defer cancelBg()
  // Set up a connection to the server.
  conn, err := grpc.DialContext(contextBg, handler.grpcTarget, grpc.WithInsecure(), grpc.WithBlock())
  if err != nil {
    handler.logger.Error("gRPC - sendViaGrpc - cannot connect via gRPC", err)
    return custom_errors.GrpcSendError{
      Status:  custom_errors.ConnectionError,
      Message: "Cannot connect to api-devices",
    }
  }
  defer conn.Close()
  client := device3.NewDeviceClient(conn)

  clientDeadline := time.Now().Add(time.Duration(200) * time.Millisecond)
  ctx, cancel := context.WithDeadline(contextBg, clientDeadline)
  defer cancel()

  switch getType(value) {
  case "*OnOffValue":
    response, errSend := client.SetOnOff(ctx, &device3.OnOffValueRequest{
      Id:       device.ID.Hex(),
      Uuid:     device.UUID,
      Mac:      device.Mac,
      On:       value.(*models.OnOffValue).On,
      ApiToken: apiToken, // RENAME TO ApiToken in proto3
    })
    fmt.Println("Device set value status: ", response.GetStatus())
    fmt.Println("Device set value message: ", response.GetMessage())
    return errSend
  case "*TemperatureValue":
    response, errSend := client.SetTemperature(ctx, &device3.TemperatureValueRequest{
      Id:          device.ID.Hex(),
      Uuid:        device.UUID,
      Mac:         device.Mac,
      Temperature: int32(value.(*models.TemperatureValue).Temperature),
      ApiToken:    apiToken, // RENAME TO ApiToken in proto3
    })
    fmt.Println("Device set value status: ", response.GetStatus())
    fmt.Println("Device set value message: ", response.GetMessage())
    return errSend
  case "*ModeValue":
    response, errSend := client.SetMode(ctx, &device3.ModeValueRequest{
      Id:       device.ID.Hex(),
      Uuid:     device.UUID,
      Mac:      device.Mac,
      Mode:     int32(value.(*models.ModeValue).Mode),
      ApiToken: apiToken, // RENAME TO ApiToken in proto3
    })
    fmt.Println("Device set value status: ", response.GetStatus())
    fmt.Println("Device set value message: ", response.GetMessage())
    return errSend
  case "*FanModeValue":
    response, errSend := client.SetFanMode(ctx, &device3.FanModeValueRequest{
      Id:       device.ID.Hex(),
      Uuid:     device.UUID,
      Mac:      device.Mac,
      FanMode:  int32(value.(*models.FanModeValue).FanMode),
      ApiToken: apiToken, // RENAME TO ApiToken in proto3
    })
    fmt.Println("Device set value status: ", response.GetStatus())
    fmt.Println("Device set value message: ", response.GetMessage())
    return errSend
  case "*FanSpeedValue":
    response, errSend := client.SetFanSpeed(ctx, &device3.FanSpeedValueRequest{
      Id:       device.ID.Hex(),
      Uuid:     device.UUID,
      Mac:      device.Mac,
      FanSpeed: int32(value.(*models.FanSpeedValue).FanSpeed),
      ApiToken: apiToken, // RENAME TO ApiToken in proto3
    })
    fmt.Println("Device set value status: ", response.GetStatus())
    fmt.Println("Device set value message: ", response.GetMessage())
    return errSend
  default:
    handler.logger.Error("gRPC - sendViaGrpc - unknown type")
    return custom_errors.GrpcSendError{
      Status:  custom_errors.BadParams,
      Message: "Cannot cast value",
    }
  }
}

func (handler *Devices) getProfile(session *sessions.Session) (models.Profile, error) {
  profileSession := (*session).Get("profile").(models.Profile)
  // search profile in DB
  // This is required to get fresh data from db, because data in session could be outdated
  var profile models.Profile
  err := handler.collectionProfiles.FindOne(handler.ctx, bson.M{
    "_id": profileSession.ID,
  }).Decode(&profile)
  return profile, err
}

func (handler *Devices) isDeviceInProfile(profile *models.Profile, deviceId primitive.ObjectID) bool {
  // check if the profile contains that device -> if profile is the owner of that device
  return contains(profile.Devices, deviceId)
}

func (handler *Devices) getDevice(deviceId primitive.ObjectID) (models.Device, error) {
  handler.logger.Info("gRPC - getDevice - searching device with objectId: ", deviceId)
  var device models.Device
  err := handler.collection.FindOne(handler.ctx, bson.M{
    "_id": deviceId,
  }).Decode(&device)
  handler.logger.Info("Device found: ", device)
  return device, err
}

func getType(value interface{}) string {
  if t := reflect.TypeOf(value); t.Kind() == reflect.Ptr {
    return "*" + t.Elem().Name()
  } else {
    return t.Name()
  }
}
