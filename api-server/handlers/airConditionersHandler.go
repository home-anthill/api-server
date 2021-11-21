package handlers

import (
	"fmt"
	"github.com/gin-gonic/contrib/sessions"
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
	collection        *mongo.Collection
	collectionProfile *mongo.Collection
	ctx               context.Context
}

func NewACsHandler(ctx context.Context, collection *mongo.Collection, collectionProfile *mongo.Collection) *ACsHandler {
	return &ACsHandler{
		collection:        collection,
		collectionProfile: collectionProfile,
		ctx:               ctx,
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
	// retrieve current profile ID from session
	session := sessions.Default(c)
	profileSession := session.Get("profile").(models.Profile)
	fmt.Println("GetACsHandler with profileID = ", profileSession.ID)

	// search profile in DB
	// This is required to get fresh data from db, because data in session could be outdated
	var profile models.Profile
	err := handler.collectionProfile.FindOne(handler.ctx, bson.M{
		"_id": profileSession.ID,
	}).Decode(&profile)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot find profile"})
		return
	}
	// extract a list of ObjectIDs from profile.devices
	var objectIds []primitive.ObjectID
	for _, val := range profile.Devices {
		objectId, _ := primitive.ObjectIDFromHex(val)
		objectIds = append(objectIds, objectId)
	}
	// extract ACs from db
	cur, errAc := handler.collection.Find(handler.ctx, bson.M{
		"_id": bson.M{"$in": objectIds},
	})
	if errAc != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errAc.Error()})
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

	//// set default status values
	//var status models.Status
	//status.On = true
	//status.Mode = 0
	//status.TargetTemperature = 0
	//status.Fan.Mode = 0
	//status.Fan.Speed = 0
	//
	//ac.Status = status

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
			//"status":       ac.Status,
			"modifiedAt": time.Now(),
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
