package api

import (
	"api-server/customerrors"
	"api-server/db"
	"api-server/models"
	"api-server/utils"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"net/http"
	"os"
)

// InitFCMTokenReq struct
type InitFCMTokenReq struct {
	APIToken string `json:"apiToken" validate:"required,uuid4"`
	FCMToken string `json:"fcmToken" validate:"required"`
}

// FCMToken struct
type FCMToken struct {
	client             *mongo.Client
	collProfiles       *mongo.Collection
	ctx                context.Context
	logger             *zap.SugaredLogger
	keepAliveOnlineURL string
	fcmTokenOnlineURL  string
	validate           *validator.Validate
}

// NewFCMToken function
func NewFCMToken(ctx context.Context, logger *zap.SugaredLogger, client *mongo.Client, validate *validator.Validate) *FCMToken {
	onlineServerURL := os.Getenv("HTTP_ONLINE_SERVER") + ":" + os.Getenv("HTTP_ONLINE_PORT")
	keepAliveOnlineURL := onlineServerURL + os.Getenv("HTTP_ONLINE_KEEPALIVE_API")
	fcmTokenOnlineURL := onlineServerURL + os.Getenv("HTTP_ONLINE_FCMTOKEN_API")

	return &FCMToken{
		client:             client,
		collProfiles:       db.GetCollections(client).Profiles,
		ctx:                ctx,
		logger:             logger,
		keepAliveOnlineURL: keepAliveOnlineURL,
		fcmTokenOnlineURL:  fcmTokenOnlineURL,
		validate:           validate,
	}
}

// PostFCMToken function to associate smartphone app with Firebase client to this server via APIToken
// This will be sent to online server to store that data in Redis to be able to send Push Notifications
func (handler *FCMToken) PostFCMToken(c *gin.Context) {
	handler.logger.Info("REST - PostFCMToken called")

	var initFCMTokenBody InitFCMTokenReq
	if err := c.ShouldBindJSON(&initFCMTokenBody); err != nil {
		handler.logger.Errorf("REST - PostFCMToken - Cannot bind request body. Err = %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	err := handler.validate.Struct(initFCMTokenBody)
	if err != nil {
		handler.logger.Errorf("REST - PostFCMToken - request body is not valid, err %#v", err)
		var errFields = utils.GetErrorMessage(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body, these fields are not valid:" + errFields})
		return
	}

	// search if profile token exists and retrieve profile
	var profileFound models.Profile
	errProfile := handler.collProfiles.FindOne(handler.ctx, bson.M{
		"apiToken": initFCMTokenBody.APIToken,
	}).Decode(&profileFound)
	if errProfile != nil {
		handler.logger.Errorf("REST - PostFCMToken - Cannot find profile with that apiToken. Err = %v\n", errProfile)
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot initialize FCM Token, profile token missing or not valid"})
		return
	}

	err = handler.initFCMTokenViaHTTP(&initFCMTokenBody)
	if err != nil {
		handler.logger.Errorf("REST - PostFCMToken - cannot initialize FCM Token via HTTP. Err %v\n", err)
		if re, ok := err.(*customerrors.ErrorWrapper); ok {
			handler.logger.Errorf("REST - PostFCMToken - cannot initialize FCM Token with status = %d, message = %s\n", re.Code, re.Message)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot initialize FCM Token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "FCMToken assigned to APIToken"})
}

func (handler *FCMToken) initFCMTokenViaHTTP(obj *InitFCMTokenReq) error {
	// check if service is available calling keep-alive
	// TODO remove this in a production code
	_, _, keepAliveErr := utils.Get(handler.keepAliveOnlineURL)
	if keepAliveErr != nil {
		return customerrors.Wrap(http.StatusInternalServerError, keepAliveErr, "Cannot call keepAlive of remote online service")
	}

	// do the real call to the remote online service
	payloadJSON, err := json.Marshal(obj)
	if err != nil {
		return customerrors.Wrap(http.StatusInternalServerError, err, "Cannot create payload to call fcmToken service")
	}

	_, _, err = utils.Post(handler.fcmTokenOnlineURL, payloadJSON)
	if err != nil {
		return customerrors.Wrap(http.StatusInternalServerError, err, "Cannot init fcmToken")
	}
	return nil
}
