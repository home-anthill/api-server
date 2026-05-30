package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
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

// SpecFormat represents the format of spec
type SpecFormat string

// Supported Spec Formats.
const (
	Bool  SpecFormat = "bool"
	Int   SpecFormat = "int"
	Float SpecFormat = "float"
	List  SpecFormat = "list"
)

type SpecListItem struct {
	Value float32 `json:"value" bson:"value"`
	Text  string  `json:"text" bson:"text"`
}

type Spec struct {
	Format SpecFormat     `json:"format" bson:"format"`
	Min    *float64       `json:"min,omitempty" bson:"min,omitempty"`
	Max    *float64       `json:"max,omitempty" bson:"max,omitempty"`
	Step   *float64       `json:"step,omitempty" bson:"step,omitempty"`
	List   []SpecListItem `json:"list,omitempty" bson:"list,omitempty"`
}

// Feature struct
type Feature struct {
	UUID   string `json:"uuid" bson:"uuid"`
	Type   Type   `json:"type" bson:"type"`
	Name   string `json:"name" bson:"name"`
	Enable bool   `json:"enable" bson:"enable"`
	Order  int    `json:"order" bson:"order"`
	Unit   string `json:"unit" bson:"unit"`
	Spec   Spec   `json:"spec" bson:"spec"`
}

// Device struct
type Device struct {
	//swagger:ignore
	ID           bson.ObjectID `json:"id" bson:"_id"`
	UUID         string        `json:"uuid" bson:"uuid"`
	Mac          string        `json:"mac" bson:"mac"`
	Name         string        `json:"name" bson:"name"`
	Manufacturer string        `json:"manufacturer" bson:"manufacturer"`
	Model        string        `json:"model" bson:"model"`
	Features     []Feature     `json:"features" bson:"features"`
	CreatedAt    time.Time     `json:"createdAt" bson:"createdAt"`
	ModifiedAt   time.Time     `json:"modifiedAt" bson:"modifiedAt"`
}

// DeviceFeatureState struct
type DeviceFeatureState struct {
	FeatureUUID string  `json:"featureUuid" validate:"required"`
	Type        Type    `json:"type" validate:"required"` // feature type
	Name        string  `json:"name" validate:"required"` // feature name
	Value       float32 `json:"value" validate:"min=0"`   // feature value
	CreatedAt   int64   `json:"createdAt"`                // as unix epoch in milliseconds
	ModifiedAt  int64   `json:"modifiedAt"`               // as unix epoch in milliseconds
}
