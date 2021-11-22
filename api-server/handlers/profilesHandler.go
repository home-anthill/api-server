package handlers

import (
	"github.com/gin-gonic/contrib/sessions"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"time"

	"air-conditioner/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/net/context"
)

type ProfileResponse struct {
	ID     primitive.ObjectID `json:"id"`
	Github GithubResponse     `json:"github"`
}

type GithubResponse struct {
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatarURL"`
}

type ProfilesHandler struct {
	collection *mongo.Collection
	ctx        context.Context
}

func NewProfilesHandler(ctx context.Context, collection *mongo.Collection) *ProfilesHandler {
	return &ProfilesHandler{
		collection: collection,
		ctx:        ctx,
	}
}

func (handler *ProfilesHandler) GetProfileHandler(c *gin.Context) {
	var profile models.Profile
	var ok bool
	session := sessions.Default(c).Get("profile")
	if profile, ok = session.(models.Profile); ok {
		var profileToReturn ProfileResponse
		profileToReturn.ID = profile.ID
		profileToReturn.Github = GithubResponse{
			Email:     profile.Github.Email,
			Login:     profile.Github.Login,
			Name:      profile.Github.Name,
			AvatarURL: profile.Github.AvatarURL,
		}
		c.JSON(http.StatusOK, gin.H{"profile": profileToReturn})
		return
	}

	c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot get user profile"})
}

// swagger:operation POST /profiles/:id/token
// Generate/update profile token to register new devices
// ---
// produces:
// - application/json
// responses:
//     '200':
//         description: Successful operation
//     '400':
//         description: Invalid input
func (handler *ProfilesHandler) PostProfilesTokenHandler(c *gin.Context) {
	id := c.Param("id")
	var profile models.Profile
	if err := c.ShouldBindJSON(&profile); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	apiToken := uuid.NewString()

	_, err := handler.collection.UpdateOne(handler.ctx, bson.M{
		"_id": id,
	}, bson.M{
		"$set": bson.M{
			"apiToken":   apiToken,
			"modifiedAt": time.Now(),
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"apiToken": apiToken})
}
