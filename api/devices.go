package api

import (
	"api-server/customerrors"
	"api-server/db"
	"api-server/models"
	"api-server/utils"
	"context"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/writeconcern"
	"go.uber.org/zap"
)

// AssignDeviceReq is the request body for assigning a device to a home room.
type AssignDeviceReq struct {
	HomeID string `json:"homeId" validate:"required"`
	RoomID string `json:"roomId" validate:"required"`
	Name   string `json:"name" validate:"omitempty,max=32"`
}

// Devices handles device registration, lookup, and deletion.
type Devices struct {
	client          *mongo.Client
	collDevices     *mongo.Collection
	collProfiles    *mongo.Collection
	collHomes       *mongo.Collection
	logger          *zap.SugaredLogger
	validate        *validator.Validate
	grpcTarget      string
	onlineByUUIDURL string
}

// NewDevices constructs a Devices handler with the given dependencies.
func NewDevices(logger *zap.SugaredLogger, client *mongo.Client, validate *validator.Validate) *Devices {
	grpcURL := os.Getenv("GRPC_URL")
	onlineServerURL := os.Getenv("HTTP_ONLINE_SERVER") + ":" + os.Getenv("HTTP_ONLINE_PORT")
	onlineByUUIDURL := onlineServerURL + os.Getenv("HTTP_ONLINE_API")

	return &Devices{
		client:          client,
		collDevices:     db.GetCollections(client).Devices,
		collProfiles:    db.GetCollections(client).Profiles,
		collHomes:       db.GetCollections(client).Homes,
		logger:          logger,
		validate:        validate,
		grpcTarget:      grpcURL,
		onlineByUUIDURL: onlineByUUIDURL,
	}
}

// GetDevices function
func (d *Devices) GetDevices(c *gin.Context) {
	d.logger.Info("REST - GET - GetDevices called")

	// retrieve current profile identity from the authenticated context
	profileSession, err := utils.GetProfileFromContext(c)
	if err != nil {
		d.logger.Error("REST - GET - GetDevices - cannot find profile")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "cannot find profile"})
		return
	}

	// search profile in DB
	// This is required to get fresh data from db, because data in session could be outdated
	var profile models.Profile
	err = d.collProfiles.FindOne(c.Request.Context(), bson.M{
		"_id": profileSession.ID,
	}).Decode(&profile)
	if err != nil {
		d.logger.Error("REST - GET - GetDevices - cannot find profile")
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot find profile"})
		return
	}
	// extract Devices from db
	cur, errDevices := d.collDevices.Find(c.Request.Context(), bson.M{
		"_id": bson.M{"$in": profile.Devices},
	})
	if errDevices != nil {
		d.logger.Error("REST - GET - GetDevices - cannot find device in profile")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot find device in profile"})
		return
	}
	defer cur.Close(c.Request.Context())

	devices := make([]models.Device, 0)
	for cur.Next(c.Request.Context()) {
		var device models.Device
		if err := cur.Decode(&device); err != nil {
			d.logger.Errorf("REST - GET - GetDevices - cannot decode device, err = %v", err)
			continue
		}
		devices = append(devices, device)
	}
	c.JSON(http.StatusOK, devices)
}

// DeleteDevice function
func (d *Devices) DeleteDevice(c *gin.Context) {
	d.logger.Info("REST - DELETE - DeleteDevice called")

	objectID, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		d.logger.Error("REST - GET - DeleteDevice - wrong format of device id")
		c.JSON(http.StatusBadRequest, gin.H{"error": "wrong format of device id"})
		return
	}

	// retrieve current profile identity from the authenticated context
	profileSession, err := utils.GetProfileFromContext(c)
	if err != nil {
		d.logger.Error("REST - GET - DeleteDevice - cannot find profile")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "cannot find profile"})
		return
	}

	// search profile in DB
	// This is required to get fresh data from db, because data in session could be outdated
	var profile models.Profile
	err = d.collProfiles.FindOne(c.Request.Context(), bson.M{
		"_id": profileSession.ID,
	}).Decode(&profile)
	if err != nil {
		d.logger.Error("REST - DELETE - DeleteDevices - cannot find profile")
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot find profile"})
		return
	}
	// check if the profile contains that device -> if profile is the owner of that device
	found := utils.Contains(profile.Devices, objectID)
	if !found {
		d.logger.Error("REST - DELETE - DeleteDevices - cannot delete device, because it is not in your profile")
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete device, because it is not in your profile"})
		return
	}

	// search device in DB
	var device models.Device
	err = d.collDevices.FindOne(c.Request.Context(), bson.M{
		"_id": objectID,
	}).Decode(&device)
	if err != nil {
		d.logger.Error("REST - DELETE - DeleteDevices - cannot find device")
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot find device"})
		return
	}

	// start-session
	dbSession, err := d.client.StartSession()
	if err != nil {
		d.logger.Errorf("REST - DELETE - DeleteDevices - cannot start a db session %#v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unknown error while trying to remove a device"})
		return
	}
	// Defers ending the session after the transaction is committed or ended
	defer dbSession.EndSession(c.Request.Context())

	_, errTrans := dbSession.WithTransaction(c.Request.Context(), func(sessionCtx context.Context) (interface{}, error) {
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
		if _, err := d.collHomes.UpdateMany(sessionCtx, filter, update); err != nil {
			d.logger.Errorf("REST - DELETE - DeleteDevices - cannot update all rooms, err = %#v", err)
			return nil, err
		}

		// update profile removing the device from devices
		if _, err := d.collProfiles.UpdateOne(
			sessionCtx,
			bson.M{"_id": profileSession.ID},
			bson.M{"$pull": bson.M{"devices": objectID}},
		); err != nil {
			d.logger.Errorf("REST - DELETE - DeleteDevices - cannot remove device from profile, err = %#v", err)
			return nil, err
		}

		// remove device
		if _, err := d.collDevices.DeleteOne(sessionCtx, bson.M{
			"_id": objectID,
		}); err != nil {
			d.logger.Errorf("REST - DELETE - DeleteDevices - cannot remove device")
			return nil, err
		}

		return nil, nil
	}, options.Transaction().SetWriteConcern(writeconcern.Majority()))
	if errTrans != nil {
		d.logger.Errorf("REST - DELETE - DeleteDevices - cannot remove device updating rooms and profile in transaction, errTrans = %#v", errTrans)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot remove device updating rooms and profile"})
		return
	}

	// if a device is a sensor with online feature, remove it also calling online service
	// We do this OUTSIDE the transaction because HTTP requests are side effects that
	// break idempotency if the transaction needs to retry.
	if utils.HasOnlineFeature(device.Features) {
		d.logger.Debug("REST - DELETE - DeleteDevices - removing online sensor from online service")
		if !utils.IsValidUUID(device.UUID) {
			d.logger.Errorf("REST - DELETE - DeleteDevices - invalid UUID format: device=%s", device.UUID)
		}
		_, result, err := d.deleteOnlineByUUIDService(d.onlineByUUIDURL + url.PathEscape(device.UUID))
		if err != nil {
			d.logger.Errorf("REST - DELETE - DeleteDevices - cannot delete online from remote service = %#v", err)
			if re, ok := err.(*customerrors.ErrorWrapper); ok {
				d.logger.Errorf("REST - DELETE - DeleteDevices - cannot delete online with status = %d, message = %s\n", re.Code, re.Message)
			}
			// DB transaction succeeded, we only log the remote error.
		} else {
			d.logger.Debugf("REST - DELETE - DeleteDevices - result = %#v", result)
		}
	}

	d.logger.Infow("AUDIT - device deleted",
		"profileID", profileSession.ID.Hex(),
		"deviceID", objectID.Hex(),
		"deviceUUID", device.UUID,
	)
	c.JSON(http.StatusOK, gin.H{"message": "device has been deleted"})
}

// PutAssignDeviceToHomeRoom assigns a device to a room within a home and optionally sets the device name.
func (d *Devices) PutAssignDeviceToHomeRoom(c *gin.Context) {
	d.logger.Info("REST - PUT - PutAssignDeviceToHomeRoom called")

	deviceID, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		d.logger.Error("REST - PUT - PutAssignDeviceToHomeRoom - wrong format of device 'id' path param")
		c.JSON(http.StatusBadRequest, gin.H{"error": "wrong format of device 'id' path param"})
		return
	}

	var assignDeviceReq AssignDeviceReq
	if err = c.ShouldBindJSON(&assignDeviceReq); err != nil {
		d.logger.Error("REST - PUT - PutAssignDeviceToHomeRoom - Cannot bind request body", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}
	if err = d.validate.Struct(assignDeviceReq); err != nil {
		d.logger.Errorf("REST - PUT - PutAssignDeviceToHomeRoom - request body is not valid, err %#v", err)
		var errFields = utils.GetErrorMessage(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body, these fields are not valid:" + errFields})
		return
	}
	homeObjID, errHome := bson.ObjectIDFromHex(assignDeviceReq.HomeID)
	roomObjID, errRoom := bson.ObjectIDFromHex(assignDeviceReq.RoomID)
	if errHome != nil || errRoom != nil {
		d.logger.Error("REST - PUT - PutAssignDeviceToHomeRoom - wrong format of one of the values in body")
		c.JSON(http.StatusBadRequest, gin.H{"error": "wrong format of one of the values in body"})
		return
	}

	// retrieve current profile identity from the authenticated context
	profileSession, err := utils.GetProfileFromContext(c)
	if err != nil {
		d.logger.Error("REST - GET - PutAssignDeviceToHomeRoom - cannot find profile")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "cannot find profile"})
		return
	}
	// get the profile from db
	var profile models.Profile
	err = d.collProfiles.FindOne(c.Request.Context(), bson.M{
		"_id": profileSession.ID,
	}).Decode(&profile)
	if err != nil {
		d.logger.Error("REST - GET - PutAssignDeviceToHomeRoom - cannot find profile in db")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "cannot find profile"})
		return
	}

	// 1. profile must be the owner of device with id = `deviceID`
	if _, found := utils.Find(profile.Devices, deviceID); !found {
		d.logger.Errorf("REST - GET - PutAssignDeviceToHomeRoom - profile must be the owner of device with id = '%s'", deviceID)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "you are not the owner of this device id = " + deviceID.Hex()})
		return
	}

	// 2. profile must be the owner of home with id = `assignDeviceReq.HomeID`
	if _, found := utils.Find(profile.Homes, homeObjID); !found {
		d.logger.Errorf("REST - GET - PutAssignDeviceToHomeRoom - profile must be the owner of home with id = '%s'", assignDeviceReq.HomeID)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "you are not the owner of home id = " + assignDeviceReq.HomeID})
		return
	}

	// 3. `assignDeviceReq.RoomID` must be a room of home with id = `assignDeviceReq.HomeID`
	var home models.Home
	err = d.collHomes.FindOne(c.Request.Context(), bson.M{
		"_id": homeObjID,
	}).Decode(&home)
	if err != nil {
		d.logger.Errorf("REST - PUT - PutAssignDeviceToHomeRoom - cannot find home with id = '%s'", assignDeviceReq.HomeID)
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
		d.logger.Errorf("REST - PUT - PutAssignDeviceToHomeRoom - cannot find room with id = '%s'", assignDeviceReq.RoomID)
		c.JSON(http.StatusNotFound, gin.H{"error": "Cannot find room id = " + assignDeviceReq.RoomID})
		return
	}

	// fetch device to get MAC address for default name
	var deviceDoc models.Device
	err = d.collDevices.FindOne(c.Request.Context(), bson.M{"_id": deviceID}).Decode(&deviceDoc)
	if err != nil {
		d.logger.Errorf("REST - PUT - PutAssignDeviceToHomeRoom - cannot find device with id = '%s'", deviceID.Hex())
		c.JSON(http.StatusNotFound, gin.H{"error": "Cannot find device id = " + deviceID.Hex()})
		return
	}

	// use MAC address as default name if none provided
	deviceName := assignDeviceReq.Name
	if deviceName == "" {
		deviceName = deviceDoc.Mac
	}

	// start-session
	dbSession, err := d.client.StartSession()
	if err != nil {
		d.logger.Errorf("REST - PUT - PutAssignDeviceToHomeRoom - cannot start a db session %#v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unknown error while trying to assign a device to a room"})
		return
	}
	// Defers ending the session after the transaction is committed or ended
	defer dbSession.EndSession(c.Request.Context())

	_, errTrans := dbSession.WithTransaction(c.Request.Context(), func(sessionCtx context.Context) (interface{}, error) {
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
		_, errClean := d.collHomes.UpdateMany(sessionCtx, filterProfileHomes, updateClean)
		if errClean != nil {
			d.logger.Errorf("REST - DELETE - PutAssignDeviceToHomeRoom - cannot remove device from all rooms, errClean = %#v", errClean)
			return nil, errClean
		}

		// 5. assign device with id = `deviceID` to room with id = `roomObjID` of home with id = `homeObjID`
		filterHome := bson.D{bson.E{Key: "_id", Value: homeObjID}}
		arrayFiltersRoom := bson.A{bson.M{"x._id": roomObjID}}
		opts := []options.Lister[options.UpdateOneOptions]{
			options.UpdateOne().SetArrayFilters(arrayFiltersRoom),
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
		_, errUpdate := d.collHomes.UpdateOne(sessionCtx, filterHome, update, opts...)
		if errUpdate != nil {
			d.logger.Errorf("REST - PUT - PutAssignDeviceToHomeRoom - cannot assign device to room, errUpdate = %#v", errUpdate)
			return nil, errUpdate
		}

		// 6. update the device name
		_, errName := d.collDevices.UpdateOne(sessionCtx,
			bson.M{"_id": deviceID},
			bson.M{"$set": bson.M{"name": deviceName}},
		)
		if errName != nil {
			d.logger.Errorf("REST - PUT - PutAssignDeviceToHomeRoom - cannot update device name, errName = %#v", errName)
		}
		return nil, errName
	}, options.Transaction().SetWriteConcern(writeconcern.Majority()))
	if errTrans != nil {
		d.logger.Errorf("REST - PUT - PutAssignDeviceToHomeRoom - cannot assign device to room in transaction, errTrans = %#v", errTrans)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot assign device to room in DB"})
		return
	}

	d.logger.Infow("AUDIT - device assigned to room",
		"profileID", profileSession.ID.Hex(),
		"deviceID", deviceID.Hex(),
		"homeID", homeObjID.Hex(),
		"roomID", roomObjID.Hex(),
		"deviceName", deviceName,
	)
	c.JSON(http.StatusOK, gin.H{"message": "device has been assigned to room"})
}

func (d *Devices) deleteOnlineByUUIDService(urlOnline string) (int, string, error) {
	return utils.Delete(urlOnline)
}
