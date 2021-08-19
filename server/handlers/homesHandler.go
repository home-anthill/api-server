package handlers

import (
	"net/http"
	"time"

	"air-conditioner/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/net/context"
)

type HomesHandler struct {
	collection *mongo.Collection
	ctx        context.Context
}

func NewHomesHandler(ctx context.Context, collection *mongo.Collection) *HomesHandler {
	return &HomesHandler{
		collection: collection,
		ctx:        ctx,
	}
}

// swagger:operation GET /homes homes getHomes
// Returns list of homes
// ---
// produces:
// - application/json
// responses:
//     '200':
//         description: Successful operation
func (handler *HomesHandler) GetHomesHandler(c *gin.Context) {
	cur, err := handler.collection.Find(handler.ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cur.Close(handler.ctx)

	homes := make([]models.Home, 0)
	for cur.Next(handler.ctx) {
		var home models.Home
		cur.Decode(&home)
		homes = append(homes, home)
	}
	c.JSON(http.StatusOK, homes)
}

// swagger:operation POST /homes homes postHome
// Create a new home
// ---
// produces:
// - application/json
// responses:
//     '200':
//         description: Successful operation
//     '400':
//         description: Invalid input
func (handler *HomesHandler) PostHomeHandler(c *gin.Context) {
	var home models.Home
	if err := c.ShouldBindJSON(&home); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	home.ID = primitive.NewObjectID()
	home.CreatedAt = time.Now()
	home.ModifiedAt = time.Now()
	_, err := handler.collection.InsertOne(handler.ctx, home)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while inserting a new home"})
		return
	}
	c.JSON(http.StatusOK, home)
}

// swagger:operation PUT /homes/{id} homes putHome
// Update an existing home
// ---
// parameters:
// - name: name
//   location: plain string
//   rooms: Room array
// produces:
// - application/json
// responses:
//     '200':
//         description: Successful operation
//     '400':
//         description: Invalid input
//     '404':
//         description: Invalid home ID
func (handler *HomesHandler) PutHomeHandler(c *gin.Context) {
	id := c.Param("id")
	var home models.Home
	if err := c.ShouldBindJSON(&home); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	objectId, _ := primitive.ObjectIDFromHex(id)
	_, err := handler.collection.UpdateOne(handler.ctx, bson.M{
		"_id": objectId,
	}, bson.D{{"$set", bson.D{
		{"name", home.Name},
		{"location", home.Location},
		{"rooms", home.Rooms},
		{"modifiedAt", time.Now()},
	}}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Home has been updated"})
}

// swagger:operation DELETE /homes/{id} homes deleteHome
// Delete an existing home
// ---
// produces:
// - application/json
// responses:
//     '200':
//         description: Successful operation
//     '404':
//         description: Invalid home ID
func (handler *HomesHandler) DeleteHomeHandler(c *gin.Context) {
	id := c.Param("id")
	objectId, _ := primitive.ObjectIDFromHex(id)
	_, err := handler.collection.DeleteOne(handler.ctx, bson.M{
		"_id": objectId,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Home has been deleted"})
}

// swagger:operation GET /homes/:id/rooms rooms getRooms
// Returns list of rooms of a home
// ---
// produces:
// - application/json
// responses:
//     '200':
//         description: Successful operation
func (handler *HomesHandler) GetRoomsHandler(c *gin.Context) {
	id := c.Param("id")
	objectId, _ := primitive.ObjectIDFromHex(id)

	var home models.Home
	err := handler.collection.FindOne(handler.ctx, bson.M{
		"_id": objectId,
	}).Decode(&home)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, home.Rooms)
}

// swagger:operation POST /rooms rooms postRoom
// Create a new room in a home
// ---
// produces:
// - application/json
// responses:
//     '200':
//         description: Successful operation
//     '400':
//         description: Invalid input
func (handler *HomesHandler) PostRoomHandler(c *gin.Context) {
	id := c.Param("id")
	objectId, _ := primitive.ObjectIDFromHex(id)

	var room models.Room
	if err := c.ShouldBindJSON(&room); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var home models.Home
	err := handler.collection.FindOne(handler.ctx, bson.M{
		"_id": objectId,
	}).Decode(&home)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	room.ID = primitive.NewObjectID()
	room.CreatedAt = time.Now()
	room.ModifiedAt = time.Now()
	home.Rooms = append(home.Rooms, room)

	_, err2 := handler.collection.UpdateOne(handler.ctx, bson.M{
		"_id": objectId,
	}, bson.D{{"$set", bson.D{
		{"rooms", home.Rooms},
		{"modifiedAt", time.Now()},
	}}})
	if err2 != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Room added to the home"})
}