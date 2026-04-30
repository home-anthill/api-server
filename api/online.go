package api

import (
	"api-server/customerrors"
	"api-server/db"
	"api-server/models"
	"api-server/utils"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"
)

type onlineResponse struct {
	UUID        string `json:"uuid"`
	APIToken    string `json:"apiToken"`
	CreatedAt   int64  `json:"createdAt"`
	ModifiedAt  int64  `json:"modifiedAt"`
	CurrentTime int64  `json:"currentTime"`
}

// Online handles device online-status lookups via the external online service.
type Online struct {
	client          *mongo.Client
	collDevices     *mongo.Collection
	collProfiles    *mongo.Collection
	logger          *zap.SugaredLogger
	onlineByUUIDURL string
}

// NewOnline constructs an Online handler with the given dependencies.
func NewOnline(logger *zap.SugaredLogger, client *mongo.Client) *Online {
	onlineServerURL := os.Getenv("HTTP_ONLINE_SERVER") + ":" + os.Getenv("HTTP_ONLINE_PORT")
	onlineByUUIDURL := onlineServerURL + os.Getenv("HTTP_ONLINE_API")

	return &Online{
		client:          client,
		collDevices:     db.GetCollections(client).Devices,
		collProfiles:    db.GetCollections(client).Profiles,
		logger:          logger,
		onlineByUUIDURL: onlineByUUIDURL,
	}
}

// GetOnline function
func (o *Online) GetOnline(c *gin.Context) {
	o.logger.Info("REST - GET - GetOnline called")

	objectID, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		o.logger.Error("REST - GET - GetOnline - wrong format of the path param 'id'")
		c.JSON(http.StatusBadRequest, gin.H{"error": "wrong format of the path param 'id'"})
		return
	}

	// retrieve current profile object from database using the authenticated context
	profile, err := utils.GetLoggedProfileFromContext(c, o.collProfiles)
	if err != nil {
		o.logger.Error("REST - GET - GetOnline - cannot find profile")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "cannot find profile"})
		return
	}

	// check if device is in profile (device owned by profile)
	if !utils.Contains(profile.Devices, objectID) {
		o.logger.Error("REST - GET - GetOnline - this device is not in your profile")
		c.JSON(http.StatusBadRequest, gin.H{"error": "this device is not in your profile"})
		return
	}
	// get device from db
	device, err := o.getDevice(c.Request.Context(), objectID)
	if err != nil {
		o.logger.Error("REST - GET - GetOnline - cannot find device")
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot find device"})
		return
	}
	// get online feature of device from db
	onlineFeature := utils.GetOnlineFeature(device.Features)
	if onlineFeature == nil {
		o.logger.Error("REST - GET - GetOnline - cannot find online feature in this device")
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot find online feature in this device"})
		return
	}

	if !utils.IsValidUUID(device.UUID) || !utils.IsValidUUID(onlineFeature.UUID) {
		o.logger.Error("REST - GET - GetOnline - invalid UUID format in device or feature")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot get online"})
		return
	}
	path := o.onlineByUUIDURL + url.PathEscape(device.UUID) + "/features/" + url.PathEscape(onlineFeature.UUID)
	o.logger.Debugf("REST - GET - GetOnline - calling external 'online' service = %s", path)
	_, result, err := o.onlineByUUIDService(path)
	if err != nil {
		o.logger.Errorf("REST - GetOnline - cannot get online from remote service = %#v", err)
		if re, ok := err.(*customerrors.ErrorWrapper); ok {
			o.logger.Errorf("REST - GetOnline - cannot get online with status = %d, message = %s\n", re.Code, re.Message)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot get online"})
		return
	}
	o.logger.Debugf("REST - GetOnline - result = %#v", result)

	onlineResp := onlineResponse{}
	err = json.Unmarshal([]byte(result), &onlineResp)
	if err != nil {
		o.logger.Errorf("REST - GetOnline - cannot unmarshal JSON response from online remote service = %#v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot get online response"})
		return
	}
	o.logger.Debugf("REST - GetOnline - external 'online' service response = %#v", onlineResp)

	response := models.Online{}
	response.CreatedAt = time.UnixMilli(onlineResp.CreatedAt)
	response.ModifiedAt = time.UnixMilli(onlineResp.ModifiedAt)
	response.CurrentTime = time.UnixMilli(onlineResp.CurrentTime)
	c.JSON(http.StatusOK, &response)
}

func (o *Online) getDevice(ctx context.Context, deviceID bson.ObjectID) (models.Device, error) {
	o.logger.Debug("getDevice - searching device with objectId: ", deviceID)
	var device models.Device
	err := o.collDevices.FindOne(ctx, bson.M{
		"_id": deviceID,
	}).Decode(&device)
	o.logger.Debug("Device found: ", device)
	return device, err
}

func (o *Online) onlineByUUIDService(urlOnline string) (int, string, error) {
	return utils.Get(urlOnline)
}
