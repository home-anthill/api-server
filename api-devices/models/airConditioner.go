package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

// swagger:parameters acs newAC
type Fan struct {
	Mode  int `json:"mode" bson:"mode"`
	Speed int `json:"speed" bson:"speed"`
}

// swagger:parameters acs newAC
type Status struct {
	On                bool    `json:"on" bson:"on"`
	Mode              int     `json:"mode" bson:"mode"`
	TargetTemperature float32 `json:"targetTemperature" bson:"targetTemperature"`
	Fan               Fan     `json:"fan" bson:"fan"`
}

// swagger:parameters acs newAC
type AirConditioner struct {
	//swagger:ignore
	ID           primitive.ObjectID `json:"id" bson:"_id"`
	Name         string             `json:"name" bson:"name"`
	Manufacturer string             `json:"manufacturer" bson:"manufacturer"`
	Model        string             `json:"model" bson:"model"`
	Status       Status             `json:"status" bson:"status"`

	CreatedAt  time.Time `json:"createdAt" bson:"createdAt"`
	ModifiedAt time.Time `json:"modifiedAt" bson:"modifiedAt"`
}
