package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Device struct {
	//swagger:ignore
	ID           primitive.ObjectID `json:"id" bson:"_id"`
	Mac          string             `json:"mac" bson:"mac"`
	Name         string             `json:"name" bson:"name"`
	Manufacturer string             `json:"manufacturer" bson:"manufacturer"`
	Model        string             `json:"model" bson:"model"`
	CreatedAt    time.Time          `json:"createdAt" bson:"createdAt"`
	ModifiedAt   time.Time          `json:"modifiedAt" bson:"modifiedAt"`
}