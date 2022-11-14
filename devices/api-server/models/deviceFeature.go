package models

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
