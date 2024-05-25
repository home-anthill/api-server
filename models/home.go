package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

// Home struct
type Home struct {
	ID         primitive.ObjectID `json:"id" bson:"_id"`
	Name       string             `json:"name" bson:"name"`
	Location   string             `json:"location" bson:"location"`
	Rooms      []Room             `json:"rooms" bson:"rooms"`
	CreatedAt  time.Time          `json:"createdAt" bson:"createdAt"`
	ModifiedAt time.Time          `json:"modifiedAt" bson:"modifiedAt"`
}
