package handlers

import (
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	"time"

	"air-conditioner/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/net/context"
)

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
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	apiToken := uuid.NewString()

	_, err := handler.collection.UpdateOne(handler.ctx, bson.M{
		"_id": id,
	}, bson.M{
		"$set": bson.M{
			"apiToken": apiToken,
			"modifiedAt": time.Now(),
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"apiToken": apiToken})
}
