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
	"strings"
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

const offlineThreshold = time.Minute

type onlineBulkDeviceFeature struct {
	DeviceUUID  string `json:"deviceUuid"`
	FeatureUUID string `json:"featureUuid"`
}

type onlineBulkRequest struct {
	DeviceFeatures []onlineBulkDeviceFeature `json:"deviceFeatures"`
}

type onlineBulkStatus struct {
	DeviceUUID  string `json:"deviceUuid"`
	FeatureUUID string `json:"featureUuid"`
	Status      string `json:"status"`
	CreatedAt   *int64 `json:"createdAt"`
	ModifiedAt  *int64 `json:"modifiedAt"`
}

type onlineBulkResponse struct {
	Statuses    []onlineBulkStatus `json:"statuses"`
	CurrentTime int64              `json:"currentTime"`
}

type onlineStatusCandidate struct {
	deviceID    string
	deviceUUID  string
	featureUUID string
}

// Online handles device online-status lookups via the external alarm-api service.
type Online struct {
	client       *mongo.Client
	collDevices  *mongo.Collection
	collProfiles *mongo.Collection
	logger       *zap.SugaredLogger
	onlineURL    string
}

// NewOnline constructs an Online handler with the given dependencies.
func NewOnline(logger *zap.SugaredLogger, client *mongo.Client) *Online {
	onlineServerURL := os.Getenv("HTTP_ALARM_SERVER") + ":" + os.Getenv("HTTP_ALARM_PORT")
	onlineURL := onlineServerURL + os.Getenv("HTTP_ALARM_ONLINE_API")

	return &Online{
		client:       client,
		collDevices:  db.GetCollections(client).Devices,
		collProfiles: db.GetCollections(client).Profiles,
		logger:       logger,
		onlineURL:    onlineURL,
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
	o.logger.Debugf("REST - GetOnline - external 'alarm-api' service response = %#v", onlineResp)

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
	candidates := make([]onlineStatusCandidate, 0, len(devices))
	request := onlineBulkRequest{DeviceFeatures: make([]onlineBulkDeviceFeature, 0, len(devices))}
	fallbackCurrentTime := time.Now().UTC()
	for _, device := range devices {
		// filter only devices that have 'online feature' enabled
		onlineFeature := utils.GetEnabledOnlineFeature(device.Features)
		if onlineFeature == nil {
			continue
		}

		if !utils.IsValidUUID(device.UUID) || !utils.IsValidUUID(onlineFeature.UUID) {
			o.logger.Errorw(
				"REST - GET - GetProfileOnline - invalid UUID format in device or feature",
				"deviceID", device.ID.Hex(),
			)
			response = append(response, models.OnlineDeviceStatus{
				DeviceID:    device.ID.Hex(),
				FeatureUUID: onlineFeature.UUID,
				Status:      models.OnlineStatusUnknown,
				CurrentTime: fallbackCurrentTime,
			})
			continue
		}

		// candidates because those devices are only eligible to be checked—not yet known to be online
		candidates = append(candidates, onlineStatusCandidate{
			deviceID:    device.ID.Hex(),
			deviceUUID:  device.UUID,
			featureUUID: onlineFeature.UUID,
		})
		request.DeviceFeatures = append(request.DeviceFeatures, onlineBulkDeviceFeature{
			DeviceUUID:  device.UUID,
			FeatureUUID: onlineFeature.UUID,
		})
	}

	if len(candidates) == 0 {
		c.JSON(http.StatusOK, response)
		return
	}

	bulkResponse, err := o.getOnlineBulk(request)
	if err != nil {
		o.logger.Errorf("REST - GET - GetProfileOnline - cannot get bulk online statuses = %#v", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "Cannot get online"})
		return
	}
	if bulkResponse.CurrentTime <= 0 {
		o.logger.Error("REST - GET - GetProfileOnline - invalid currentTime in bulk response")
		c.JSON(http.StatusBadGateway, gin.H{"error": "Cannot get online"})
		return
	}

	currentTime := time.UnixMilli(bulkResponse.CurrentTime)
	statusesByKey := make(map[string]onlineBulkStatus, len(bulkResponse.Statuses))
	for _, status := range bulkResponse.Statuses {
		statusesByKey[onlineStatusKey(status.DeviceUUID, status.FeatureUUID)] = status
	}

	for _, candidate := range candidates {
		status, found := statusesByKey[onlineStatusKey(candidate.deviceUUID, candidate.featureUUID)]
		result := models.OnlineDeviceStatus{
			DeviceID:    candidate.deviceID,
			FeatureUUID: candidate.featureUUID,
			Status:      models.OnlineStatusUnknown,
			CurrentTime: currentTime,
		}

		if found && status.Status == "missing" {
			result.Status = models.OnlineStatusOffline
		} else if found && status.Status == "found" && status.CreatedAt != nil && status.ModifiedAt != nil {
			createdAt := time.UnixMilli(*status.CreatedAt)
			modifiedAt := time.UnixMilli(*status.ModifiedAt)
			result.CreatedAt = &createdAt
			result.ModifiedAt = &modifiedAt
			result.Status = models.OnlineStatusOnline
			if modifiedAt.Before(currentTime.Add(-offlineThreshold)) {
				result.Status = models.OnlineStatusOffline
			}
		}

		response = append(response, result)
	}

	c.JSON(http.StatusOK, response)
}

func onlineStatusKey(deviceUUID, featureUUID string) string {
	return deviceUUID + ":" + featureUUID
}

func (o *Online) getOnlineBulk(request onlineBulkRequest) (onlineBulkResponse, error) {
	payload, err := json.Marshal(request)
	if err != nil {
		return onlineBulkResponse{}, err
	}

	_, result, err := utils.Post(strings.TrimSuffix(o.onlineURL, "/")+"/bulk", payload)
	if err != nil {
		return onlineBulkResponse{}, err
	}

	response := onlineBulkResponse{}
	if err = json.Unmarshal([]byte(result), &response); err != nil {
		return onlineBulkResponse{}, err
	}
	return response, nil
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
	path := o.onlineURL + url.PathEscape(deviceUUID) + "/features/" + url.PathEscape(featureUUID)
	o.logger.Debugf("getOnlineByDeviceFeature - calling external 'alarm-api' service = %s", path)

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
