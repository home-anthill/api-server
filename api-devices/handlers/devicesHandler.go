package handlers

import (
	"air-conditioner/models"
	mqttClient "air-conditioner/mqtt-client"
	"encoding/json"
	"fmt"
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