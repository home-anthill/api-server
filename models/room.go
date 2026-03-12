package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Room struct
type Room struct {
	ID         bson.ObjectID   `json:"id" bson:"_id"`
	Name       string          `json:"name" bson:"name"`
	Floor      int             `json:"floor" bson:"floor"`
	CreatedAt  time.Time       `json:"createdAt" bson:"createdAt"`
	ModifiedAt time.Time       `json:"modifiedAt" bson:"modifiedAt"`
	Devices    []bson.ObjectID `json:"devices" bson:"devices,omitempty"`
}
