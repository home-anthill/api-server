package models

type OnOffValue struct {
  Uuid         string `json:"uuid"`
  ProfileToken string `json:"profileToken"`
  On           bool   `json:"on"`
}

type TemperatureValue struct {
  Uuid         string `json:"uuid"`
  ProfileToken string `json:"profileToken"`
  Temperature  int    `json:"temperature"`
}

type ModeValue struct {
  Uuid         string `json:"uuid"`
  ProfileToken string `json:"profileToken"`
  Mode         int    `json:"mode"`
}

type FanModeValue struct {
  Uuid         string `json:"uuid"`
  ProfileToken string `json:"profileToken"`
  FanMode      int    `json:"fanMode"`
}

type FanSpeedValue struct {
  Uuid         string `json:"uuid"`
  ProfileToken string `json:"profileToken"`
  FanSpeed     int    `json:"fanSpeed"`
}
