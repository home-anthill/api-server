package handlers

import (
	"log"
	"net/http"
	"time"

	"air-conditioner/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/net/context"
)

type ACsHandler struct {
	collection *mongo.Collection
	ctx        context.Context
}

func NewACsHandler(ctx context.Context, collection *mongo.Collection) *ACsHandler {
	return &ACsHandler{
		collection: collection,
		ctx:        ctx,
	}
}

// swagger:operation GET /airconditioners airconditioners getACs
// Returns list of airconditioners
// ---
// produces:
// - application/json
// responses:
//     '200':
//         description: Successful operation
func (handler *ACsHandler) GetACsHandler(c *gin.Context) {
	log.Printf("Request to MongoDB")
	cur, err := handler.collection.Find(handler.ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cur.Close(handler.ctx)

	airconditioners := make([]models.AirConditioner, 0)
	for cur.Next(handler.ctx) {
		var ac models.AirConditioner
		cur.Decode(&ac)
		airconditioners = append(airconditioners, ac)
	}
	c.JSON(http.StatusOK, airconditioners)
}

// swagger:operation POST /airconditioners airconditioners postAC
// Create a new airconditioner
// ---
// produces:
// - application/json
// responses:
//     '200':
//         description: Successful operation
//     '400':
//         description: Invalid input
func (handler *ACsHandler) PostACHandler(c *gin.Context) {
	var ac models.AirConditioner
	if err := c.ShouldBindJSON(&ac); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ac.ID = primitive.NewObjectID()
	ac.CreatedAt = time.Now()
	ac.ModifiedAt = time.Now()

	// set default status values
	var status models.Status
	status.On = true
	status.Mode = 0
	status.TargetTemperature = 0
	status.Fan.Mode = 0
	status.Fan.Speed = 0

	ac.Status = status

	_, err := handler.collection.InsertOne(handler.ctx, ac)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while inserting a new ac"})
		return
	}
	c.JSON(http.StatusOK, ac)
}

// swagger:operation PUT /airconditioners/{id} airconditioners putAC
// Update an existing airconditioner
// ---
// parameters:
// - name: name
//   manufacturer: AC manufacturer
//   model: AC model
// produces:
// - application/json
// responses:
//     '200':
//         description: Successful operation
//     '400':
//         description: Invalid input
//     '404':
//         description: Invalid home ID
func (handler *ACsHandler) PutACHandler(c *gin.Context) {
	id := c.Param("id")
	var ac models.AirConditioner
	if err := c.ShouldBindJSON(&ac); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	objectId, _ := primitive.ObjectIDFromHex(id)
	_, err := handler.collection.UpdateOne(handler.ctx, bson.M{
		"_id": objectId,
	}, bson.M{
		"$set": bson.M{
			"name":         ac.Name,
			"manufacturer": ac.Manufacturer,
			"model":        ac.Model,
			"status":       ac.Status,
			"modifiedAt":   time.Now(),
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "AC has been updated"})
}

// swagger:operation DELETE /airconditioners/{id} airconditioners deleteAC
// Delete an existing airconditioner
// ---
// produces:
// - application/json
// responses:
//     '200':
//         description: Successful operation
//     '404':
//         description: Invalid home ID
func (handler *ACsHandler) DeleteACHandler(c *gin.Context) {
	id := c.Param("id")
	objectId, _ := primitive.ObjectIDFromHex(id)
	_, err := handler.collection.DeleteOne(handler.ctx, bson.M{
		"_id": objectId,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "AC has been deleted"})
}
