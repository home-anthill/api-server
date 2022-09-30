package models

import (
  "go.mongodb.org/mongo-driver/bson/primitive"
  "time"
)

// swagger:parameters rooms newRoom
type Room struct {
  //swagger:ignore
  ID         primitive.ObjectID   `json:"id" bson:"_id"`
  Name       string               `json:"name" bson:"name"`
  Floor      int                  `json:"floor" bson:"floor"`
  CreatedAt  time.Time            `json:"createdAt" bson:"createdAt"`
  ModifiedAt time.Time            `json:"modifiedAt" bson:"modifiedAt"`
  Devices    []primitive.ObjectID `json:"devices" bson:"devices,omitempty"`
}
