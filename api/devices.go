package api

import (
	"api-server/db"
	"api-server/models"
	"api-server/utils"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"net/http"
	"os"
)

// Devices struct
type Devices struct {
	client       *mongo.Client
	collDevices  *mongo.Collection
	collProfiles *mongo.Collection
	collHomes    *mongo.Collection
	ctx          context.Context
	logger       *zap.SugaredLogger
	grpcTarget   string
}

// NewDevices function
func NewDevices(ctx context.Context, logger *zap.SugaredLogger, client *mongo.Client) *Devices {
	grpcURL := os.Getenv("GRPC_URL")
	return &Devices{
		client:       client,
		collDevices:  db.GetCollections(client).Devices,
		collProfiles: db.GetCollections(client).Profiles,
		collHomes:    db.GetCollections(client).Homes,
		ctx:          ctx,
		logger:       logger,
		grpcTarget:   grpcURL,
	}
}

// GetDevices function
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
	err = handler.collProfiles.FindOne(handler.ctx, bson.M{
		"_id": profileSession.ID,
	}).Decode(&profile)
	if err != nil {
		handler.logger.Error("REST - GET - GetDevices - cannot find profile")
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot find profile"})
		return
	}
	// extract Devices from db
	cur, errDevices := handler.collDevices.Find(handler.ctx, bson.M{
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

// DeleteDevice function
func (handler *Devices) DeleteDevice(c *gin.Context) {
	handler.logger.Info("REST - DELETE - DeleteDevice called")

	objectID, errID := primitive.ObjectIDFromHex(c.Param("id"))
	if errID != nil {
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
	err = handler.collProfiles.FindOne(handler.ctx, bson.M{
		"_id": profileSession.ID,
	}).Decode(&profile)
	if err != nil {
		handler.logger.Error("REST - DELETE - DeleteDevices - cannot find profile")
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot find profile"})
		return
	}
	// check if the profile contains that device -> if profile is the owner of that device
	found := utils.Contains(profile.Devices, objectID)
	if !found {
		handler.logger.Error("REST - DELETE - DeleteDevices - cannot delete device, because it is not in your profile")
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete device, because it is not in your profile"})
		return
	}

	// start-session
	dbSession, err := handler.client.StartSession()
	if err != nil {
		handler.logger.Errorf("REST - DELETE - DeleteDevices - cannot start a db session %#v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unknown error while trying to remove a device"})
		return
	}
	// Defers ending the session after the transaction is committed or ended
	defer dbSession.EndSession(context.TODO())

	_, errTrans := dbSession.WithTransaction(context.TODO(), func(sessionCtx mongo.SessionContext) (interface{}, error) {
		// Official `mongo-driver` documentation state: "callback may be run
		// multiple times during WithTransaction due to retry attempts, so it must be idempotent."

		// update all rooms of homes owned by the profile removing the deviceId
		filter := bson.M{
			// filter homes owned by the profile
			"_id": bson.M{"$in": profile.Homes},
		}
		update := bson.M{
			"$pull": bson.M{
				// using the `all positional operator` https://www.mongodb.com/docs/manual/reference/operator/update/positional-all/
				"rooms.$[].devices": objectID,
			},
		}
		_, err1 := handler.collHomes.UpdateMany(sessionCtx, filter, update)
		if err1 != nil {
			handler.logger.Errorf("REST - DELETE - DeleteDevices - cannot update all rooms, err1 = %#v", err1)
			return nil, err1
		}

		// update profile removing the device from devices
		_, err1 = handler.collProfiles.UpdateOne(
			sessionCtx,
			bson.M{"_id": profileSession.ID},
			bson.M{"$pull": bson.M{"devices": objectID}},
		)
		if err1 != nil {
			handler.logger.Errorf("REST - DELETE - DeleteDevices - cannot remove device from profile, err1 = %#v", err1)
			return nil, err1
		}

		// remove device
		_, err1 = handler.collDevices.DeleteOne(sessionCtx, bson.M{
			"_id": objectID,
		})
		if err1 != nil {
			handler.logger.Errorf("REST - DELETE - DeleteDevices - cannot remove device")
		}
		return nil, err1
	}, options.Transaction().SetWriteConcern(writeconcern.Majority()))
	if errTrans != nil {
		handler.logger.Errorf("REST - DELETE - DeleteDevices - cannot remove device updating rooms and profile in transaction, errTrans = %#v", errTrans)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot remove device updating rooms and profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "device has been deleted"})
}
