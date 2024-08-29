package api

import (
	"api-server/db"
	"api-server/models"
	"api-server/utils"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"net/http"
	"time"
)

// GithubResponse struct
type GithubResponse struct {
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatarURL"`
}

// Profiles struct
type Profiles struct {
	client       *mongo.Client
	collProfiles *mongo.Collection
	ctx          context.Context
	logger       *zap.SugaredLogger
}

// NewProfiles function
func NewProfiles(ctx context.Context, logger *zap.SugaredLogger, client *mongo.Client) *Profiles {
	return &Profiles{
		client:       client,
		collProfiles: db.GetCollections(client).Profiles,
		ctx:          ctx,
		logger:       logger,
	}
}

// GetProfile function
func (handler *Profiles) GetProfile(c *gin.Context) {
	handler.logger.Info("REST - GET - GetProfile called")

	session := sessions.Default(c)
	profile, err := utils.GetProfileFromSession(&session)
	if err != nil {
		handler.logger.Error("REST - GET - GetProfile - Cannot get user profile")
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

// PostProfilesToken function
func (handler *Profiles) PostProfilesToken(c *gin.Context) {
	handler.logger.Info("REST - POST - PostProfilesToken called")

	// get profileID from path params
	profileID, errID := primitive.ObjectIDFromHex(c.Param("id"))
	if errID != nil {
		handler.logger.Error("REST - POST - PostProfilesToken - wrong format of the path param 'id'")
		c.JSON(http.StatusBadRequest, gin.H{"error": "wrong format of the path param 'id'"})
		return
	}

	// retrieve current profile object from database (using session profile as input)
	session := sessions.Default(c)
	profileSession, err := utils.GetProfileFromSession(&session)
	if err != nil {
		handler.logger.Error("REST - POST - PostProfilesToken - cannot find profile in session")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "cannot find profile in session"})
		return
	}

	// check if the profile you are trying to update (path param) is your profile (session profile)
	if profileSession.ID != profileID {
		handler.logger.Error("REST - POST - ProfilesToken - Current profileID is different than profileID in session")
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot re-generate APIToken for a different profile then yours"})
		return
	}

	apiToken := uuid.NewString()

	_, err = handler.collProfiles.UpdateOne(handler.ctx, bson.M{
		"_id": profileSession.ID,
	}, bson.M{
		"$set": bson.M{
			"apiToken":   apiToken,
			"modifiedAt": time.Now(),
		},
	})
	if err != nil {
		handler.logger.Error("REST - POST - PostProfilesToken - Cannot update profile with the new apiToken")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot update apiToken"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"apiToken": apiToken})
}
