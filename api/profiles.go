package api

import (
	"api-server/db"
	"api-server/models"
	"api-server/utils"
	"net/http"
	"time"

	"github.com/gin-contrib/sessions"
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

// GithubResponse is the GitHub user data returned in a profile response.
type GithubResponse struct {
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatarURL"`
}

// Profiles handles user profile retrieval and token management.
type Profiles struct {
	client       *mongo.Client
	collProfiles *mongo.Collection
	logger       *zap.SugaredLogger
	validate     *validator.Validate
}

// NewProfiles constructs a Profiles handler with the given dependencies.
func NewProfiles(logger *zap.SugaredLogger, client *mongo.Client, validate *validator.Validate) *Profiles {
	return &Profiles{
		client:       client,
		collProfiles: db.GetCollections(client).Profiles,
		logger:       logger,
		validate:     validate,
	}
}

// GetProfile function
func (p *Profiles) GetProfile(c *gin.Context) {
	p.logger.Info("REST - GET - GetProfile called")

	session := sessions.Default(c)
	profile, err := utils.GetProfileFromSession(&session)
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

// PostProfilesAPIToken function to regenerate the API Token
func (p *Profiles) PostProfilesAPIToken(c *gin.Context) {
	p.logger.Info("REST - POST - PostProfilesAPIToken called")

	// get profileID from path params
	profileID, errID := bson.ObjectIDFromHex(c.Param("id"))
	if errID != nil {
		p.logger.Error("REST - POST - PostProfilesAPIToken - wrong format of the path param 'id'")
		c.JSON(http.StatusBadRequest, gin.H{"error": "wrong format of the path param 'id'"})
		return
	}

	// retrieve current profile object from database (using session profile as input)
	session := sessions.Default(c)
	profileSession, err := utils.GetProfileFromSession(&session)
	if err != nil {
		p.logger.Error("REST - POST - PostProfilesAPIToken - cannot find profile in session")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "cannot find profile in session"})
		return
	}

	// check if the profile you are trying to update (path param) is your profile (session profile)
	if profileSession.ID != profileID {
		p.logger.Error("REST - POST - PostProfilesAPIToken - Current profileID is different than profileID in session")
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot re-generate APIToken for a different profile then yours"})
		return
	}

	apiToken := uuid.NewString()

	_, err = p.collProfiles.UpdateOne(c.Request.Context(), bson.M{
		"_id": profileSession.ID,
	}, bson.M{
		"$set": bson.M{
			"apiToken":   apiToken,
			"modifiedAt": time.Now(),
		},
	})
	if err != nil {
		p.logger.Error("REST - POST - PostProfilesAPIToken - Cannot update profile with the new apiToken")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot update apiToken"})
		return
	}
	p.logger.Infow("AUDIT - API token regenerated",
		"profileID", profileSession.ID.Hex(),
		"clientIP", c.ClientIP(),
	)
	c.JSON(http.StatusOK, gin.H{"apiToken": apiToken})
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

	// retrieve current profile object from database (using session profile as input)
	session := sessions.Default(c)
	profileSession, err := utils.GetProfileFromSession(&session)
	if err != nil {
		p.logger.Error("REST - POST - PostProfilesFCMToken - cannot find profile in session")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "cannot find profile in session"})
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
		"clientIP", c.ClientIP(),
	)
	c.JSON(http.StatusOK, gin.H{"message": "Profile update with FCM Token"})
}
