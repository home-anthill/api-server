package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Type string
type Type string

// Controller and Sensor types
const (
	Controller Type = "controller"
	Sensor     Type = "sensor"
)

// Feature struct
type Feature struct {
	UUID   string `json:"uuid" bson:"uuid"`
	Type   Type   `json:"type" bson:"type"`
	Name   string `json:"name" bson:"name"`
	Enable bool   `json:"enable" bson:"enable"`
	Order  int    `json:"order" bson:"order"`
	Unit   string `json:"unit" bson:"unit"`
}

// Device struct
type Device struct {
	//swagger:ignore
	ID           primitive.ObjectID `json:"id" bson:"_id"`
	UUID         string             `json:"uuid" bson:"uuid"`
	Mac          string             `json:"mac" bson:"mac"`
	Manufacturer string             `json:"manufacturer" bson:"manufacturer"`
	Model        string             `json:"model" bson:"model"`
	Features     []Feature          `json:"features" bson:"features"`
	CreatedAt    time.Time          `json:"createdAt" bson:"createdAt"`
	ModifiedAt   time.Time          `json:"modifiedAt" bson:"modifiedAt"`
}

// DeviceState struct
type DeviceState struct {
	// For 'On' you cannot use required for boolean, otherwise you cannot set as false
	On          bool  `json:"on" validate:"boolean"`
	Temperature int   `json:"temperature" validate:"required,min=17,max=30"`
	Mode        int   `json:"mode" validate:"required,min=1,max=5"`
	FanSpeed    int   `json:"fanSpeed" validate:"required,min=1,max=5"`
	CreatedAt   int64 `json:"createdAt"`  // as unix epoch in milliseconds
	ModifiedAt  int64 `json:"modifiedAt"` // as unix epoch in milliseconds
}

// SensorValue struct
type SensorValue struct {
	UUID       string  `json:"uuid"` // feature uuid
	Value      float64 `json:"value"`
	CreatedAt  int64   `json:"createdAt"`  // as unix epoch in milliseconds
	ModifiedAt int64   `json:"modifiedAt"` // as unix epoch in milliseconds
}
