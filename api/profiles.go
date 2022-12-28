package api

import (
	"api-server/utils"
	"github.com/gin-gonic/contrib/sessions"
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

type GithubResponse struct {
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatarURL"`
}

type Profiles struct {
	collection *mongo.Collection
	ctx        context.Context
	logger     *zap.SugaredLogger
}

func NewProfiles(ctx context.Context, logger *zap.SugaredLogger, collection *mongo.Collection) *Profiles {
	return &Profiles{
		collection: collection,
		ctx:        ctx,
		logger:     logger,
	}
}

func (handler *Profiles) GetProfile(c *gin.Context) {
	handler.logger.Info("REST - GET - GetProfile called")

	session := sessions.Default(c)
	profile, err := utils.GetProfileFromSession(&session)
	if err != nil {
		handler.logger.Error("REST - GET - GetProfile - Cannot get user profile")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Cannot get user profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id": profile.ID,
		"github": GithubResponse{
			Email:     profile.Github.Email,
			Login:     profile.Github.Login,
			Name:      profile.Github.Name,
			AvatarURL: profile.Github.AvatarURL,
		},
	})
}

// swagger:operation POST /profiles/:id/token
// Generate/update profile token to register new devices
// ---
// produces:
// - application/json
// responses:
//
//	'200':
//	    description: Successful operation
//	'400':
//	    description: Invalid input
func (handler *Profiles) PostProfilesToken(c *gin.Context) {
	handler.logger.Info("REST - POST - PostProfilesToken called")

	// get profileId from path params
	profileId, errId := primitive.ObjectIDFromHex(c.Param("id"))
	if errId != nil {
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
	if profileSession.ID != profileId {
		handler.logger.Error("REST - POST - ProfilesToken - Current profileId is different than profileId in session")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot re-generate ApiToken for a different profile then yours"})
		return
	}

	apiToken := uuid.NewString()

	_, err = handler.collection.UpdateOne(handler.ctx, bson.M{
		"_id": profileSession.ID,
	}, bson.M{
		"$set": bson.M{
			"apiToken":   apiToken,
			"modifiedAt": time.Now(),
		},
	})
	if err != nil {
		handler.logger.Error("REST - POST - PostProfilesToken - Cannot update profile with the new apiToken")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"apiToken": apiToken})
}
