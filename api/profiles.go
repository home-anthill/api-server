package api

import (
	"api-server/customerrors"
	"api-server/db"
	"api-server/models"
	"api-server/utils"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"
)

// ProfileUpdateFCMTokenReq is the request body for updating a profile's FCM token.
type ProfileUpdateFCMTokenReq struct {
	FCMToken string `json:"fcmToken" validate:"required,max=512"`
}

type rotateOnlineAPITokenReq struct {
	OldAPIToken    string                   `json:"oldApiToken"`
	NewAPIToken    string                   `json:"newApiToken"`
	DeviceFeatures []rotateOnlineDeviceFeat `json:"deviceFeatures"`
}

type rotateOnlineDeviceFeat struct {
	DeviceUUID  string `json:"deviceUuid"`
	FeatureUUID string `json:"featureUuid"`
}

// GithubResponse is the GitHub user data returned in a profile response.
type GithubResponse struct {
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatarURL"`
}

// Profiles handles user profile retrieval and token management.
type Profiles struct {
	client                  *mongo.Client
	collProfiles            *mongo.Collection
	collDevices             *mongo.Collection
	collSensors             *mongo.Collection
	collControls            *mongo.Collection
	onlineKeepAliveURL      string
	onlineRotateAPITokenURL string
	logger                  *zap.SugaredLogger
	validate                *validator.Validate
}

// NewProfiles constructs a Profiles handler with the given dependencies.
func NewProfiles(logger *zap.SugaredLogger, client *mongo.Client, validate *validator.Validate) *Profiles {
	onlineServerURL := os.Getenv("HTTP_ONLINE_SERVER") + ":" + os.Getenv("HTTP_ONLINE_PORT")
	return &Profiles{
		client:                  client,
		collProfiles:            db.GetCollections(client).Profiles,
		collDevices:             db.GetCollections(client).Devices,
		collSensors:             client.Database(sensorDbName()).Collection("sensors"),
		collControls:            client.Database(controllerDbName()).Collection("controllers"),
		onlineKeepAliveURL:      onlineServerURL + os.Getenv("HTTP_ONLINE_KEEPALIVE_API"),
		onlineRotateAPITokenURL: onlineServerURL + os.Getenv("HTTP_ONLINE_ROTATE_APITOKEN_API"),
		logger:                  logger,
		validate:                validate,
	}
}

// GetProfile function
func (p *Profiles) GetProfile(c *gin.Context) {
	p.logger.Info("REST - GET - GetProfile called")

	profile, err := utils.GetLoggedProfileFromContext(c, p.collProfiles)
	if err != nil {
		p.logger.Error("REST - GET - GetProfile - Cannot get user profile")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Cannot get user profile"})
		return
	}

	// build profile response re-using the Profile model with only the desired fields
	profileRes := models.Profile{}
	profileRes.ID = profile.ID
	profileRes.CreatedAt = profile.CreatedAt
	profileRes.ModifiedAt = profile.ModifiedAt
	profileRes.Github = profile.Github
	c.JSON(http.StatusOK, &profileRes)
}

// PostRotateAPIToken regenerates the API token for the logged-in profile.
func (p *Profiles) PostRotateAPIToken(c *gin.Context) {
	p.logger.Info("REST - POST - PostRotateAPIToken called")

	// get profileID from path params
	profileID, errID := bson.ObjectIDFromHex(c.Param("id"))
	if errID != nil {
		p.logger.Error("REST - POST - PostRotateAPIToken - wrong format of the path param 'id'")
		c.JSON(http.StatusBadRequest, gin.H{"error": "wrong format of the path param 'id'"})
		return
	}

	// retrieve current profile identity from the authenticated context
	profileSession, err := utils.GetProfileFromContext(c)
	if err != nil {
		p.logger.Error("REST - POST - PostRotateAPIToken - cannot find profile")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "cannot find profile"})
		return
	}

	// check if the profile you are trying to update (path param) is your profile (session profile)
	if profileSession.ID != profileID {
		p.logger.Error("REST - POST - PostRotateAPIToken - Current profileID is different than profileID in session")
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot re-generate APIToken for a different profile then yours"})
		return
	}
	var profile models.Profile
	if findErr := p.collProfiles.FindOne(c.Request.Context(), bson.M{"_id": profileSession.ID}).Decode(&profile); findErr != nil {
		p.logger.Error("REST - POST - PostRotateAPIToken - cannot find profile")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "cannot find profile"})
		return
	}
	oldAPIToken, err := decryptProfileAPIToken(&profile)
	if err != nil {
		p.logger.Error("REST - POST - PostRotateAPIToken - Cannot decrypt current apiToken")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot update apiToken"})
		return
	}

	newAPIToken := uuid.NewString()
	newAPITokenEncrypted, err := utils.EncryptAPIToken(newAPIToken)
	if err != nil {
		p.logger.Error("REST - POST - PostRotateAPIToken - Cannot encrypt new apiToken")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot update apiToken"})
		return
	}
	newAPITokenHash, err := utils.HashAPIToken(newAPIToken)
	if err != nil {
		p.logger.Error("REST - POST - PostRotateAPIToken - Cannot hash newAPIToken")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot update apiToken"})
		return
	}

	if err = p.rotateProfileAndDeviceTokens(c.Request.Context(), profileSession.ID, newAPITokenHash, newAPITokenEncrypted); err != nil {
		p.logger.Error("REST - POST - PostRotateAPIToken - Cannot update profile with the new apiToken")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot update apiToken"})
		return
	}
	onlineDeviceFeatures, err := p.getProfileOnlineDeviceFeatures(c.Request.Context(), profile)
	if err != nil {
		p.logger.Errorw("REST - POST - PostRotateAPIToken - Cannot build online apiToken rotation targets", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot update apiToken"})
		return
	}
	if err = p.rotateOnlineAPIToken(oldAPIToken, newAPIToken, onlineDeviceFeatures); err != nil {
		p.logger.Errorw("REST - POST - PostRotateAPIToken - Cannot rotate apiToken in online service", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot update apiToken"})
		return
	}
	p.logger.Infow("AUDIT - API token regenerated",
		"profileID", profileSession.ID.Hex(),
	)
	c.JSON(http.StatusOK, gin.H{"apiToken": newAPIToken})
}

func (p *Profiles) rotateProfileAndDeviceTokens(ctx context.Context, profileID bson.ObjectID, apiTokenHash, newAPITokenEncrypted string) error {
	dbSession, err := p.client.StartSession()
	if err != nil {
		return err
	}
	defer dbSession.EndSession(ctx)

	now := time.Now().UTC()
	_, err = dbSession.WithTransaction(ctx, func(sessionCtx context.Context) (interface{}, error) {
		credentialUpdate := bson.M{
			"$set": bson.M{
				"apiTokenHash":      apiTokenHash,
				"apiTokenEncrypted": newAPITokenEncrypted,
			},
		}
		if _, updateErr := p.collSensors.UpdateMany(sessionCtx, bson.M{"profileOwnerId": profileID}, credentialUpdate); updateErr != nil {
			p.logger.Errorw("PostRotateAPIToken - Cannot update sensor apiToken credentials", "error", updateErr)
			return nil, updateErr
		}
		if _, updateErr := p.collControls.UpdateMany(sessionCtx, bson.M{"profileOwnerId": profileID}, credentialUpdate); updateErr != nil {
			p.logger.Errorw("PostRotateAPIToken - Cannot update controller apiToken credentials", "error", updateErr)
			return nil, updateErr
		}
		if _, updateErr := p.collProfiles.UpdateOne(sessionCtx, bson.M{
			"_id": profileID,
		}, bson.M{
			"$set": bson.M{
				"apiTokenHash":      apiTokenHash,
				"apiTokenEncrypted": newAPITokenEncrypted,
				"modifiedAt":        now,
			},
		}); updateErr != nil {
			p.logger.Errorw("PostRotateAPIToken - Cannot update profile apiToken credentials", "error", updateErr)
			return nil, updateErr
		}
		return nil, nil
	})
	return err
}

func (p *Profiles) getProfileOnlineDeviceFeatures(ctx context.Context, profile models.Profile) ([]rotateOnlineDeviceFeat, error) {
	if len(profile.Devices) == 0 {
		return []rotateOnlineDeviceFeat{}, nil
	}

	cursor, err := p.collDevices.Find(ctx, bson.M{"_id": bson.M{"$in": profile.Devices}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var devices []models.Device
	if err = cursor.All(ctx, &devices); err != nil {
		return nil, err
	}

	deviceFeatures := make([]rotateOnlineDeviceFeat, 0)
	for _, device := range devices {
		for _, feature := range device.Features {
			deviceFeatures = append(deviceFeatures, rotateOnlineDeviceFeat{
				DeviceUUID:  device.UUID,
				FeatureUUID: feature.UUID,
			})
		}
	}
	return deviceFeatures, nil
}

func (p *Profiles) rotateOnlineAPIToken(oldAPIToken, newAPIToken string, deviceFeatures []rotateOnlineDeviceFeat) error {
	_, _, keepAliveErr := utils.Get(p.onlineKeepAliveURL)
	if keepAliveErr != nil {
		return customerrors.Wrap(http.StatusInternalServerError, keepAliveErr, "Cannot call keepAlive of remote online service")
	}

	payloadJSON, err := json.Marshal(rotateOnlineAPITokenReq{
		OldAPIToken:    oldAPIToken,
		NewAPIToken:    newAPIToken,
		DeviceFeatures: deviceFeatures,
	})
	if err != nil {
		return customerrors.Wrap(http.StatusInternalServerError, err, "Cannot create payload to rotate online apiToken")
	}

	_, _, err = utils.Post(p.onlineRotateAPITokenURL, payloadJSON)
	if err != nil {
		return customerrors.Wrap(http.StatusInternalServerError, err, "Cannot rotate online apiToken")
	}
	return nil
}

func sensorDbName() string {
	if os.Getenv("ENV") == "testing" {
		return "sensors_test"
	}
	return "sensors"
}

func controllerDbName() string {
	if os.Getenv("ENV") == "testing" {
		return "controllers-test"
	}
	return "controllers"
}

// PostProfilesFCMToken function to store the Firebase Cloud Messaging Token
// this api is unused, because I set FCM Token on profile while calling fcm_token POST API
func (p *Profiles) PostProfilesFCMToken(c *gin.Context) {
	p.logger.Info("REST - POST - PostProfilesFCMToken called")

	// get profileID from path params
	profileID, errID := bson.ObjectIDFromHex(c.Param("id"))
	if errID != nil {
		p.logger.Error("REST - POST - PostProfilesFCMToken - wrong format of the path param 'id'")
		c.JSON(http.StatusBadRequest, gin.H{"error": "wrong format of the path param 'id'"})
		return
	}

	// retrieve current profile identity from the authenticated context
	profileSession, err := utils.GetProfileFromContext(c)
	if err != nil {
		p.logger.Error("REST - POST - PostProfilesFCMToken - cannot find profile")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "cannot find profile"})
		return
	}

	// check if the profile you are trying to update (path param) is your profile (session profile)
	if profileSession.ID != profileID {
		p.logger.Error("REST - POST - PostProfilesFCMToken - Current profileID is different than profileID in session")
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot set FCMToken for a different profile then yours"})
		return
	}

	var profileUpdateFCMTokenReq ProfileUpdateFCMTokenReq
	if err = c.ShouldBindJSON(&profileUpdateFCMTokenReq); err != nil {
		p.logger.Error("REST - POST - PostProfilesFCMToken - Cannot bind request body", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	err = p.validate.Struct(profileUpdateFCMTokenReq)
	if err != nil {
		p.logger.Errorf("REST - POST - PostProfilesFCMToken - request body is not valid, err %#v", err)
		var errFields = utils.GetErrorMessage(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body, these fields are not valid:" + errFields})
		return
	}

	_, err = p.collProfiles.UpdateOne(c.Request.Context(), bson.M{
		"_id": profileSession.ID,
	}, bson.M{
		"$set": bson.M{
			"fcmToken":   profileUpdateFCMTokenReq.FCMToken,
			"modifiedAt": time.Now(),
		},
	})
	if err != nil {
		p.logger.Error("REST - POST - PostProfilesFCMToken - Cannot update profile with fcmToken")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot set fcmToken"})
		return
	}
	p.logger.Infow("AUDIT - FCM token updated on profile",
		"profileID", profileSession.ID.Hex(),
	)
	c.JSON(http.StatusOK, gin.H{"message": "Profile update with FCM Token"})
}
