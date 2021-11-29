package models

type OnOffValue struct {
  On bool `json:"on"`
}

type TemperatureValue struct {
  Temperature int `json:"temperature"`
}

type ModeValue struct {
  Mode int `json:"mode"`
}

type FanModeValue struct {
  Fan int `json:"fan"`
}

type FanSwingValue struct {
  Swing bool `json:"swing"`
}
