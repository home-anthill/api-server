package api

import (
	"api-server/customerrors"
	"api-server/db"
	"api-server/utils"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"
)

// InitFCMTokenReq is the request body for registering a Firebase Cloud Messaging token.
type InitFCMTokenReq struct {
	FCMToken string `json:"fcmToken" validate:"required,max=512"`
}

// OnlineFCMReq is the payload forwarded to the online service to associate an FCM token with an API token.
type OnlineFCMReq struct {
	APIToken string `json:"apiToken" validate:"required"`
	FCMToken string `json:"fcmToken" validate:"required,max=512"`
}

// FCMToken handles Firebase Cloud Messaging token registration for push notifications.
type FCMToken struct {
	client             *mongo.Client
	collProfiles       *mongo.Collection
	logger             *zap.SugaredLogger
	keepAliveOnlineURL string
	fcmTokenOnlineURL  string
	validate           *validator.Validate
}

// NewFCMToken constructs an FCMToken handler with the given dependencies.
func NewFCMToken(logger *zap.SugaredLogger, client *mongo.Client, validate *validator.Validate) *FCMToken {
	onlineServerURL := os.Getenv("HTTP_ONLINE_SERVER") + ":" + os.Getenv("HTTP_ONLINE_PORT")
	keepAliveOnlineURL := onlineServerURL + os.Getenv("HTTP_ONLINE_KEEPALIVE_API")
	fcmTokenOnlineURL := onlineServerURL + os.Getenv("HTTP_ONLINE_FCMTOKEN_API")

	return &FCMToken{
		client:             client,
		collProfiles:       db.GetCollections(client).Profiles,
		logger:             logger,
		keepAliveOnlineURL: keepAliveOnlineURL,
		fcmTokenOnlineURL:  fcmTokenOnlineURL,
		validate:           validate,
	}
}

// PostFCMToken function to associate smartphone app with Firebase client to this server via APIToken
// This will be sent to online server to store that data in Redis to be able to send Push Notifications
func (ft *FCMToken) PostFCMToken(c *gin.Context) {
	ft.logger.Info("REST - POST - PostFCMToken called")

	var initFCMTokenBody InitFCMTokenReq
	if err := c.ShouldBindJSON(&initFCMTokenBody); err != nil {
		ft.logger.Errorf("REST - POST - PostFCMToken - Cannot bind request body. Err = %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	err := ft.validate.Struct(initFCMTokenBody)
	if err != nil {
		ft.logger.Errorf("REST - POST - PostFCMToken - request body is not valid, err %#v", err)
		var errFields = utils.GetErrorMessage(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body, these fields are not valid:" + errFields})
		return
	}

	// retrieve current profile object from database (using session profile as input)
	session := sessions.Default(c)
	profile, err := utils.GetLoggedProfile(c.Request.Context(), &session, ft.collProfiles)
	if err != nil {
		ft.logger.Error("REST - POST - PostFCMToken - cannot find profile in session")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "cannot find profile in session"})
		return
	}

	// store FCM Token also on profile
	_, err = ft.collProfiles.UpdateOne(c.Request.Context(), bson.M{
		"_id": profile.ID,
	}, bson.M{
		"$set": bson.M{
			"fcmToken":   initFCMTokenBody.FCMToken,
			"modifiedAt": time.Now(),
		},
	})
	if err != nil {
		ft.logger.Error("REST - POST - PostFCMToken - Cannot update profile with fcmToken")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot update profile with fcmToken"})
		return
	}

	var onlineFCMReq = OnlineFCMReq{
		APIToken: profile.APIToken,
		FCMToken: initFCMTokenBody.FCMToken,
	}
	err = ft.initFCMTokenViaHTTP(&onlineFCMReq)
	if err != nil {
		ft.logger.Errorf("REST - POST - PostFCMToken - cannot initialize FCM Token via HTTP. Err %v\n", err)
		if re, ok := err.(*customerrors.ErrorWrapper); ok {
			ft.logger.Errorf("REST - POST - PostFCMToken - cannot initialize FCM Token with status = %d, message = %s\n", re.Code, re.Message)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot initialize FCM Token"})
		return
	}
	ft.logger.Infow("AUDIT - FCM token registered",
		"profileID", profile.ID.Hex(),
		"clientIP", c.ClientIP(),
	)
	c.JSON(http.StatusOK, gin.H{"message": "FCMToken assigned to APIToken"})
}

func (ft *FCMToken) initFCMTokenViaHTTP(obj *OnlineFCMReq) error {
	// check if service is available calling keep-alive
	_, _, keepAliveErr := utils.Get(ft.keepAliveOnlineURL)
	if keepAliveErr != nil {
		return customerrors.Wrap(http.StatusInternalServerError, keepAliveErr, "Cannot call keepAlive of remote online service")
	}

	// do the real call to the remote online service
	payloadJSON, err := json.Marshal(obj)
	if err != nil {
		return customerrors.Wrap(http.StatusInternalServerError, err, "Cannot create payload to call fcmToken service")
	}

	_, _, err = utils.Post(ft.fcmTokenOnlineURL, payloadJSON)
	if err != nil {
		return customerrors.Wrap(http.StatusInternalServerError, err, "Cannot init fcmToken")
	}
	return nil
}
