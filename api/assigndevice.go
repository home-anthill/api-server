package api

import (
	"api-server/models"
	"api-server/utils"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"net/http"
	"time"
)

type AssignDeviceReq struct {
	HomeId string `json:"homeId" validate:"required"`
	RoomId string `json:"roomId" validate:"required"`
}

type AssignDevice struct {
	collectionProfiles *mongo.Collection
	collectionHomes    *mongo.Collection
	ctx                context.Context
	logger             *zap.SugaredLogger
	validate           *validator.Validate
}

func NewAssignDevice(ctx context.Context, logger *zap.SugaredLogger, collectionProfiles *mongo.Collection, collectionHomes *mongo.Collection, validate *validator.Validate) *AssignDevice {
	return &AssignDevice{
		collectionProfiles: collectionProfiles,
		collectionHomes:    collectionHomes,
		ctx:                ctx,
		logger:             logger,
		validate:           validate,
	}
}

func (handler *AssignDevice) PutAssignDeviceToHomeRoom(c *gin.Context) {
	handler.logger.Info("REST - PUT - PutAssignDeviceToHomeRoom called")

	deviceId, errId := primitive.ObjectIDFromHex(c.Param("id"))
	if errId != nil {
		handler.logger.Error("REST - PUT - PutAssignDeviceToHomeRoom - wrong format of device 'id' path param")
		c.JSON(http.StatusBadRequest, gin.H{"error": "wrong format of device 'id' path param"})
		return
	}

	var assignDeviceReq AssignDeviceReq
	if err := c.ShouldBindJSON(&assignDeviceReq); err != nil {
		handler.logger.Error("REST - PUT - PutAssignDeviceToHomeRoom - Cannot bind request body", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}
	if err := handler.validate.Struct(assignDeviceReq); err != nil {
		handler.logger.Errorf("REST - PUT - PutAssignDeviceToHomeRoom - request body is not valid, err %#v", err)
		var errFields = utils.GetErrorMessage(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body, these fields are not valid:" + errFields})
		return
	}
	homeObjId, errHomeObjId := primitive.ObjectIDFromHex(assignDeviceReq.HomeId)
	roomObjId, errRoomObjId := primitive.ObjectIDFromHex(assignDeviceReq.RoomId)
	if errHomeObjId != nil || errRoomObjId != nil {
		handler.logger.Error("REST - PUT - PutAssignDeviceToHomeRoom - wrong format of one of the values in body")
		c.JSON(http.StatusBadRequest, gin.H{"error": "wrong format of one of the values in body"})
		return
	}

	// retrieve current profile object from session
	session := sessions.Default(c)
	profileSession, err := utils.GetProfileFromSession(&session)
	if err != nil {
		handler.logger.Error("REST - GET - PutAssignDeviceToHomeRoom - cannot find profile in session")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "cannot find profile in session"})
		return
	}
	// get the profile from db
	var profile models.Profile
	err = handler.collectionProfiles.FindOne(handler.ctx, bson.M{
		"_id": profileSession.ID,
	}).Decode(&profile)
	if err != nil {
		handler.logger.Error("REST - GET - PutAssignDeviceToHomeRoom - cannot find profile in db")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "cannot find profile"})
		return
	}

	// 1. profile must be the owner of device with id = `deviceId`
	if _, found := utils.Find(profile.Devices, deviceId); !found {
		handler.logger.Errorf("REST - GET - PutAssignDeviceToHomeRoom - profile must be the owner of device with id = '%s'", deviceId)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "you are not the owner of this device id = " + deviceId.Hex()})
		return
	}

	// 2. profile must be the owner of home with id = `assignDeviceReq.HomeId`
	if _, found := utils.Find(profile.Homes, homeObjId); !found {
		handler.logger.Errorf("REST - GET - PutAssignDeviceToHomeRoom - profile must be the owner of home with id = '%s'", assignDeviceReq.HomeId)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "you are not the owner of home id = " + assignDeviceReq.HomeId})
		return
	}

	// 3. `assignDeviceReq.RoomId` must be a room of home with id = `assignDeviceReq.HomeId`
	var home models.Home
	err = handler.collectionHomes.FindOne(handler.ctx, bson.M{
		"_id": homeObjId,
	}).Decode(&home)
	if err != nil {
		handler.logger.Errorf("REST - PUT - PutAssignDeviceToHomeRoom - cannot find home with id = '%s'", assignDeviceReq.HomeId)
		c.JSON(http.StatusNotFound, gin.H{"error": "Cannot find home id = " + assignDeviceReq.HomeId})
		return
	}
	// `roomID` must be a room of `home`
	var roomFound bool
	for _, val := range home.Rooms {
		if val.ID == roomObjId {
			roomFound = true
		}
	}
	if !roomFound {
		handler.logger.Errorf("REST - PUT - PutAssignDeviceToHomeRoom - cannot find room with id = '%s'", assignDeviceReq.RoomId)
		c.JSON(http.StatusNotFound, gin.H{"error": "Cannot find room id = " + assignDeviceReq.RoomId})
		return
	}

	// 4. remove device with id = `deviceId` from all rooms of profile's homes
	filterProfileHomes := bson.M{"_id": bson.M{"$in": profile.Homes}} // filter homes owned by the profile
	updateClean := bson.M{
		"$pull": bson.M{
			// using the `all positional operator` https://www.mongodb.com/docs/manual/reference/operator/update/positional-all/
			"rooms.$[].devices": deviceId,
		},
	}
	_, errClean := handler.collectionHomes.UpdateMany(handler.ctx, filterProfileHomes, updateClean)
	if errClean != nil {
		handler.logger.Errorf("REST - DELETE - PutAssignDeviceToHomeRoom - cannot remove device from all rooms %#v", errClean)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot assign device to home and room"})
		return
	}

	// 5. assign device with id = `deviceId` to room with id = `assignDeviceReq.RoomId` of home with id = `assignDeviceReq.HomeId`
	filterHome := bson.D{bson.E{Key: "_id", Value: homeObjId}}
	arrayFiltersRoom := options.ArrayFilters{Filters: bson.A{bson.M{"x._id": roomObjId}}}
	opts := options.UpdateOptions{
		ArrayFilters: &arrayFiltersRoom,
	}
	update := bson.M{
		"$push": bson.M{
			"rooms.$[x].devices": deviceId,
		},
		"$set": bson.M{
			"rooms.$[x].modifiedAt": time.Now(),
		},
		// TODO I should update `modifiedAt` of both `home` and `room` documents
	}
	_, errUpdate := handler.collectionHomes.UpdateOne(handler.ctx, filterHome, update, &opts)
	if errUpdate != nil {
		handler.logger.Errorf("REST - PUT - PutAssignDeviceToHomeRoom - Cannot assign device to room in DB, errUpdate = %#v", errUpdate)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot assign device to room"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "device has been assigned to room"})
}
