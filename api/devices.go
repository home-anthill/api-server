package api

import (
	"api-server/models"
	"api-server/utils"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"net/http"
	"os"
)

type Devices struct {
	collection         *mongo.Collection
	collectionProfiles *mongo.Collection
	collectionHomes    *mongo.Collection
	ctx                context.Context
	logger             *zap.SugaredLogger
	grpcTarget         string
}

func NewDevices(ctx context.Context, logger *zap.SugaredLogger, collection *mongo.Collection, collectionProfiles *mongo.Collection, collectionHomes *mongo.Collection) *Devices {
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
//
//	'200':
//	    description: Successful operation
func (handler *Devices) GetDevices(c *gin.Context) {
	handler.logger.Info("REST - GET - GetDevices called")

	// retrieve current profile object from session
	session := sessions.Default(c)
	profileSession, err := utils.GetProfileFromSession(&session)
	if err != nil {
		handler.logger.Error("REST - GET - GetDevices - cannot find profile in session")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "cannot find profile in session"})
		return
	}

	// search profile in DB
	// This is required to get fresh data from db, because data in session could be outdated
	var profile models.Profile
	err = handler.collectionProfiles.FindOne(handler.ctx, bson.M{
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
//
//	'200':
//	    description: Successful operation
//	'404':
//	    description: Invalid home ID
func (handler *Devices) DeleteDevice(c *gin.Context) {
	handler.logger.Info("REST - DELETE - DeleteDevice called")

	objectId, errId := primitive.ObjectIDFromHex(c.Param("id"))
	if errId != nil {
		handler.logger.Error("REST - GET - DeleteDevice - wrong format of device id")
		c.JSON(http.StatusBadRequest, gin.H{"error": "wrong format of device id"})
		return
	}

	// retrieve current profile object from session
	session := sessions.Default(c)
	profileSession, err := utils.GetProfileFromSession(&session)
	if err != nil {
		handler.logger.Error("REST - GET - DeleteDevice - cannot find profile in session")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "cannot find profile in session"})
		return
	}

	// search profile in DB
	// This is required to get fresh data from db, because data in session could be outdated
	var profile models.Profile
	err = handler.collectionProfiles.FindOne(handler.ctx, bson.M{
		"_id": profileSession.ID,
	}).Decode(&profile)
	if err != nil {
		handler.logger.Error("REST - DELETE - DeleteDevices - cannot find profile")
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot find profile"})
		return
	}
	// check if the profile contains that device -> if profile is the owner of that device
	found := utils.Contains(profile.Devices, objectId)
	if !found {
		handler.logger.Error("REST - DELETE - DeleteDevices - cannot delete device, because it is not in your profile")
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete device, because it is not in your profile"})
		return
	}

	// update all rooms of homes owned by the profile removing the deviceId
	filter := bson.M{
		// filter homes owned by the profile
		"_id": bson.M{"$in": profile.Homes},
	}
	update := bson.M{
		"$pull": bson.M{
			// using the `all positional operator` https://www.mongodb.com/docs/manual/reference/operator/update/positional-all/
			"rooms.$[].devices": objectId,
		},
	}
	_, err2 := handler.collectionHomes.UpdateMany(handler.ctx, filter, update)
	if err2 != nil {
		handler.logger.Errorf("REST - DELETE - DeleteDevices - cannot update all rooms %#v", err2)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot update rooms"})
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

func (handler *Devices) getDevice(deviceId primitive.ObjectID) (models.Device, error) {
	handler.logger.Info("gRPC - getDevice - searching device with objectId: ", deviceId)
	var device models.Device
	err := handler.collection.FindOne(handler.ctx, bson.M{
		"_id": deviceId,
	}).Decode(&device)
	handler.logger.Info("Device found: ", device)
	return device, err
}
