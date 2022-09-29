package models

type Type string

const (
  Controller Type = "controller"
  Sensor     Type = "sensor"
)

type Feature struct {
  Type     Type   `json:"type"`
  Name     string `json:"name"`
  Enable   bool   `json:"enable"`
  Priority int    `json:"priority"`
}
