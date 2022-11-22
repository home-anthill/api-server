package models

import (
  "go.mongodb.org/mongo-driver/bson/primitive"
  "time"
)

// swagger:parameters acs newAC
type Status struct {
  On          bool `json:"on" bson:"on"`
  Mode        int  `json:"mode" bson:"mode"`
  Temperature int  `json:"temperature" bson:"temperature"`
  FanSpeed    int  `json:"fanSpeed" bson:"fanSpeed"`
  Swing       bool `json:"swing" bson:"swing"`
}

// swagger:parameters acs newAC
type AirConditioner struct {
  //swagger:ignore
  ID             primitive.ObjectID `json:"id" bson:"_id"`
  UUID           string             `json:"uuid" bson:"uuid"`
  Mac            string             `json:"mac" bson:"mac"`
  Name           string             `json:"name" bson:"name"`
  Manufacturer   string             `json:"manufacturer" bson:"manufacturer"`
  Model          string             `json:"model" bson:"model"`
  ProfileOwnerId string             `json:"profileOwnerId" bson:"profileOwnerId"`
  ApiToken       string             `json:"apiToken" bson:"apiToken"`
  Status         Status             `json:"status" bson:"status"`

  CreatedAt  time.Time `json:"createdAt" bson:"createdAt"`
  ModifiedAt time.Time `json:"modifiedAt" bson:"modifiedAt"`
}
