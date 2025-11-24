package api

import (
	"api-server/customerrors"
	"api-server/db"
	"api-server/models"
	"api-server/utils"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type onlineResponse struct {
	UUID        string `json:"uuid"`
	APIToken    string `json:"apiToken"`
	CreatedAt   int64  `json:"createdAt"`
	ModifiedAt  int64  `json:"modifiedAt"`
	CurrentTime int64  `json:"currentTime"`
}

// Online struct
type Online struct {
	client          *mongo.Client
	collDevices     *mongo.Collection
	collProfiles    *mongo.Collection
	ctx             context.Context
	logger          *zap.SugaredLogger
	onlineByUUIDURL string
}

// NewOnline function
func NewOnline(ctx context.Context, logger *zap.SugaredLogger, client *mongo.Client) *Online {
	onlineServerURL := os.Getenv("HTTP_ONLINE_SERVER") + ":" + os.Getenv("HTTP_ONLINE_PORT")
	onlineByUUIDURL := onlineServerURL + os.Getenv("HTTP_ONLINE_API")

	return &Online{
		client:          client,
		collDevices:     db.GetCollections(client).Devices,
		collProfiles:    db.GetCollections(client).Profiles,
		ctx:             ctx,
		logger:          logger,
		onlineByUUIDURL: onlineByUUIDURL,
	}
}

// GetOnline function
func (handler *Online) GetOnline(c *gin.Context) {
	handler.logger.Info("REST - GET - GetOnline called")

	objectID, errID := primitive.ObjectIDFromHex(c.Param("id"))
	if errID != nil {
		handler.logger.Error("REST - GET - GetOnline - wrong format of the path param 'id'")
		c.JSON(http.StatusBadRequest, gin.H{"error": "wrong format of the path param 'id'"})
		return
	}

	// retrieve current profile object from database (using session profile as input)
	session := sessions.Default(c)
	profile, err := utils.GetLoggedProfile(handler.ctx, &session, handler.collProfiles)
	if err != nil {
		handler.logger.Error("REST - GET - GetOnline - cannot find profile in session")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "cannot find profile in session"})
		return
	}

	// check if device is in profile (device owned by profile)
	if !utils.Contains(profile.Devices, objectID) {
		handler.logger.Error("REST - GET - GetOnline - this device is not in your profile")
		c.JSON(http.StatusBadRequest, gin.H{"error": "this device is not in your profile"})
		return
	}
	// get device from db
	device, err := handler.getDevice(objectID)
	if err != nil {
		handler.logger.Error("REST - GET - GetOnline - cannot find device")
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot find device"})
		return
	}
	// get online feature of device from db
	onlineFeature := utils.GetOnlineFeature(device.Features)
	if onlineFeature == nil {
		handler.logger.Error("REST - GET - GetOnline - cannot find online feature in this device")
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot find online feature in this device"})
		return
	}

	path := handler.onlineByUUIDURL + device.UUID + "/features/" + onlineFeature.UUID
	handler.logger.Debugf("REST - GET - GetOnline - calling external 'online' service = %s", path)
	_, result, err := handler.onlineByUUIDService(path)
	if err != nil {
		handler.logger.Errorf("REST - GetOnline - cannot get online from remote service = %#v", err)
		if re, ok := err.(*customerrors.ErrorWrapper); ok {
			handler.logger.Errorf("REST - GetOnline - cannot get online with status = %d, message = %s\n", re.Code, re.Message)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot get online"})
		return
	}
	handler.logger.Debugf("REST - GetOnline - result = %#v", result)

	onlineResp := onlineResponse{}
	err = json.Unmarshal([]byte(result), &onlineResp)
	if err != nil {
		handler.logger.Errorf("REST - GetOnline - cannot unmarshal JSON response from online remote service = %#v", err)
		// TODO manage errors
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot get online response"})
		return
	}
	handler.logger.Debugf("REST - GetOnline - external 'online' service response = %#v", onlineResp)

	response := models.Online{}
	response.CreatedAt = time.UnixMilli(onlineResp.CreatedAt)
	response.ModifiedAt = time.UnixMilli(onlineResp.ModifiedAt)
	response.CurrentTime = time.UnixMilli(onlineResp.CurrentTime)
	c.JSON(http.StatusOK, &response)
}

// TODO this is equals to the method defined in devices_values
func (handler *Online) getDevice(deviceID primitive.ObjectID) (models.Device, error) {
	handler.logger.Debug("getDevice - searching device with objectId: ", deviceID)
	var device models.Device
	err := handler.collDevices.FindOne(handler.ctx, bson.M{
		"_id": deviceID,
	}).Decode(&device)
	handler.logger.Debug("Device found: ", device)
	return device, err
}

func (handler *Online) onlineByUUIDService(urlOnline string) (int, string, error) {
	response, err := http.Get(urlOnline)
	if err != nil {
		return -1, "", customerrors.Wrap(http.StatusInternalServerError, err, "Cannot call online service via HTTP")
	}
	defer response.Body.Close()
	body, _ := io.ReadAll(response.Body)
	return response.StatusCode, string(body), nil
}
