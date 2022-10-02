package models

type Type string

const (
  Controller Type = "controller"
  Sensor     Type = "sensor"
)

type Feature struct {
  Type     Type   `json:"type" validate:"required,oneof='controller' 'sensor'"`
  Name     string `json:"name" validate:"required,min=2,max=20"`
  Enable   bool   `json:"enable" validate:"required,boolean"`
  Priority int    `json:"priority" validate:"required,gte=1"`
}
