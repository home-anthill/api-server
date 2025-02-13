package api

import (
	"api-server/db"
	"api-server/models"
	"api-server/utils"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"net/http"
	"time"
)

// HomeNewReq struct
type HomeNewReq struct {
	Name     string       `json:"name" validate:"required,min=1,max=50"`
	Location string       `json:"location" validate:"required,min=1,max=50"`
	Rooms    []RoomNewReq `json:"rooms" validate:"required,dive"`
}

// HomeUpdateReq struct
type HomeUpdateReq struct {
	Name     string `json:"name" validate:"required,min=1,max=50"`
	Location string `json:"location" validate:"required,min=1,max=50"`
}

// RoomNewReq struct
type RoomNewReq struct {
	Name string `json:"name" validate:"required,min=1,max=50"`
	// cannot use 'required' because I should be able to set floor=0. It isn't a problem, because the default value is 0 :)
	Floor int `json:"floor" validate:"min=-50,max=300"`
}

// RoomUpdateReq struct
type RoomUpdateReq struct {
	Name string `json:"name" validate:"required,min=1,max=50"`
	// cannot use 'required' because I should be able to set floor=0. It isn't a problem, because the default value is 0 :)
	Floor int `json:"floor" validate:"min=-50,max=300"`
}

// Homes struct
type Homes struct {
	client       *mongo.Client
	collProfiles *mongo.Collection
	collHomes    *mongo.Collection
	ctx          context.Context
	logger       *zap.SugaredLogger
	validate     *validator.Validate
}

// NewHomes function
func NewHomes(ctx context.Context, logger *zap.SugaredLogger, client *mongo.Client, validate *validator.Validate) *Homes {
	return &Homes{
		client:       client,
		collProfiles: db.GetCollections(client).Profiles,
		collHomes:    db.GetCollections(client).Homes,
		ctx:          ctx,
		logger:       logger,
		validate:     validate,
	}
}

// GetHomes function
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
	err = handler.collProfiles.FindOne(handler.ctx, bson.M{
		"_id": profileSession.ID,
	}).Decode(&profile)
	if err != nil {
		handler.logger.Error("REST - GET - GetHomes - Cannot find profile in DB", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot find profile"})
		return
	}

	// extract Homes of that profile from db
	cur, err := handler.collHomes.Find(handler.ctx, bson.M{
		"_id": bson.M{"$in": profile.Homes},
	})
	//lint:ignore SA5001 no need to check this error on close
	defer cur.Close(handler.ctx)
	if err != nil {
		handler.logger.Error("REST - GET - GetHomes - Cannot get homes of profile in session", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot get your homes"})
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

// PostHome function
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
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
	home.Rooms = []models.Room{}
	for i := 0; i < len(newHome.Rooms); i++ {
		var room models.Room
		room.ID = primitive.NewObjectID()
		room.Name = newHome.Rooms[i].Name
		room.Floor = newHome.Rooms[i].Floor
		room.CreatedAt = newDate
		room.ModifiedAt = newDate
		home.Rooms = append(home.Rooms, room)
	}

	// start-session
	dbSession, err := handler.client.StartSession()
	if err != nil {
		handler.logger.Errorf("REST - POST - PostHome - cannot start a db session, err = %#v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unknown error while trying to add a new home"})
		return
	}
	// Defers ending the session after the transaction is committed or ended
	defer dbSession.EndSession(context.TODO())

	_, errTrans := dbSession.WithTransaction(context.TODO(), func(sessionCtx mongo.SessionContext) (interface{}, error) {
		// Official `mongo-driver` documentation state: "callback may be run
		// multiple times during WithTransaction due to retry attempts, so it must be idempotent."
		_, err1 := handler.collHomes.InsertOne(sessionCtx, home)
		if err1 != nil {
			handler.logger.Errorf("REST - POST - PostHome - Cannot insert new home in DB, err1 = %#v", err1)
			return nil, err1
		}
		// assign the new home to the user profile
		_, errUpd := handler.collProfiles.UpdateOne(
			sessionCtx,
			bson.M{"_id": profileSession.ID},
			bson.M{"$addToSet": bson.M{"homes": home.ID}},
		)
		if errUpd != nil {
			handler.logger.Errorf("REST - POST - PostHome - Cannot add new home to profile in DB, errUpd = %#v", errUpd)
		}
		return nil, errUpd
	}, options.Transaction().SetWriteConcern(writeconcern.Majority()))
	if errTrans != nil {
		handler.logger.Errorf("REST - POST - PostHome - Cannot add new home to profile in transaction, errTrans = %#v", errTrans)
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot add a new home to profile"})
		return
	}

	c.JSON(http.StatusOK, home)
}

// PutHome function
func (handler *Homes) PutHome(c *gin.Context) {
	handler.logger.Info("REST - PUT - PutHome called")

	objectID, errID := primitive.ObjectIDFromHex(c.Param("id"))
	if errID != nil {
		handler.logger.Error("REST - PUT - PutHome - wrong format of the path param 'id'")
		c.JSON(http.StatusBadRequest, gin.H{"error": "wrong format of the path param 'id'"})
		return
	}

	var home HomeUpdateReq
	if err := c.ShouldBindJSON(&home); err != nil {
		handler.logger.Error("REST - PUT - PutHome - Cannot bind request body", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
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
	isOwned := handler.isHomeOwnedBy(session, objectID)
	if !isOwned {
		handler.logger.Error("REST - PUT - PutHome - Request payload cannot contain Rooms. This API is made to change only the home object.")
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot update a home that is not in your profile"})
		return
	}

	_, errUpd := handler.collHomes.UpdateOne(handler.ctx, bson.M{
		"_id": objectID,
	}, bson.M{
		"$set": bson.M{
			"name":       home.Name,
			"location":   home.Location,
			"modifiedAt": time.Now(),
		},
	})
	if errUpd != nil {
		handler.logger.Error("REST - PUT - PutHome - Cannot update home in DB.")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot update home in Db"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "home has been updated"})
}

// DeleteHome function
func (handler *Homes) DeleteHome(c *gin.Context) {
	handler.logger.Info("REST - DELETE - DeleteHome called")

	objectID, errID := primitive.ObjectIDFromHex(c.Param("id"))
	if errID != nil {
		handler.logger.Error("REST - DELETE - DeleteHome - wrong format of the path param 'id'")
		c.JSON(http.StatusBadRequest, gin.H{"error": "wrong format of the path param 'id'"})
		return
	}

	// you can update a home only if you are the owner of that home
	session := sessions.Default(c)
	isOwned := handler.isHomeOwnedBy(session, objectID)

	if !isOwned {
		handler.logger.Error("REST - DELETE - DeleteHome - Cannot delete a home that is not in your profile")
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete a home that is not in your profile"})
		return
	}

	// retrieve current profile object from session
	profile, err := utils.GetLoggedProfile(handler.ctx, &session, handler.collProfiles)
	if err != nil {
		handler.logger.Error("REST - DELETE - DeleteHome - cannot find profile in session")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "cannot find profile in session"})
		return
	}
	var newHomes []primitive.ObjectID
	for _, homeID := range profile.Homes {
		if homeID != objectID {
			newHomes = append(newHomes, homeID)
		}
	}

	// start-session
	dbSession, err := handler.client.StartSession()
	if err != nil {
		handler.logger.Errorf("REST - DELETE - DeleteHome - cannot start a db session, err = %#v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unknown error while trying to remove an home"})
		return
	}
	// Defers ending the session after the transaction is committed or ended
	defer dbSession.EndSession(context.TODO())

	_, errTrans := dbSession.WithTransaction(context.TODO(), func(sessionCtx mongo.SessionContext) (interface{}, error) {
		// Official `mongo-driver` documentation state: "callback may be run
		// multiple times during WithTransaction due to retry attempts, so it must be idempotent."
		_, errUpd := handler.collProfiles.UpdateOne(sessionCtx, bson.M{
			"_id": profile.ID,
		}, bson.M{
			"$set": bson.M{
				"homes": newHomes,
			},
		})
		if errUpd != nil {
			handler.logger.Errorf("REST - DELETE - DeleteHome - Cannot remove home from profile in DB, errUpd = %#v", errUpd)
			return nil, errUpd
		}

		_, errDel := handler.collHomes.DeleteOne(sessionCtx, bson.M{
			"_id": objectID,
		})
		if errDel != nil {
			handler.logger.Errorf("REST - DELETE - DeleteHome - Cannot remove home from DB, errDel = %#v", errDel)
		}
		return nil, errDel
	}, options.Transaction().SetWriteConcern(writeconcern.Majority()))
	if errTrans != nil {
		handler.logger.Errorf("REST - DELETE - DeleteHome - Cannot delete home in transaction, errTrans = %#v", errTrans)
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete home from profile"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "home has been deleted"})
}

// GetRooms function
func (handler *Homes) GetRooms(c *gin.Context) {
	handler.logger.Info("REST - GET - GetRooms called")

	objectID, errID := primitive.ObjectIDFromHex(c.Param("id"))
	if errID != nil {
		handler.logger.Error("REST - GET - GetRooms - wrong format of the path param 'id'")
		c.JSON(http.StatusBadRequest, gin.H{"error": "wrong format of the path param 'id'"})
		return
	}

	// you can update a home only if you are the owner of that home
	session := sessions.Default(c)
	isOwned := handler.isHomeOwnedBy(session, objectID)

	if !isOwned {
		handler.logger.Error("REST - GET - GetRooms - Cannot get rooms, because you aren't the owner of that house")
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot get rooms of an home that is not in your profile"})
		return
	}

	var home models.Home
	err := handler.collHomes.FindOne(handler.ctx, bson.M{
		"_id": objectID,
	}).Decode(&home)
	if err != nil {
		handler.logger.Error("REST - GET - GetRooms - Cannot find rooms of the home with that id")
		c.JSON(http.StatusNotFound, gin.H{"error": "cannot find rooms for that home"})
		return
	}
	c.JSON(http.StatusOK, home.Rooms)
}

// PostRoom function
func (handler *Homes) PostRoom(c *gin.Context) {
	handler.logger.Info("REST - POST - PostRoom called")

	objectID, errID := primitive.ObjectIDFromHex(c.Param("id"))
	if errID != nil {
		handler.logger.Error("REST - POST - PostRoom - wrong format of the path param 'id'")
		c.JSON(http.StatusBadRequest, gin.H{"error": "wrong format of the path param 'id'"})
		return
	}

	var newRoom RoomNewReq
	if err := c.ShouldBindJSON(&newRoom); err != nil {
		handler.logger.Error("REST - POST - PostRoom - Cannot bind request body", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
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
	isOwned := handler.isHomeOwnedBy(session, objectID)

	if !isOwned {
		handler.logger.Error("REST - POST - PostRoom - Cannot create a room in an home that is not in session profile")
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot create a room in an home that is not in your profile"})
		return
	}

	var home models.Home
	err = handler.collHomes.FindOne(handler.ctx, bson.M{
		"_id": objectID,
	}).Decode(&home)
	if err != nil {
		handler.logger.Error("REST - POST - PostRoom - Cannot find rooms of the home with that id")
		c.JSON(http.StatusNotFound, gin.H{"error": "cannot find home"})
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

	_, errUpd := handler.collHomes.UpdateOne(handler.ctx, bson.M{
		"_id": objectID,
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

	c.JSON(http.StatusOK, gin.H{"message": "room added to the home"})
}

// PutRoom function
func (handler *Homes) PutRoom(c *gin.Context) {
	handler.logger.Info("REST - PUT - PutRoom called")

	homeID, errID := primitive.ObjectIDFromHex(c.Param("id"))
	roomID, errRid := primitive.ObjectIDFromHex(c.Param("rid"))
	if errID != nil || errRid != nil {
		handler.logger.Error("REST - PUT - PutRoom - wrong format of one of the path params")
		c.JSON(http.StatusBadRequest, gin.H{"error": "wrong format of one of the path params"})
		return
	}

	var updateRoom RoomUpdateReq
	if err := c.ShouldBindJSON(&updateRoom); err != nil {
		handler.logger.Error("REST - PUT - PutRoom - Cannot bind request body", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}
	if err := handler.validate.Struct(updateRoom); err != nil {
		handler.logger.Errorf("REST - PUT - PutRoom - request body is not valid, err %#v", err)
		var errFields = utils.GetErrorMessage(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body, these fields are not valid:" + errFields})
		return
	}

	// you can update a home only if you are the owner of that home
	session := sessions.Default(c)
	isOwned := handler.isHomeOwnedBy(session, homeID)

	if !isOwned {
		handler.logger.Error("REST - PUT - PutRoom - Cannot update a room in an home that is not in session profile")
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot update a room in an home that is not in your profile"})
		return
	}

	// get Home
	var home models.Home
	err := handler.collHomes.FindOne(handler.ctx, bson.M{
		"_id": homeID,
	}).Decode(&home)
	if err != nil {
		handler.logger.Error("REST - PUT - PutRoom - Cannot find rooms of the home with that id")
		c.JSON(http.StatusNotFound, gin.H{"error": "cannot find rooms for that home"})
		return
	}

	// `roomID` must be a room of `home`
	var roomFound bool
	for _, val := range home.Rooms {
		if val.ID == roomID {
			roomFound = true
		}
	}
	if !roomFound {
		handler.logger.Errorf("REST - PUT - PutRoom - Cannot find room with id: %v", roomID)
		c.JSON(http.StatusNotFound, gin.H{"error": "room not found"})
		return
	}

	// update room
	filter := bson.D{primitive.E{Key: "_id", Value: homeID}}
	arrayFilters := options.ArrayFilters{Filters: bson.A{bson.M{"x._id": roomID}}}
	upsert := true
	opts := options.UpdateOptions{
		ArrayFilters: &arrayFilters,
		Upsert:       &upsert,
	}
	update := bson.M{
		"$set": bson.M{
			"rooms.$[x].name":       updateRoom.Name,
			"rooms.$[x].floor":      updateRoom.Floor,
			"rooms.$[x].modifiedAt": time.Now(),
		},
	}
	_, errUpdate := handler.collHomes.UpdateOne(handler.ctx, filter, update, &opts)
	if errUpdate != nil {
		handler.logger.Errorf("REST - PUT - PutRoom - Cannot update a room in DB, errUpdate = %#v", errUpdate)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot update room"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "room has been updated"})
}

// DeleteRoom function
func (handler *Homes) DeleteRoom(c *gin.Context) {
	handler.logger.Info("REST - DELETE - DeleteRoom called")

	objectID, errID := primitive.ObjectIDFromHex(c.Param("id"))
	objectRid, errRid := primitive.ObjectIDFromHex(c.Param("rid"))
	if errID != nil || errRid != nil {
		handler.logger.Error("REST - PUT - PutRoom - wrong format of one of the path params")
		c.JSON(http.StatusBadRequest, gin.H{"error": "wrong format of one of the path params"})
		return
	}

	// you can update a home only if you are the owner of that home
	session := sessions.Default(c)
	isOwned := handler.isHomeOwnedBy(session, objectID)

	if !isOwned {
		handler.logger.Error("REST - DELETE - DeleteRoom - Cannot delete a room in an home that is not in session profile")
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete a room in an home that is not in your profile"})
		return
	}

	var home models.Home
	err := handler.collHomes.FindOne(handler.ctx, bson.M{
		"_id": objectID,
	}).Decode(&home)
	if err != nil {
		handler.logger.Error("REST - DELETE - DeleteRoom - Cannot find home")
		c.JSON(http.StatusNotFound, gin.H{"error": "home not found"})
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
		c.JSON(http.StatusNotFound, gin.H{"error": "room not found"})
		return
	}

	// delete room by id
	filter := bson.D{primitive.E{Key: "_id", Value: objectID}}
	update := bson.M{
		"$pull": bson.M{
			"rooms": bson.D{primitive.E{Key: "_id", Value: objectRid}},
		},
	}
	_, err2 := handler.collHomes.UpdateOne(handler.ctx, filter, update)
	if err2 != nil {
		handler.logger.Error("REST - PUT - PutRoom - Cannot delete room in DB")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot delete room"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "room has been deleted"})
}

func (handler *Homes) isHomeOwnedBy(session sessions.Session, objectID primitive.ObjectID) bool {
	// you can update a home only if you are the owner of that home
	// read profile from db. This is required to get fresh data from db, because data in session could be outdated

	profile, err := utils.GetLoggedProfile(handler.ctx, &session, handler.collProfiles)
	if err != nil {
		handler.logger.Error("isHomeOwnedBy - cannot find profile in session")
		return false
	}

	found := utils.Contains(profile.Homes, objectID)
	if !found {
		handler.logger.Error("isHomeOwnedBy - cannot update a home that is not in your profile")
		return false
	}
	return true
}
