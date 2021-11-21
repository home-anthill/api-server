package handlers

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	"time"

	"air-conditioner/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/net/context"
)

type DeviceRequest struct {
	//swagger:ignore
	Mac          string `json:"mac"`
	Name         string `json:"name"`
	Manufacturer string `json:"manufacturer"`
	Model        string `json:"model"`
	Type         string `json:"type"`
	APIToken     string `json:"apiToken"`
}


type RegisterHandler struct {
	collection         *mongo.Collection
	collectionProfiles *mongo.Collection
	ctx                context.Context
}

func NewRegisterHandler(ctx context.Context, collection *mongo.Collection, collectionProfiles *mongo.Collection) *RegisterHandler {
	return &RegisterHandler{
		collection:         collection,
		collectionProfiles: collectionProfiles,
		ctx:                ctx,
	}
}

func (handler *RegisterHandler) PostRegisterHandler(c *gin.Context) {
	// receive a payload from devices with
	var registerBody DeviceRequest
	if err := c.ShouldBindJSON(&registerBody); err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// search if profile token exists and retrieve profile
	var profileFound models.Profile
	errProfile := handler.collectionProfiles.FindOne(handler.ctx, bson.M{
		"apiToken": registerBody.APIToken,
	}).Decode(&profileFound)
	if errProfile != nil {
		fmt.Println("Cannot find profile with that apiToken", errProfile)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot register, profile token missing or not valid"})
		return
	}

	// search and skip db add if already exists
	var ac models.AirConditioner
	err := handler.collection.FindOne(handler.ctx, bson.M{
		"mac": registerBody.Mac,
	}).Decode(&ac)
	if err == nil {
		// if err == nil => ac found in db (already exists)
		// skip register process returning "already registered"
		c.JSON(http.StatusOK, gin.H{"message": "already registered"})
		return
	}

	ac = models.AirConditioner{}
	ac.ID = primitive.NewObjectID()
	ac.Mac = registerBody.Mac
	ac.Name = registerBody.Name
	ac.Manufacturer = registerBody.Manufacturer
	ac.Model = registerBody.Model
	ac.CreatedAt = time.Now()
	ac.ModifiedAt = time.Now()

	_, err2 := handler.collection.InsertOne(handler.ctx, ac)
	if err2 != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while inserting a new ac"})
		return
	}

	// push AC.ID to profile.devices
	result, errUpd := handler.collectionProfiles.UpdateOne(
		handler.ctx,
		bson.M{"_id": profileFound.ID},
		bson.M{"$push": bson.M{"devices": ac.ID}},
	)

	fmt.Println("result: ", result)
	fmt.Println("errUpd: ", errUpd)

	// call devices-server to add AC to its DB

	c.JSON(http.StatusOK, ac)
}
