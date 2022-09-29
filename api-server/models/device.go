package models

import (
  "go.mongodb.org/mongo-driver/bson/primitive"
  "time"
)

type Device struct {
  //swagger:ignore
  ID           primitive.ObjectID `json:"id" bson:"_id"`
  UUID         string             `json:"uuid" bson:"uuid"`
  Mac          string             `json:"mac" bson:"mac"`
  Manufacturer string             `json:"manufacturer" bson:"manufacturer"`
  Model        string             `json:"model" bson:"model"`
  Features     []Feature          `json:"feature" bson:"feature"`
  CreatedAt    time.Time          `json:"createdAt" bson:"createdAt"`
  ModifiedAt   time.Time          `json:"modifiedAt" bson:"modifiedAt"`
}
