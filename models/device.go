package models

import (
  "go.mongodb.org/mongo-driver/bson/primitive"
  "time"
)

type Type string

const (
  Controller Type = "controller"
  Sensor     Type = "sensor"
)

type Feature struct {
  UUID     string `json:"uuid" bson:"uuid"`
  Type     Type   `json:"type" bson:"type"`
  Name     string `json:"name" bson:"name"`
  Enable   bool   `json:"enable" bson:"enable"`
  Priority int    `json:"priority" bson:"priority"`
  Unit     string `json:"unit" bson:"unit"`
}

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

type DeviceState struct {
  // For 'On' you cannot use required for boolean, otherwise you cannot set as false
  On          bool `json:"on" validate:"boolean"`
  Temperature int  `json:"temperature" validate:"required,min=17,max=30"`
  Mode        int  `json:"mode" validate:"required,min=1,max=5"`
  FanSpeed    int  `json:"fanSpeed" validate:"required,min=1,max=5"`
}
