package api

import (
	"api-server/customerrors"
	"api-server/db"
	"api-server/models"
	"api-server/utils"
	"encoding/json"
	"net/http"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"
)

type OnlineNotificationsResponse struct {
	Notifications []OnlineNotification `json:"notifications"`
}

type OnlineNotification struct {
	ID                string                     `json:"id"`
	SentAt            uint64                     `json:"sentAt"`
	Title             string                     `json:"title"`
	Body              string                     `json:"body"`
	DeviceCount       uint64                     `json:"deviceCount"`
	Devices           []OnlineNotificationDevice `json:"devices"`
	Provider          string                     `json:"provider"`
	ProviderMessageID string                     `json:"providerMessageId"`
}

type OnlineNotificationDevice struct {
	DeviceUUID string `json:"deviceUuid"`
}

type NotificationsResponse struct {
	Notifications []Notification `json:"notifications"`
}

type Notification struct {
	ID                string          `json:"id"`
	SentAt            uint64          `json:"sentAt"`
	Title             string          `json:"title"`
	Body              string          `json:"body"`
	DeviceCount       uint64          `json:"deviceCount"`
	Devices           []models.Device `json:"devices"`
	Provider          string          `json:"provider"`
	ProviderMessageID string          `json:"providerMessageId"`
}

// Notifications handles notification-history lookups via the external alarm service.
type Notifications struct {
	client                 *mongo.Client
	collDevices            *mongo.Collection
	collProfiles           *mongo.Collection
	logger                 *zap.SugaredLogger
	onlineNotificationsURL string
}

// NewNotifications constructs a Notifications handler with the given dependencies.
func NewNotifications(logger *zap.SugaredLogger, client *mongo.Client) *Notifications {
	onlineServerURL := os.Getenv("HTTP_ALARM_SERVER") + ":" + os.Getenv("HTTP_ALARM_PORT")
	onlineNotificationsAPI := os.Getenv("HTTP_ALARM_NOTIFICATIONS_API")

	return &Notifications{
		client:                 client,
		collDevices:            db.GetCollections(client).Devices,
		collProfiles:           db.GetCollections(client).Profiles,
		logger:                 logger,
		onlineNotificationsURL: onlineServerURL + onlineNotificationsAPI,
	}
}

// GetNotifications returns notification history for the authenticated profile.
func (n *Notifications) GetNotifications(c *gin.Context) {
	n.logger.Info("REST - GET - GetNotifications called")

	profile, err := utils.GetLoggedProfileFromContext(c, n.collProfiles)
	if err != nil {
		n.logger.Error("REST - GET - GetNotifications - cannot find profile")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "cannot find profile"})
		return
	}

	apiToken, err := decryptProfileAPIToken(&profile)
	if err != nil {
		n.logger.Error("REST - GET - GetNotifications - cannot load profile api token")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot get notifications"})
		return
	}
	if !utils.IsValidUUID(apiToken) {
		n.logger.Error("REST - GET - GetNotifications - invalid profile api token format")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot get notifications"})
		return
	}

	path := n.onlineNotificationsURL + url.PathEscape(apiToken)
	n.logger.Debug("REST - GET - GetNotifications - calling external 'alarm' notifications service using apiToken")
	_, result, err := n.notificationsService(path)
	if err != nil {
		n.logger.Errorf("REST - GET - GetNotifications - cannot get notifications from remote service = %#v", err)
		if re, ok := err.(*customerrors.ErrorWrapper); ok {
			n.logger.Errorf("REST - GET - GetNotifications - remote status = %d, message = %s\n", re.Code, re.Message)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot get notifications"})
		return
	}

	var onlineResponse OnlineNotificationsResponse
	if err = json.Unmarshal([]byte(result), &onlineResponse); err != nil {
		n.logger.Errorf("REST - GET - GetNotifications - cannot unmarshal JSON response from online remote service = %#v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot get notifications response"})
		return
	}

	response, err := n.enrichNotifications(c, profile, onlineResponse)
	if err != nil {
		n.logger.Errorf("REST - GET - GetNotifications - cannot enrich notification devices = %#v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot get notifications response"})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (n *Notifications) enrichNotifications(c *gin.Context, profile models.Profile, onlineResponse OnlineNotificationsResponse) (NotificationsResponse, error) {
	deviceUUIDs := make([]string, 0)
	for _, item := range onlineResponse.Notifications {
		for _, device := range item.Devices {
			if device.DeviceUUID != "" {
				deviceUUIDs = append(deviceUUIDs, device.DeviceUUID)
			}
		}
	}

	devicesByUUID, err := n.getProfileDevicesByUUID(c, profile, deviceUUIDs)
	if err != nil {
		return NotificationsResponse{}, err
	}

	response := NotificationsResponse{
		Notifications: make([]Notification, 0, len(onlineResponse.Notifications)),
	}
	for _, item := range onlineResponse.Notifications {
		devices := make([]models.Device, 0, len(item.Devices))
		for _, onlineDevice := range item.Devices {
			device, ok := devicesByUUID[onlineDevice.DeviceUUID]
			if ok {
				devices = append(devices, device)
			}
		}

		response.Notifications = append(response.Notifications, Notification{
			ID:                item.ID,
			SentAt:            item.SentAt,
			Title:             item.Title,
			Body:              item.Body,
			DeviceCount:       item.DeviceCount,
			Devices:           devices,
			Provider:          item.Provider,
			ProviderMessageID: item.ProviderMessageID,
		})
	}
	return response, nil
}

func (n *Notifications) getProfileDevicesByUUID(c *gin.Context, profile models.Profile, deviceUUIDs []string) (map[string]models.Device, error) {
	devicesByUUID := make(map[string]models.Device)
	if len(deviceUUIDs) == 0 || len(profile.Devices) == 0 {
		return devicesByUUID, nil
	}

	cur, err := n.collDevices.Find(c.Request.Context(), bson.M{
		// resolve the device references returned by the alarm service notification payload
		"_id": bson.M{"$in": profile.Devices},
		// ensure we only return devices owned by the authenticated profile
		"uuid": bson.M{"$in": deviceUUIDs},
	})
	if err != nil {
		return nil, err
	}
	defer cur.Close(c.Request.Context())

	for cur.Next(c.Request.Context()) {
		var device models.Device
		if err = cur.Decode(&device); err != nil {
			return nil, err
		}
		devicesByUUID[device.UUID] = device
	}
	if err = cur.Err(); err != nil {
		return nil, err
	}
	return devicesByUUID, nil
}

func (n *Notifications) notificationsService(urlNotifications string) (int, string, error) {
	return utils.Get(urlNotifications)
}
