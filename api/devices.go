package api

import (
	"api-server/customerrors"
	"api-server/db"
	"api-server/models"
	"api-server/utils"
	"context"
	"io"
	"net/http"
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/writeconcern"
	"go.uber.org/zap"
)

// Devices handles device registration, lookup, and deletion.
type Devices struct {
	client          *mongo.Client
	collDevices     *mongo.Collection
	collProfiles    *mongo.Collection
	collHomes       *mongo.Collection
	logger          *zap.SugaredLogger
	grpcTarget      string
	onlineByUUIDURL string
}

// NewDevices constructs a Devices handler with the given dependencies.
func NewDevices(logger *zap.SugaredLogger, client *mongo.Client) *Devices {
	grpcURL := os.Getenv("GRPC_URL")
	onlineServerURL := os.Getenv("HTTP_ONLINE_SERVER") + ":" + os.Getenv("HTTP_ONLINE_PORT")
	onlineByUUIDURL := onlineServerURL + os.Getenv("HTTP_ONLINE_API")

	return &Devices{
		client:          client,
		collDevices:     db.GetCollections(client).Devices,
		collProfiles:    db.GetCollections(client).Profiles,
		collHomes:       db.GetCollections(client).Homes,
		logger:          logger,
		grpcTarget:      grpcURL,
		onlineByUUIDURL: onlineByUUIDURL,
	}
}

// GetDevices function
func (d *Devices) GetDevices(c *gin.Context) {
	d.logger.Info("REST - GET - GetDevices called")

	// retrieve current profile object from session
	session := sessions.Default(c)
	profileSession, err := utils.GetProfileFromSession(&session)
	if err != nil {
		d.logger.Error("REST - GET - GetDevices - cannot find profile in session")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "cannot find profile in session"})
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

	// retrieve current profile object from session
	session := sessions.Default(c)
	profileSession, err := utils.GetProfileFromSession(&session)
	if err != nil {
		d.logger.Error("REST - GET - DeleteDevice - cannot find profile in session")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "cannot find profile in session"})
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
	// We do this OUTSIDE the transaction because HTTP requests are side-effects that
	// break idempotency if the transaction needs to retry.
	if utils.HasOnlineFeature(device.Features) {
		d.logger.Debug("REST - DELETE - DeleteDevices - removing online sensor from online service")
		_, result, err := d.deleteOnlineByUUIDService(d.onlineByUUIDURL + device.UUID)
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
		"clientIP", c.ClientIP(),
	)
	c.JSON(http.StatusOK, gin.H{"message": "device has been deleted"})
}

func (d *Devices) deleteOnlineByUUIDService(urlOnline string) (int, string, error) {
	req, err := http.NewRequest(http.MethodDelete, urlOnline, nil)
	if err != nil {
		return -1, "", customerrors.Wrap(http.StatusInternalServerError, err, "Cannot create HTTP request for online service")
	}
	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return -1, "", customerrors.Wrap(http.StatusInternalServerError, err, "Cannot call online service via HTTP")
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return response.StatusCode, "", customerrors.Wrap(response.StatusCode, err, "Cannot read response body from online service")
	}
	return response.StatusCode, string(body), nil
}
