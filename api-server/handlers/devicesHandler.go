package handlers

import (
	pbd "api-server/device"
	"api-server/errors"
	"api-server/models"
	"fmt"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"net/http"
	"reflect"
	"time"
)

type DevicesHandler struct {
	collection         *mongo.Collection
	collectionProfiles *mongo.Collection
	collectionHomes    *mongo.Collection
	ctx                context.Context
}

func NewDevicesHandler(ctx context.Context,
		collection *mongo.Collection,
		collectionProfiles *mongo.Collection,
		collectionHomes *mongo.Collection) *DevicesHandler {
	return &DevicesHandler{
		collection:         collection,
		collectionProfiles: collectionProfiles,
		collectionHomes:    collectionHomes,
		ctx:                ctx,
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
func (handler *DevicesHandler) GetDevicesHandler(c *gin.Context) {
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot find profile"})
		return
	}
	// extract Devices from db
	cur, errDevices := handler.collection.Find(handler.ctx, bson.M{
		"_id": bson.M{"$in": profile.Devices},
	})
	if errDevices != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errDevices.Error()})
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
func (handler *DevicesHandler) DeleteDeviceHandler(c *gin.Context) {
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot find profile"})
		return
	}
	// check if the profile contains that device -> if profile is the owner of that device
	found := contains(profile.Devices, objectId)
	if !found {
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot remove device from profile"})
		return
	}

	// remove device
	_, errDel := handler.collection.DeleteOne(handler.ctx, bson.M{
		"_id": objectId,
	})
	if errDel != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errDel.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "device has been deleted"})
}

func (handler *DevicesHandler) PostOnOffDeviceHandler(c *gin.Context) {
	id := c.Param("id")
	objectId, _ := primitive.ObjectIDFromHex(id)

	var value models.OnOffValue
	if err := c.ShouldBindJSON(&value); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	session := sessions.Default(c)
	// get profile from db by user id from session
	profile, err := handler.getProfile(&session)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot find profile"})
		return
	}
	// check if device is in profile (device owned by profile)
	if !handler.isDeviceInProfile(&profile, objectId) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "This device is not in your profile"})
		return
	}
	// get device from db
	device, err := handler.getDevice(objectId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot find device"})
		return
	}
	// send via gRPC
	err = handler.sendViaGrpc(&device, &value, profile.ApiToken)
	if err != nil {
		fmt.Println("Cannot set value via GRPC", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot set value"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Set value success"})
}

func (handler *DevicesHandler) PostTemperatureDeviceHandler(c *gin.Context) {
	id := c.Param("id")
	objectId, _ := primitive.ObjectIDFromHex(id)

	var value models.TemperatureValue
	if err := c.ShouldBindJSON(&value); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	session := sessions.Default(c)
	// get profile from db by user id from session
	profile, err := handler.getProfile(&session)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot find profile"})
		return
	}
	// check if device is in profile (device owned by profile)
	if !handler.isDeviceInProfile(&profile, objectId) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "This device is not in your profile"})
		return
	}
	// get device from db
	device, err := handler.getDevice(objectId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot find device"})
		return
	}
	// send via gRPC
	err = handler.sendViaGrpc(&device, &value, profile.ApiToken)
	if err != nil {
		fmt.Println("Cannot set value via GRPC", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot set value"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Set value success"})
}
func (handler *DevicesHandler) PostModeDeviceHandler(c *gin.Context) {
	id := c.Param("id")
	objectId, _ := primitive.ObjectIDFromHex(id)

	var value models.ModeValue
	if err := c.ShouldBindJSON(&value); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	session := sessions.Default(c)
	// get profile from db by user id from session
	profile, err := handler.getProfile(&session)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot find profile"})
		return
	}
	// check if device is in profile (device owned by profile)
	if !handler.isDeviceInProfile(&profile, objectId) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "This device is not in your profile"})
		return
	}
	// get device from db
	device, err := handler.getDevice(objectId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot find device"})
		return
	}
	// send via gRPC
	err = handler.sendViaGrpc(&device, &value, profile.ApiToken)
	if err != nil {
		fmt.Println("Cannot set value via GRPC", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot set value"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Set value success"})
}
func (handler *DevicesHandler) PostFanModeDeviceHandler(c *gin.Context) {
	id := c.Param("id")
	objectId, _ := primitive.ObjectIDFromHex(id)

	var value models.FanModeValue
	if err := c.ShouldBindJSON(&value); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	session := sessions.Default(c)
	// get profile from db by user id from session
	profile, err := handler.getProfile(&session)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot find profile"})
		return
	}
	// check if device is in profile (device owned by profile)
	if !handler.isDeviceInProfile(&profile, objectId) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "This device is not in your profile"})
		return
	}
	// get device from db
	device, err := handler.getDevice(objectId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot find device"})
		return
	}
	// send via gRPC
	err = handler.sendViaGrpc(&device, &value, profile.ApiToken)
	if err != nil {
		fmt.Println("Cannot set value via GRPC", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot set value"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Set value success"})
}
func (handler *DevicesHandler) PostFanSwingDeviceHandler(c *gin.Context) {
	id := c.Param("id")
	objectId, _ := primitive.ObjectIDFromHex(id)

	var value models.FanSwingValue
	if err := c.ShouldBindJSON(&value); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	session := sessions.Default(c)
	// get profile from db by user id from session
	profile, err := handler.getProfile(&session)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot find profile"})
		return
	}
	// check if device is in profile (device owned by profile)
	if !handler.isDeviceInProfile(&profile, objectId) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "This device is not in your profile"})
		return
	}
	// get device from db
	device, err := handler.getDevice(objectId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot find device"})
		return
	}
	// send via gRPC
	err = handler.sendViaGrpc(&device, &value, profile.ApiToken)
	if err != nil {
		fmt.Println("Cannot set value via GRPC", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot set value"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Set value success"})
}

func (handler *DevicesHandler) sendViaGrpc(device *models.Device, value interface{}, apiToken string) error {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		fmt.Println("Cannot connect via GRPC", err)
		return errors.SendGrpcError{
			Status:  errors.ConnectionError,
			Message: "Cannot connect to api-devices",
		}
	}
	defer conn.Close()
	client := pbd.NewDeviceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	switch getType(value) {
	case "*OnOffValue":
		response, errSend := client.SetOnOff(ctx, &pbd.OnOffValueRequest{
			Id:           device.ID.Hex(),
			Mac:          device.Mac,
			On:           value.(*models.OnOffValue).On,
			ProfileToken: apiToken, // RENAME TO ApiToken in proto3
		})
		fmt.Println("Device set value status: ", response.GetStatus())
		fmt.Println("Device set value message: ", response.GetMessage())
		return errSend
	case "*TemperatureValue":
		response, errSend := client.SetTemperature(ctx, &pbd.TemperatureValueRequest{
			Id:           device.ID.Hex(),
			Mac:          device.Mac,
			Temperature:  int32(value.(*models.TemperatureValue).Temperature),
			ProfileToken: apiToken, // RENAME TO ApiToken in proto3
		})
		fmt.Println("Device set value status: ", response.GetStatus())
		fmt.Println("Device set value message: ", response.GetMessage())
		return errSend
	case "*ModeValue":
		response, errSend := client.SetMode(ctx, &pbd.ModeValueRequest{
			Id:           device.ID.Hex(),
			Mac:          device.Mac,
			Mode:         int32(value.(*models.ModeValue).Mode),
			ProfileToken: apiToken, // RENAME TO ApiToken in proto3
		})
		fmt.Println("Device set value status: ", response.GetStatus())
		fmt.Println("Device set value message: ", response.GetMessage())
		return errSend
	case "*FanModeValue":
		response, errSend := client.SetFanMode(ctx, &pbd.FanModeValueRequest{
			Id:           device.ID.Hex(),
			Mac:          device.Mac,
			Fan:          int32(value.(*models.FanModeValue).Fan),
			ProfileToken: apiToken, // RENAME TO ApiToken in proto3
		})
		fmt.Println("Device set value status: ", response.GetStatus())
		fmt.Println("Device set value message: ", response.GetMessage())
		return errSend
	case "*FanSwingValue":
		response, errSend := client.SetFanSwing(ctx, &pbd.FanSwingValueRequest{
			Id:           device.ID.Hex(),
			Mac:          device.Mac,
			Swing:        value.(*models.FanSwingValue).Swing,
			ProfileToken: apiToken, // RENAME TO ApiToken in proto3
		})
		fmt.Println("Device set value status: ", response.GetStatus())
		fmt.Println("Device set value message: ", response.GetMessage())
		return errSend
	default:
		return errors.SendGrpcError{
			Status:  errors.BadParams,
			Message: "Cannot cast value",
		}
	}
}

func (handler *DevicesHandler) getProfile(session *sessions.Session) (models.Profile, error) {
	profileSession := (*session).Get("profile").(models.Profile)
	// search profile in DB
	// This is required to get fresh data from db, because data in session could be outdated
	var profile models.Profile
	err := handler.collectionProfiles.FindOne(handler.ctx, bson.M{
		"_id": profileSession.ID,
	}).Decode(&profile)
	return profile, err
}

func (handler *DevicesHandler) isDeviceInProfile(profile *models.Profile, deviceId primitive.ObjectID) bool {
	// check if the profile contains that device -> if profile is the owner of that device
	return contains(profile.Devices, deviceId)
}

func (handler *DevicesHandler) getDevice(deviceId primitive.ObjectID) (models.Device, error) {
	var device models.Device
	err := handler.collection.FindOne(handler.ctx, bson.M{
		"_id": deviceId,
	}).Decode(&device)
	return device, err
}

func getType(value interface{}) string {
	if t := reflect.TypeOf(value); t.Kind() == reflect.Ptr {
		return "*" + t.Elem().Name()
	} else {
		return t.Name()
	}
}
