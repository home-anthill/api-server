package api

import (
  "api-server/models"
  "api-server/utils"
  "github.com/gin-contrib/sessions"
  "github.com/gin-gonic/gin"
  "github.com/go-playground/validator/v10"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/bson/primitive"
  "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/mongo/options"
  "go.uber.org/zap"
  "golang.org/x/net/context"
  "net/http"
  "time"
)

type HomeNewReq struct {
	Name     string       `json:"name" validate:"required,min=1,max=50"`
	Location string       `json:"location" validate:"required,min=1,max=50"`
	Rooms    []RoomNewReq `json:"rooms" validate:"required,dive"`
}

type HomeUpdateReq struct {
	Name     string `json:"name" validate:"required,min=1,max=50"`
	Location string `json:"location" validate:"required,min=1,max=50"`
}

type RoomNewReq struct {
	Name  string `json:"name" validate:"required,min=1,max=50"`
	Floor int    `json:"floor" validate:"required,min=-50,max=300"`
}

type RoomUpdateReq struct {
	Name    string               `json:"name" validate:"required,min=1,max=50"`
	Floor   int                  `json:"floor" validate:"required,min=-50,max=300"`
	Devices []primitive.ObjectID `json:"devices" bson:"devices,omitempty"`
}

type Homes struct {
	collection         *mongo.Collection
	collectionProfiles *mongo.Collection
	ctx                context.Context
	logger             *zap.SugaredLogger
	validate           *validator.Validate
}

func NewHomes(ctx context.Context, logger *zap.SugaredLogger, collection *mongo.Collection, collectionProfiles *mongo.Collection, validate *validator.Validate) *Homes {
	return &Homes{
		collection:         collection,
		collectionProfiles: collectionProfiles,
		ctx:                ctx,
		logger:             logger,
		validate:           validate,
	}
}

// swagger:operation GET /homes homes getHomes
// Returns list of homes
// ---
// produces:
// - application/json
// responses:
//
//	'200':
//	    description: Successful operation
func (handler *Homes) GetHomes(c *gin.Context) {
	handler.logger.Info("REST - GET - GetHomes called")

	// retrieve current profile object from session
	session := sessions.Default(c)
	profileSession, err := utils.GetProfileFromSession(&session)
	if err != nil {
		handler.logger.Error("REST - GET - GetHomes - cannot find profile in session")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "cannot find profile in session"})
		return
	}

	// read profile from db. This is required to get fresh data from db, because data in session could be outdated
	var profile models.Profile
	err = handler.collectionProfiles.FindOne(handler.ctx, bson.M{
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
//
//	'200':
//	    description: Successful operation
//	'400':
//	    description: Invalid input
func (handler *Homes) PostHome(c *gin.Context) {
	handler.logger.Info("REST - POST - PostHome called")

	// retrieve current profile object from session
	session := sessions.Default(c)
	profileSession, err := utils.GetProfileFromSession(&session)
	if err != nil {
		handler.logger.Error("REST - POST - PostHome - cannot find profile in session")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "cannot find profile in session"})
		return
	}

	var newHome HomeNewReq
	if err = c.ShouldBindJSON(&newHome); err != nil {
		handler.logger.Error("REST - POST - PostHome - Cannot bind request body", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	err = handler.validate.Struct(newHome)
	if err != nil {
		handler.logger.Errorf("REST - POST - PostHome - request body is not valid, err %#v", err)
		var errFields = utils.GetErrorMessage(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body, these fields are not valid:" + errFields})
		return
	}

	newDate := time.Now()

	// create a Home document object
	var home models.Home
	home.ID = primitive.NewObjectID()
	home.Name = newHome.Name
	home.Location = newHome.Location
	home.CreatedAt = newDate
	home.ModifiedAt = newDate
	for i := 0; i < len(newHome.Rooms); i++ {
		var room models.Room
		room.ID = primitive.NewObjectID()
		room.Name = newHome.Rooms[i].Name
		room.Floor = newHome.Rooms[i].Floor
		room.CreatedAt = newDate
		room.ModifiedAt = newDate
		home.Rooms = append(home.Rooms, room)
	}

	_, err = handler.collection.InsertOne(handler.ctx, home)
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
//   - name: name
//     location: plain string
//
// produces:
// - application/json
// responses:
//
//	'200':
//	    description: Successful operation
//	'400':
//	    description: Invalid input
//	'404':
//	    description: Invalid home ID
func (handler *Homes) PutHome(c *gin.Context) {
	handler.logger.Info("REST - PUT - PutHome called")

	objectId, errId := primitive.ObjectIDFromHex(c.Param("id"))
	if errId != nil {
		handler.logger.Error("REST - PUT - PutHome - wrong format of the path param 'id'")
		c.JSON(http.StatusBadRequest, gin.H{"error": "wrong format of the path param 'id'"})
		return
	}

	var home HomeUpdateReq
	if err := c.ShouldBindJSON(&home); err != nil {
		handler.logger.Error("REST - PUT - PutHome - Cannot bind request body", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	err := handler.validate.Struct(home)
	if err != nil {
		handler.logger.Errorf("REST - PUT - PutHome - request body is not valid, err %#v", err)
		var errFields = utils.GetErrorMessage(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body, these fields are not valid:" + errFields})
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
//
//	'200':
//	    description: Successful operation
//	'404':
//	    description: Invalid home ID
func (handler *Homes) DeleteHome(c *gin.Context) {
	handler.logger.Info("REST - DELETE - DeleteHome called")

	objectId, errId := primitive.ObjectIDFromHex(c.Param("id"))
	if errId != nil {
		handler.logger.Error("REST - DELETE - DeleteHome - wrong format of the path param 'id'")
		c.JSON(http.StatusBadRequest, gin.H{"error": "wrong format of the path param 'id'"})
		return
	}

	// you can update a home only if you are the owner of that home
	session := sessions.Default(c)
	isOwned := handler.isHomeOwnedBy(session, objectId)

	if !isOwned {
		handler.logger.Error("REST - DELETE - DeleteHome - Cannot delete a home that is not in your profile")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete a home that is not in your profile"})
		return
	}

	// retrieve current profile object from session
	profile, err := utils.GetLoggedProfile(handler.ctx, &session, handler.collectionProfiles)
	if err != nil {
		handler.logger.Error("REST - DELETE - DeleteHome - cannot find profile in session")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "cannot find profile in session"})
		return
	}
	var newHomes []primitive.ObjectID
	for _, homeId := range profile.Homes {
		if homeId != objectId {
			newHomes = append(newHomes, homeId)
		}
	}

	_, errUpd := handler.collectionProfiles.UpdateOne(handler.ctx, bson.M{
		"_id": profile.ID,
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
//
//	'200':
//	    description: Successful operation
func (handler *Homes) GetRooms(c *gin.Context) {
	handler.logger.Info("REST - GET - GetRooms called")

	objectId, errId := primitive.ObjectIDFromHex(c.Param("id"))
	if errId != nil {
		handler.logger.Error("REST - GET - GetRooms - wrong format of the path param 'id'")
		c.JSON(http.StatusBadRequest, gin.H{"error": "wrong format of the path param 'id'"})
		return
	}

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
//
//	'200':
//	    description: Successful operation
//	'400':
//	    description: Invalid input
func (handler *Homes) PostRoom(c *gin.Context) {
	handler.logger.Info("REST - POST - PostRoom called")

	objectId, errId := primitive.ObjectIDFromHex(c.Param("id"))
	if errId != nil {
		handler.logger.Error("REST - POST - PostRoom - wrong format of the path param 'id'")
		c.JSON(http.StatusBadRequest, gin.H{"error": "wrong format of the path param 'id'"})
		return
	}

	var newRoom RoomNewReq
	if err := c.ShouldBindJSON(&newRoom); err != nil {
		handler.logger.Error("REST - POST - PostRoom - Cannot bind request body", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	err := handler.validate.Struct(newRoom)
	if err != nil {
		handler.logger.Errorf("REST - POST - PostRoom - request body is not valid, err %#v", err)
		var errFields = utils.GetErrorMessage(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body, these fields are not valid:" + errFields})
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
	err = handler.collection.FindOne(handler.ctx, bson.M{
		"_id": objectId,
	}).Decode(&home)
	if err != nil {
		handler.logger.Error("REST - POST - PostRoom - Cannot find rooms of the home with that id")
		c.JSON(http.StatusNotFound, gin.H{"error": "Cannot find rooms for that house"})
		return
	}

	newDate := time.Now()

	// create a Home document object
	var room models.Room
	room.ID = primitive.NewObjectID()
	room.Name = newRoom.Name
	room.Floor = newRoom.Floor
	room.CreatedAt = newDate
	room.ModifiedAt = newDate

	// add the new room to the home
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
//   - name: name
//     floor: number
//
// produces:
// - application/json
// responses:
//
//	'200':
//	    description: Successful operation
//	'400':
//	    description: Invalid input
//	'404':
//	    description: Invalid home ID
func (handler *Homes) PutRoom(c *gin.Context) {
	handler.logger.Info("REST - PUT - PutRoom called")

	objectId, errId := primitive.ObjectIDFromHex(c.Param("id"))
	objectRid, errRid := primitive.ObjectIDFromHex(c.Param("rid"))
	if errId != nil || errRid != nil {
		handler.logger.Error("REST - PUT - PutRoom - wrong format of one of the path params")
		c.JSON(http.StatusBadRequest, gin.H{"error": "wrong format of one of the path params"})
		return
	}

	var updateRoom RoomUpdateReq
	if err := c.ShouldBindJSON(&updateRoom); err != nil {
		handler.logger.Error("REST - PUT - PutRoom - Cannot bind request body", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	err := handler.validate.Struct(updateRoom)
	if err != nil {
		handler.logger.Errorf("REST - PUT - PutRoom - request body is not valid, err %#v", err)
		var errFields = utils.GetErrorMessage(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body, these fields are not valid:" + errFields})
		return
	}

	var home models.Home
	err = handler.collection.FindOne(handler.ctx, bson.M{
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
		handler.logger.Errorf("REST - PUT - PutRoom - Cannot find room with id: %v", objectRid)
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
			"rooms.$[x].name":       updateRoom.Name,
			"rooms.$[x].floor":      updateRoom.Floor,
			"rooms.$[x].devices":    updateRoom.Devices,
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
//
//	'200':
//	    description: Successful operation
//	'404':
//	    description: Invalid room ID
func (handler *Homes) DeleteRoom(c *gin.Context) {
	handler.logger.Info("REST - DELETE - DeleteRoom called")

	objectId, errId := primitive.ObjectIDFromHex(c.Param("id"))
	objectRid, errRid := primitive.ObjectIDFromHex(c.Param("rid"))
	if errId != nil || errRid != nil {
		handler.logger.Error("REST - PUT - PutRoom - wrong format of one of the path params")
		c.JSON(http.StatusBadRequest, gin.H{"error": "wrong format of one of the path params"})
		return
	}

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
		handler.logger.Errorf("REST - DELETE - DeleteRoom - Cannot find room with id: %v", objectRid)
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

func (handler *Homes) isHomeOwnedBy(session sessions.Session, objectId primitive.ObjectID) bool {
	// you can update a home only if you are the owner of that home
	// read profile from db. This is required to get fresh data from db, because data in session could be outdated

	profile, err := utils.GetLoggedProfile(handler.ctx, &session, handler.collectionProfiles)
	if err != nil {
		handler.logger.Error("isHomeOwnedBy - cannot find profile in session")
		return false
	}

	found := utils.Contains(profile.Homes, objectId)
	if !found {
		handler.logger.Error("isHomeOwnedBy - cannot update a home that is not in your profile")
		return false
	}
	return true
}
