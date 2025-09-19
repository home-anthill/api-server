package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Type string
type Type string

// SupportedFeatureName string
type SupportedFeatureName string

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

// DeviceFeatureState struct
type DeviceFeatureState struct {
	FeatureUUID string  `json:"featureUuid" validate:"required"`
	Type        Type    `json:"type" validate:"required"`  // feature type
	Name        string  `json:"name" validate:"required"`  // feature name
	Value       float32 `json:"value" validate:"required"` // feature value
	CreatedAt   int64   `json:"createdAt"`                 // as unix epoch in milliseconds
	ModifiedAt  int64   `json:"modifiedAt"`                // as unix epoch in milliseconds
}
