package handlers

import (
	"fmt"
	"github.com/gin-gonic/contrib/sessions"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	collection         *mongo.Collection
	collectionProfiles *mongo.Collection
	ctx                context.Context
}

func NewHomesHandler(ctx context.Context, collection *mongo.Collection, collectionProfiles *mongo.Collection) *HomesHandler {
	return &HomesHandler{
		collection:         collection,
		collectionProfiles: collectionProfiles,
		ctx:                ctx,
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
	session := sessions.Default(c)
	profileSession := session.Get("profile").(models.Profile)
	fmt.Println("GetHomesHandler with profileID = ", profileSession.ID)

	// read profile from db. This is required to get fresh data from db, because data in session could be outdated
	var profile models.Profile
	err := handler.collectionProfiles.FindOne(handler.ctx, bson.M{
		"_id": profileSession.ID,
	}).Decode(&profile)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot find profile"})
		return
	}

	// extract Homes of that profile from db
	cur, err := handler.collection.Find(handler.ctx, bson.M{
		"_id": bson.M{"$in": profile.Homes},
	})
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
	session := sessions.Default(c)
	profileSession := session.Get("profile").(models.Profile)
	fmt.Println("PostHomeHandler with profileID = ", profileSession.ID)

	var home models.Home
	if err := c.ShouldBindJSON(&home); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	home.ID = primitive.NewObjectID()
	home.CreatedAt = time.Now()
	home.ModifiedAt = time.Now()
	for i := 0; i < len(home.Rooms); i++ {
		home.Rooms[i].ID = primitive.NewObjectID()
		home.Rooms[i].CreatedAt = time.Now()
		home.Rooms[i].ModifiedAt = time.Now()
	}

	_, err := handler.collection.InsertOne(handler.ctx, home)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while inserting a new home"})
		return
	}

	// assign the new home to the user profile
	_, errUpd := handler.collectionProfiles.UpdateOne(
		handler.ctx,
		bson.M{"_id": profileSession.ID},
		bson.M{"$push": bson.M{"homes": home.ID}},
	)
	if errUpd != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot push new home into profile"})
		return
	}

	c.JSON(http.StatusOK, home)
}

// swagger:operation PUT /homes/{id} homes putHome
// Update an existing home. You cannot pass rooms.
// ---
// parameters:
// - name: name
//   location: plain string
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
	if home.Rooms != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot pass rooms. This API is made to change only the home object"})
		return
	}

	objectId, _ := primitive.ObjectIDFromHex(id)

	// you can update a home only if you are the owner of that home
	session := sessions.Default(c)
	isOwned := handler.isHomeOwnedBy(session, objectId)
	if !isOwned {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot update a home that is not in your profile"})
		return
	}

	_, errUpd := handler.collection.UpdateOne(handler.ctx, bson.M{
		"_id": objectId,
	}, bson.M{
		"$set": bson.M{
			"name":       home.Name,
			"location":   home.Location,
			"modifiedAt": time.Now(),
		},
	})
	if errUpd != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errUpd.Error()})
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

	// you can update a home only if you are the owner of that home
	session := sessions.Default(c)
	isOwned := handler.isHomeOwnedBy(session, objectId)

	if !isOwned {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete a home that is not in your profile"})
		return
	}

	// remove home from profile.homes
	profileSession := session.Get("profile").(models.Profile)
	fmt.Println("DeleteHomeHandler with profileID = ", profileSession.ID)

	// read profile from db. This is required to get fresh data from db, because data in session could be outdated
	var profile models.Profile
	err := handler.collectionProfiles.FindOne(handler.ctx, bson.M{
		"_id": profileSession.ID,
	}).Decode(&profile)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot find profile"})
		return
	}
	fmt.Print("BEFORE profile.Homes ", profile.Homes)
	var newHomes []primitive.ObjectID
	for _, homeId := range profile.Homes {
		if homeId != objectId {
			newHomes = append(newHomes, homeId)
		}
	}
	fmt.Print("AFTER profile.Homes - newHomes ", newHomes)

	_, errUpd := handler.collectionProfiles.UpdateOne(handler.ctx, bson.M{
		"_id": profileSession.ID,
	}, bson.M{
		"$set": bson.M{
			"homes": newHomes,
		},
	})
	if errUpd != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot remove home from profile"})
		return
	}

	_, errDel := handler.collection.DeleteOne(handler.ctx, bson.M{
		"_id": objectId,
	})
	if errDel != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errDel.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Home has been deleted"})
}

// swagger:operation GET /homes/{id}/rooms rooms getRooms
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

	// you can update a home only if you are the owner of that home
	session := sessions.Default(c)
	isOwned := handler.isHomeOwnedBy(session, objectId)

	if !isOwned {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot get rooms of an home that is not in your profile"})
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
	c.JSON(http.StatusOK, home.Rooms)
}

// swagger:operation POST /homes/{id}/rooms rooms postRoom
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

	// you can update a home only if you are the owner of that home
	session := sessions.Default(c)
	isOwned := handler.isHomeOwnedBy(session, objectId)

	if !isOwned {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot create a room into an home that is not in your profile"})
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
	}, bson.M{
		"$set": bson.M{
			"rooms":      home.Rooms,
			"modifiedAt": time.Now(),
		},
	})
	if err2 != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Room added to the home"})
}

// swagger:operation PUT /homes/{id}/rooms/{rid} rooms putRoom
// Update an existing room of a home
// ---
// parameters:
// - name: name
//   floor: number
// produces:
// - application/json
// responses:
//     '200':
//         description: Successful operation
//     '400':
//         description: Invalid input
//     '404':
//         description: Invalid home ID
func (handler *HomesHandler) PutRoomHandler(c *gin.Context) {
	id := c.Param("id")
	objectId, _ := primitive.ObjectIDFromHex(id)

	rid := c.Param("rid")
	objectRid, _ := primitive.ObjectIDFromHex(rid)

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
		c.JSON(http.StatusNotFound, gin.H{"error": "Home not found"})
		return
	}

	// search if room is in rooms array
	var roomFound bool
	for _, val := range home.Rooms {
		if val.ID == objectRid {
			roomFound = true
		}
	}
	if !roomFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	// you can update a home only if you are the owner of that home
	session := sessions.Default(c)
	isOwned := handler.isHomeOwnedBy(session, objectId)

	if !isOwned {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot update a room of an home that is not in your profile"})
		return
	}

	// update room
	filter := bson.D{primitive.E{Key: "_id", Value: objectId}}
	arrayFilters := options.ArrayFilters{Filters: bson.A{bson.M{"x._id": objectRid}}}
	upsert := true
	opts := options.UpdateOptions{
		ArrayFilters: &arrayFilters,
		Upsert:       &upsert,
	}
	update := bson.M{
		"$set": bson.M{
			"rooms.$[x].name":            room.Name,
			"rooms.$[x].floor":           room.Floor,
			"rooms.$[x].airConditioners": room.AirConditioners,
			"rooms.$[x].modifiedAt":      time.Now(),
		},
	}
	_, err2 := handler.collection.UpdateOne(handler.ctx, filter, update, &opts)
	if err2 != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot update room"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Room has been updated"})
}

// swagger:operation DELETE /homes/{id}/rooms/{rid} rooms deleteRoom
// Delete an existing room for a home
// ---
// produces:
// - application/json
// responses:
//     '200':
//         description: Successful operation
//     '404':
//         description: Invalid room ID
func (handler *HomesHandler) DeleteRoomHandler(c *gin.Context) {
	id := c.Param("id")
	objectId, _ := primitive.ObjectIDFromHex(id)

	rid := c.Param("rid")
	objectRid, _ := primitive.ObjectIDFromHex(rid)

	var home models.Home
	err := handler.collection.FindOne(handler.ctx, bson.M{
		"_id": objectId,
	}).Decode(&home)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Home not found"})
		return
	}

	// you can update a home only if you are the owner of that home
	session := sessions.Default(c)
	isOwned := handler.isHomeOwnedBy(session, objectId)

	if !isOwned {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete a room of an home that is not in your profile"})
		return
	}

	// search if room is in rooms array
	var roomFound bool
	for _, val := range home.Rooms {
		if val.ID == objectRid {
			roomFound = true
		}
	}
	if !roomFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	// delete room by id
	filter := bson.D{primitive.E{Key: "_id", Value: objectId}}
	update := bson.M{
		"$pull": bson.M{
			"rooms": bson.D{primitive.E{Key: "_id", Value: objectRid}},
		},
	}
	_, err2 := handler.collection.UpdateOne(handler.ctx, filter, update)

	if err2 != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot update room"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Room has been deleted"})
}

func contains(s []primitive.ObjectID, objToFind primitive.ObjectID) bool {
	for _, v := range s {
		if v.Hex() == objToFind.Hex() {
			return true
		}
	}
	return false
}

func (handler *HomesHandler) isHomeOwnedBy(session sessions.Session, objectId primitive.ObjectID) bool {
	profileSessionId := session.Get("profile").(models.Profile).ID
	// you can update a home only if you are the owner of that home
	fmt.Println("isHomeOwnedBy profileSessionId = ", profileSessionId)
	// read profile from db. This is required to get fresh data from db, because data in session could be outdated
	var profile models.Profile
	err := handler.collectionProfiles.FindOne(handler.ctx, bson.M{
		"_id": profileSessionId,
	}).Decode(&profile)
	if err != nil {
		//c.JSON(http.StatusBadRequest, gin.H{"error": "cannot find profile"})
		fmt.Println("cannot find profile")
		return false
	}
	found := contains(profile.Homes, objectId)
	if !found {
		//c.JSON(http.StatusBadRequest, gin.H{"error": "cannot update a home that is not in your profile"})
		fmt.Println("cannot update a home that is not in your profile")
		return false
	}
	return true
}
