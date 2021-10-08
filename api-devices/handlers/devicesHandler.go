package handlers

import (
	"air-conditioner/models"
	mqttClient "air-conditioner/mqtt-client"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/net/context"
	"net/http"
	"time"
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

func (handler *DevicesHandler) PostRegisterDeviceHandler(c *gin.Context) {
	// receive a payload from devices with
	var registerBody models.Register
	if err := c.ShouldBindJSON(&registerBody); err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
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

	var newAc models.AirConditioner
	newAc.ID = primitive.NewObjectID()
	newAc.Mac = registerBody.Mac
	newAc.Name = registerBody.Name
	newAc.Manufacturer = registerBody.Manufacturer
	newAc.Model = registerBody.Model
	newAc.CreatedAt = time.Now()
	newAc.ModifiedAt = time.Now()

	// set default status values
	var status models.Status
	status.On = true
	status.Mode = 0
	status.TargetTemperature = 0
	status.Fan.Mode = 0
	status.Fan.Speed = 0

	newAc.Status = status

	_, err2 := handler.collection.InsertOne(handler.ctx, newAc)
	if err2 != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while inserting a new ac"})
		return
	}
	c.JSON(http.StatusOK, newAc)
}

func (handler *DevicesHandler) PostOnOffDeviceHandler(c *gin.Context) {
	var onoffValue models.OnOffValue
	if err := c.ShouldBindJSON(&onoffValue); err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}
	messageJSON, err := json.Marshal(onoffValue)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot process payload"})
		return
	}
	t := mqttClient.SendOnOff(onoffValue.UUID, messageJSON)
	select {
	case <- t.Done():
		if t.Error() != nil {
			fmt.Println(t.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot send message to device"})
		} else {
			fmt.Println("sending response")
			c.JSON(http.StatusOK, gin.H{"message": "onoff sent"})
		}
	}
}

func (handler *DevicesHandler) PostTemperatureDeviceHandler(c *gin.Context) {
	var temperatureValue models.TemperatureValue
	if err := c.ShouldBindJSON(&temperatureValue); err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}
	messageJSON, err := json.Marshal(temperatureValue)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot process payload"})
		return
	}
	t := mqttClient.SendTemperature(temperatureValue.UUID, messageJSON)
	select {
	case <- t.Done():
		if t.Error() != nil {
			fmt.Println(t.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot send message to device"})
		} else {
			fmt.Println("sending response")
			c.JSON(http.StatusOK, gin.H{"message": "temperature sent"})
		}
	}
}

func (handler *DevicesHandler) PostModeDeviceHandler(c *gin.Context) {
	var modeValue models.ModeValue
	if err := c.ShouldBindJSON(&modeValue); err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}
	messageJSON, err := json.Marshal(modeValue)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot process payload"})
		return
	}
	t := mqttClient.SendMode(modeValue.UUID, messageJSON)
	select {
	case <- t.Done():
		if t.Error() != nil {
			fmt.Println(t.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot send message to device"})
		} else {
			fmt.Println("sending response")
			c.JSON(http.StatusOK, gin.H{"message": "mode sent"})
		}
	}
}

func (handler *DevicesHandler) PostFanDeviceHandler(c *gin.Context) {
	var fanValue models.FanValue
	if err := c.ShouldBindJSON(&fanValue); err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}
	messageJSON, err := json.Marshal(fanValue)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot process payload"})
		return
	}
	t := mqttClient.SendFan(fanValue.UUID, messageJSON)
	select {
	case <- t.Done():
		if t.Error() != nil {
			fmt.Println(t.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot send message to device"})
		} else {
			fmt.Println("sending response")
			c.JSON(http.StatusOK, gin.H{"message": "fan sent"})
		}
	}
}

func (handler *DevicesHandler) PostSwingDeviceHandler(c *gin.Context) {
	var swingValue models.SwingValue
	if err := c.ShouldBindJSON(&swingValue); err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}
	messageJSON, err := json.Marshal(swingValue)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot process payload"})
		return
	}
	t := mqttClient.SendSwing(swingValue.UUID, messageJSON)
	select {
	case <- t.Done():
		if t.Error() != nil {
			fmt.Println(t.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot send message to device"})
		} else {
			fmt.Println("sending response")
			c.JSON(http.StatusOK, gin.H{"message": "swing sent"})
		}
	}
}