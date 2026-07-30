package api

import (
	"api-server/customerrors"
	"api-server/db"
	"api-server/models"
	"api-server/utils"
	"context"
	"encoding/json"
	"errors"
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

// Online handles device online-status lookups via the external alarm service.
type Online struct {
	client          *mongo.Client
	collDevices     *mongo.Collection
	collProfiles    *mongo.Collection
	logger          *zap.SugaredLogger
	onlineByUUIDURL string
}

// NewOnline constructs an Online handler with the given dependencies.
func NewOnline(logger *zap.SugaredLogger, client *mongo.Client) *Online {
	onlineServerURL := os.Getenv("HTTP_ALARM_SERVER") + ":" + os.Getenv("HTTP_ALARM_PORT")
	onlineByUUIDURL := onlineServerURL + os.Getenv("HTTP_ALARM_ONLINE_API")

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
	onlineResp, err := o.getOnlineByDeviceFeature(device.UUID, onlineFeature.UUID)
	if err != nil {
		if re, ok := asErrorWrapper(err); ok {
			o.logger.Errorf("REST - GetOnline - cannot get online from remote service = %#v", err)
			o.logger.Errorf("REST - GetOnline - cannot get online with status = %d, message = %s\n", re.Code, re.Message)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot get online"})
			return
		}

		o.logger.Errorf("REST - GetOnline - cannot unmarshal JSON response from online remote service = %#v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot get online response"})
		return
	}
	o.logger.Debugf("REST - GetOnline - external 'alarm' service response = %#v", onlineResp)

	response := models.Online{}
	response.CreatedAt = time.UnixMilli(onlineResp.CreatedAt)
	response.ModifiedAt = time.UnixMilli(onlineResp.ModifiedAt)
	response.CurrentTime = time.UnixMilli(onlineResp.CurrentTime)
	c.JSON(http.StatusOK, &response)
}

// GetProfileOnline returns online statuses for all online-capable devices owned by the logged profile.
func (o *Online) GetProfileOnline(c *gin.Context) {
	o.logger.Info("REST - GET - GetProfileOnline called")

	profile, err := utils.GetLoggedProfileFromContext(c, o.collProfiles)
	if err != nil {
		o.logger.Error("REST - GET - GetProfileOnline - cannot find profile")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "cannot find profile"})
		return
	}

	devices, err := o.getProfileDevices(c.Request.Context(), profile.Devices)
	if err != nil {
		o.logger.Errorf("REST - GET - GetProfileOnline - cannot get profile devices = %#v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot get online"})
		return
	}

	response := make([]models.OnlineDeviceStatus, 0, len(devices))
	for _, device := range devices {
		onlineFeature := utils.GetOnlineFeature(device.Features)
		if onlineFeature == nil {
			continue
		}

		if !utils.IsValidUUID(device.UUID) || !utils.IsValidUUID(onlineFeature.UUID) {
			o.logger.Error("REST - GET - GetProfileOnline - invalid UUID format in device or feature")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot get online"})
			return
		}

		onlineResp, err := o.getOnlineByDeviceFeature(device.UUID, onlineFeature.UUID)
		if err != nil {
			o.logger.Errorf("REST - GET - GetProfileOnline - cannot get online from remote service = %#v", err)
			if re, ok := asErrorWrapper(err); ok {
				o.logger.Errorf("REST - GET - GetProfileOnline - cannot get online with status = %d, message = %s\n", re.Code, re.Message)
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot get online"})
			return
		}

		response = append(response, models.OnlineDeviceStatus{
			CreatedAt:   time.UnixMilli(onlineResp.CreatedAt),
			ModifiedAt:  time.UnixMilli(onlineResp.ModifiedAt),
			CurrentTime: time.UnixMilli(onlineResp.CurrentTime),
			Device:      device,
			Feature:     *onlineFeature,
		})
	}

	c.JSON(http.StatusOK, response)
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

func (o *Online) getProfileDevices(ctx context.Context, deviceIDs []bson.ObjectID) ([]models.Device, error) {
	devices := make([]models.Device, 0)
	if len(deviceIDs) == 0 {
		return devices, nil
	}

	cursor, err := o.collDevices.Find(ctx, bson.M{"_id": bson.M{"$in": deviceIDs}})
	if err != nil {
		return devices, err
	}
	defer func() {
		if closeErr := cursor.Close(ctx); closeErr != nil {
			o.logger.Errorw("getProfileDevices - cannot close cursor", "error", closeErr)
		}
	}()

	if err = cursor.All(ctx, &devices); err != nil {
		return devices, err
	}

	return devices, nil
}

func (o *Online) getOnlineByDeviceFeature(deviceUUID, featureUUID string) (onlineResponse, error) {
	path := o.onlineByUUIDURL + url.PathEscape(deviceUUID) + "/features/" + url.PathEscape(featureUUID)
	o.logger.Debugf("getOnlineByDeviceFeature - calling external 'alarm' service = %s", path)

	_, result, err := o.onlineByUUIDService(path)
	if err != nil {
		return onlineResponse{}, err
	}

	onlineResp := onlineResponse{}
	if err = json.Unmarshal([]byte(result), &onlineResp); err != nil {
		return onlineResponse{}, err
	}

	return onlineResp, nil
}

func (o *Online) onlineByUUIDService(urlOnline string) (int, string, error) {
	return utils.Get(urlOnline)
}

func asErrorWrapper(err error) (customerrors.ErrorWrapper, bool) {
	var wrapper customerrors.ErrorWrapper
	if errors.As(err, &wrapper) {
		return wrapper, true
	}

	var wrapperPtr *customerrors.ErrorWrapper
	if errors.As(err, &wrapperPtr) && wrapperPtr != nil {
		return *wrapperPtr, true
	}

	return customerrors.ErrorWrapper{}, false
}
