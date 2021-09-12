package handlers

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/net/context"
	"net/http"
)

type DevicesHandler struct {
	collection *mongo.Collection
	ctx        context.Context
}

func NewDevicesHandler(ctx context.Context, collection *mongo.Collection) *DevicesHandler {
	return &DevicesHandler{
		collection: collection,
		ctx:        ctx,
	}
}

// swagger:operation POST /devices devices authorize
// Authorize a device
// ---
// produces:
// - application/json
// responses:
//     '200':
//         description: Successful operation
//     '400':
//         description: Invalid input
func (handler *DevicesHandler) AuthorizeDeviceHandler(c *gin.Context) {
	// TODO TODO TODO
	//var home models.Home
	//if err := c.ShouldBindJSON(&home); err != nil {
	//	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	//	return
	//}
	//
	//home.ID = primitive.NewObjectID()
	//home.CreatedAt = time.Now()
	//home.ModifiedAt = time.Now()
	//_, err := handler.collection.InsertOne(handler.ctx, home)
	//if err != nil {
	//	c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while inserting a new home"})
	//	return
	//}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}