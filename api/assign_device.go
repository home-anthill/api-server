package api

import (
	"api-server/db"
	"api-server/models"
	"api-server/utils"
	"context"
	"net/http"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"go.uber.org/zap"
)

// AssignDeviceReq struct
type AssignDeviceReq struct {
	HomeID string `json:"homeId" validate:"required"`
	RoomID string `json:"roomId" validate:"required"`
}

// AssignDevice struct
type AssignDevice struct {
	client       *mongo.Client
	collProfiles *mongo.Collection
	collHomes    *mongo.Collection
	ctx          context.Context
	logger       *zap.SugaredLogger
	validate     *validator.Validate
}

// NewAssignDevice function
func NewAssignDevice(ctx context.Context, logger *zap.SugaredLogger, client *mongo.Client, validate *validator.Validate) *AssignDevice {
	return &AssignDevice{
		client:       client,
		collProfiles: db.GetCollections(client).Profiles,
		collHomes:    db.GetCollections(client).Homes,
		ctx:          ctx,
		logger:       logger,
		validate:     validate,
	}
}

// PutAssignDeviceToHomeRoom function
func (handler *AssignDevice) PutAssignDeviceToHomeRoom(c *gin.Context) {
	handler.logger.Info("REST - PUT - PutAssignDeviceToHomeRoom called")

	deviceID, errID := primitive.ObjectIDFromHex(c.Param("id"))
	if errID != nil {
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
	homeObjID, errHomeObjID := primitive.ObjectIDFromHex(assignDeviceReq.HomeID)
	roomObjID, errRoomObjID := primitive.ObjectIDFromHex(assignDeviceReq.RoomID)
	if errHomeObjID != nil || errRoomObjID != nil {
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
	err = handler.collProfiles.FindOne(handler.ctx, bson.M{
		"_id": profileSession.ID,
	}).Decode(&profile)
	if err != nil {
		handler.logger.Error("REST - GET - PutAssignDeviceToHomeRoom - cannot find profile in db")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "cannot find profile"})
		return
	}

	// 1. profile must be the owner of device with id = `deviceID`
	if _, found := utils.Find(profile.Devices, deviceID); !found {
		handler.logger.Errorf("REST - GET - PutAssignDeviceToHomeRoom - profile must be the owner of device with id = '%s'", deviceID)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "you are not the owner of this device id = " + deviceID.Hex()})
		return
	}

	// 2. profile must be the owner of home with id = `assignDeviceReq.HomeID`
	if _, found := utils.Find(profile.Homes, homeObjID); !found {
		handler.logger.Errorf("REST - GET - PutAssignDeviceToHomeRoom - profile must be the owner of home with id = '%s'", assignDeviceReq.HomeID)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "you are not the owner of home id = " + assignDeviceReq.HomeID})
		return
	}

	// 3. `assignDeviceReq.RoomID` must be a room of home with id = `assignDeviceReq.HomeID`
	var home models.Home
	err = handler.collHomes.FindOne(handler.ctx, bson.M{
		"_id": homeObjID,
	}).Decode(&home)
	if err != nil {
		handler.logger.Errorf("REST - PUT - PutAssignDeviceToHomeRoom - cannot find home with id = '%s'", assignDeviceReq.HomeID)
		c.JSON(http.StatusNotFound, gin.H{"error": "Cannot find home id = " + assignDeviceReq.HomeID})
		return
	}
	// `roomID` must be a room of `home`
	var roomFound bool
	for _, val := range home.Rooms {
		if val.ID == roomObjID {
			roomFound = true
		}
	}
	if !roomFound {
		handler.logger.Errorf("REST - PUT - PutAssignDeviceToHomeRoom - cannot find room with id = '%s'", assignDeviceReq.RoomID)
		c.JSON(http.StatusNotFound, gin.H{"error": "Cannot find room id = " + assignDeviceReq.RoomID})
		return
	}

	// start-session
	dbSession, err := handler.client.StartSession()
	if err != nil {
		handler.logger.Errorf("REST - PUT - PutAssignDeviceToHomeRoom - cannot start a db session %#v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unknown error while trying to assign a device to a room"})
		return
	}
	// Defers ending the session after the transaction is committed or ended
	defer dbSession.EndSession(context.TODO())

	_, errTrans := dbSession.WithTransaction(context.TODO(), func(sessionCtx mongo.SessionContext) (interface{}, error) {
		// Official `mongo-driver` documentation state: "callback may be run
		// multiple times during WithTransaction due to retry attempts, so it must be idempotent."

		// 4. remove device with id = `deviceID` from all rooms of profile's homes
		filterProfileHomes := bson.M{"_id": bson.M{"$in": profile.Homes}} // filter homes owned by the profile
		updateClean := bson.M{
			"$pull": bson.M{
				// using the `all positional operator` https://www.mongodb.com/docs/manual/reference/operator/update/positional-all/
				"rooms.$[].devices": deviceID,
			},
		}
		_, errClean := handler.collHomes.UpdateMany(sessionCtx, filterProfileHomes, updateClean)
		if errClean != nil {
			handler.logger.Errorf("REST - DELETE - PutAssignDeviceToHomeRoom - cannot remove device from all rooms, errClean = %#v", errClean)
			return nil, errClean
		}

		// 5. assign device with id = `deviceID` to room with id = `roomObjID` of home with id = `homeObjID`
		filterHome := bson.D{bson.E{Key: "_id", Value: homeObjID}}
		arrayFiltersRoom := options.ArrayFilters{Filters: bson.A{bson.M{"x._id": roomObjID}}}
		opts := options.UpdateOptions{
			ArrayFilters: &arrayFiltersRoom,
		}
		update := bson.M{
			"$addToSet": bson.M{
				"rooms.$[x].devices": deviceID,
			},
			"$set": bson.M{
				"rooms.$[x].modifiedAt": time.Now(),
				"modifiedAt":            time.Now(),
			},
		}
		_, errUpdate := handler.collHomes.UpdateOne(sessionCtx, filterHome, update, &opts)
		if errUpdate != nil {
			handler.logger.Errorf("REST - PUT - PutAssignDeviceToHomeRoom - cannot assign device to room, errUpdate = %#v", errUpdate)
		}
		return nil, errUpdate
	}, options.Transaction().SetWriteConcern(writeconcern.Majority()))
	if errTrans != nil {
		handler.logger.Errorf("REST - PUT - PutAssignDeviceToHomeRoom - cannot assign device to room in transaction, errTrans = %#v", errTrans)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot assign device to room in DB"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "device has been assigned to room"})
}
