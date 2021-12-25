package api

import (
	"api-server/models"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"net/http"
	"time"
)

type Homes struct {
	collection         *mongo.Collection
	collectionProfiles *mongo.Collection
	ctx                context.Context
	logger             *zap.SugaredLogger
}

func NewHomes(ctx context.Context, logger *zap.SugaredLogger, collection *mongo.Collection, collectionProfiles *mongo.Collection) *Homes {
	return &Homes{
		collection:         collection,
		collectionProfiles: collectionProfiles,
		ctx:                ctx,
		logger:             logger,
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
func (handler *Homes) GetHomes(c *gin.Context) {
	handler.logger.Debug("REST - GET - GetHomes called")

	session := sessions.Default(c)
	profileSession := session.Get("profile").(models.Profile)

	// read profile from db. This is required to get fresh data from db, because data in session could be outdated
	var profile models.Profile
	err := handler.collectionProfiles.FindOne(handler.ctx, bson.M{
		"_id": profileSession.ID,
	}).Decode(&profile)
	if err != nil {
		handler.logger.Error("REST - GET - GetHomes - Cannot find profile in DB", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot find profile"})
		return
	}

	// extract Homes of that profile from db
	cur, err := handler.collection.Find(handler.ctx, bson.M{
		"_id": bson.M{"$in": profile.Homes},
	})
	defer cur.Close(handler.ctx)
	if err != nil {
		handler.logger.Error("REST - GET - GetHomes - Cannot get homes of profile in session", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot get your homes"})
		return
	}

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
func (handler *Homes) PostHome(c *gin.Context) {
	handler.logger.Debug("REST - POST - PostHome called")

	session := sessions.Default(c)
	profileSession := session.Get("profile").(models.Profile)

	var home models.Home
	if err := c.ShouldBindJSON(&home); err != nil {
		handler.logger.Error("REST - POST - PostHome - Cannot bind request body", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// update home object adding other fields before save it in DB
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
		handler.logger.Error("REST - POST - PostHome - Cannot insert new home in DB", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot add the new home"})
		return
	}

	// assign the new home to the user profile
	_, errUpd := handler.collectionProfiles.UpdateOne(
		handler.ctx,
		bson.M{"_id": profileSession.ID},
		bson.M{"$push": bson.M{"homes": home.ID}},
	)
	if errUpd != nil {
		handler.logger.Error("REST - POST - PostHome - Cannot add new home to profile in DB", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot add the new home to profile"})
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
func (handler *Homes) PutHome(c *gin.Context) {
	handler.logger.Debug("REST - PUT - PutHome called")

	id := c.Param("id")
	objectId, _ := primitive.ObjectIDFromHex(id)

	var home models.Home
	if err := c.ShouldBindJSON(&home); err != nil {
		handler.logger.Error("REST - PUT - PutHome - Cannot bind request body", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}
	if home.Rooms != nil {
		handler.logger.Error("REST - PUT - PutHome - Request payload cannot contain Rooms. This API is made to change only the home object.")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Request payload cannot contain Rooms. This API is made to change only the home object"})
		return
	}

	// you can update a home only if you are the owner of that home
	session := sessions.Default(c)
	isOwned := handler.isHomeOwnedBy(session, objectId)
	if !isOwned {
		handler.logger.Error("REST - PUT - PutHome - Request payload cannot contain Rooms. This API is made to change only the home object.")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot update a home that is not in your profile"})
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
		handler.logger.Error("REST - PUT - PutHome - Cannot update home in DB.")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot update home in Db"})
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
func (handler *Homes) DeleteHome(c *gin.Context) {
	handler.logger.Debug("REST - DELETE - DeleteHome called")

	id := c.Param("id")
	objectId, _ := primitive.ObjectIDFromHex(id)

	// you can update a home only if you are the owner of that home
	session := sessions.Default(c)
	isOwned := handler.isHomeOwnedBy(session, objectId)

	if !isOwned {
		handler.logger.Error("REST - DELETE - DeleteHome - Cannot delete a home that is not in your profile")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete a home that is not in your profile"})
		return
	}

	profileSession := session.Get("profile").(models.Profile)

	// read profile from db. This is required to get fresh data from db, because data in session could be outdated
	var profile models.Profile
	err := handler.collectionProfiles.FindOne(handler.ctx, bson.M{
		"_id": profileSession.ID,
	}).Decode(&profile)
	if err != nil {
		handler.logger.Error("REST - DELETE - DeleteHome - Cannot find profile from DB")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot find profile"})
		return
	}
	var newHomes []primitive.ObjectID
	for _, homeId := range profile.Homes {
		if homeId != objectId {
			newHomes = append(newHomes, homeId)
		}
	}

	_, errUpd := handler.collectionProfiles.UpdateOne(handler.ctx, bson.M{
		"_id": profileSession.ID,
	}, bson.M{
		"$set": bson.M{
			"homes": newHomes,
		},
	})
	if errUpd != nil {
		handler.logger.Error("REST - DELETE - DeleteHome - Cannot remove home from profile in DB")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot remove home from profile"})
		return
	}

	_, errDel := handler.collection.DeleteOne(handler.ctx, bson.M{
		"_id": objectId,
	})
	if errDel != nil {
		handler.logger.Error("REST - DELETE - DeleteHome - Cannot remove home from DB")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot delete home"})
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
func (handler *Homes) GetRooms(c *gin.Context) {
	handler.logger.Debug("REST - GET - GetRooms called")

	id := c.Param("id")
	objectId, _ := primitive.ObjectIDFromHex(id)

	// you can update a home only if you are the owner of that home
	session := sessions.Default(c)
	isOwned := handler.isHomeOwnedBy(session, objectId)

	if !isOwned {
		handler.logger.Error("REST - GET - GetRooms - Cannot get rooms, because you aren't the owner of that house")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot get rooms of an home that is not in your profile"})
		return
	}

	var home models.Home
	err := handler.collection.FindOne(handler.ctx, bson.M{
		"_id": objectId,
	}).Decode(&home)
	if err != nil {
		handler.logger.Error("REST - GET - GetRooms - Cannot find rooms of the home with that id")
		c.JSON(http.StatusNotFound, gin.H{"error": "Cannot find rooms for that house"})
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
func (handler *Homes) PostRoom(c *gin.Context) {
	handler.logger.Debug("REST - POST - PostRoom called")

	id := c.Param("id")
	objectId, _ := primitive.ObjectIDFromHex(id)

	var room models.Room
	if err := c.ShouldBindJSON(&room); err != nil {
		handler.logger.Error("REST - POST - PostRoom - Cannot bind request body", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// you can update a home only if you are the owner of that home
	session := sessions.Default(c)
	isOwned := handler.isHomeOwnedBy(session, objectId)

	if !isOwned {
		handler.logger.Error("REST - POST - PostRoom - Cannot create a room in an home that is not in session profile")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot create a room in an home that is not in your profile"})
		return
	}

	var home models.Home
	err := handler.collection.FindOne(handler.ctx, bson.M{
		"_id": objectId,
	}).Decode(&home)
	if err != nil {
		handler.logger.Error("REST - POST - PostRoom - Cannot find rooms of the home with that id")
		c.JSON(http.StatusNotFound, gin.H{"error": "Cannot find rooms for that house"})
		return
	}

	room.ID = primitive.NewObjectID()
	room.CreatedAt = time.Now()
	room.ModifiedAt = time.Now()
	home.Rooms = append(home.Rooms, room)

	_, errUpd := handler.collection.UpdateOne(handler.ctx, bson.M{
		"_id": objectId,
	}, bson.M{
		"$set": bson.M{
			"rooms":      home.Rooms,
			"modifiedAt": time.Now(),
		},
	})
	if errUpd != nil {
		handler.logger.Error("REST - POST - PostRoom - Cannot update home with the new rooms")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot update home with the new rooms"})
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
func (handler *Homes) PutRoom(c *gin.Context) {
	handler.logger.Debug("REST - PUT - PutRoom called")

	id := c.Param("id")
	objectId, _ := primitive.ObjectIDFromHex(id)

	rid := c.Param("rid")
	objectRid, _ := primitive.ObjectIDFromHex(rid)

	var room models.Room
	if err := c.ShouldBindJSON(&room); err != nil {
		handler.logger.Error("REST - PUT - PutRoom - Cannot bind request body", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	var home models.Home
	err := handler.collection.FindOne(handler.ctx, bson.M{
		"_id": objectId,
	}).Decode(&home)
	if err != nil {
		handler.logger.Error("REST - PUT - PutRoom - Cannot find rooms of the home with that id")
		c.JSON(http.StatusNotFound, gin.H{"error": "Cannot find rooms for that house"})
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
		handler.logger.Error("REST - PUT - PutRoom - Cannot find room with id: " + rid)
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	// you can update a home only if you are the owner of that home
	session := sessions.Default(c)
	isOwned := handler.isHomeOwnedBy(session, objectId)

	if !isOwned {
		handler.logger.Error("REST - PUT - PutRoom - Cannot update a room in an home that is not in session profile")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot update a room in an home that is not in your profile"})
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
			"rooms.$[x].name":       room.Name,
			"rooms.$[x].floor":      room.Floor,
			"rooms.$[x].devices":    room.Devices,
			"rooms.$[x].modifiedAt": time.Now(),
		},
	}
	_, err2 := handler.collection.UpdateOne(handler.ctx, filter, update, &opts)
	if err2 != nil {
		handler.logger.Error("REST - PUT - PutRoom - Cannot update a room in DB")
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
func (handler *Homes) DeleteRoom(c *gin.Context) {
	handler.logger.Debug("REST - DELETE - DeleteRoom called")

	id := c.Param("id")
	objectId, _ := primitive.ObjectIDFromHex(id)

	rid := c.Param("rid")
	objectRid, _ := primitive.ObjectIDFromHex(rid)

	var home models.Home
	err := handler.collection.FindOne(handler.ctx, bson.M{
		"_id": objectId,
	}).Decode(&home)
	if err != nil {
		handler.logger.Error("REST - DELETE - DeleteRoom - Cannot find home")
		c.JSON(http.StatusNotFound, gin.H{"error": "Home not found"})
		return
	}

	// you can update a home only if you are the owner of that home
	session := sessions.Default(c)
	isOwned := handler.isHomeOwnedBy(session, objectId)

	if !isOwned {
		handler.logger.Error("REST - DELETE - DeleteRoom - Cannot delete a room in an home that is not in session profile")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete a room in an home that is not in your profile"})
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
		handler.logger.Error("REST - DELETE - DeleteRoom - Cannot find room with id: " + rid)
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
		handler.logger.Error("REST - PUT - PutRoom - Cannot delete room in DB")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot delete room"})
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

func (handler *Homes) isHomeOwnedBy(session sessions.Session, objectId primitive.ObjectID) bool {
	profileSessionId := session.Get("profile").(models.Profile).ID
	// you can update a home only if you are the owner of that home
	// read profile from db. This is required to get fresh data from db, because data in session could be outdated
	var profile models.Profile
	err := handler.collectionProfiles.FindOne(handler.ctx, bson.M{
		"_id": profileSessionId,
	}).Decode(&profile)
	if err != nil {
		handler.logger.Error("Cannot find profile")
		return false
	}
	found := contains(profile.Homes, objectId)
	if !found {
		handler.logger.Error("Cannot update a home that is not in your profile")
		return false
	}
	return true
}
