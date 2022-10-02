package models

type OnOffValue struct {
  Uuid     string `json:"uuid"`
  ApiToken string `json:"apiToken"`
  On       bool   `json:"on"`
}

type TemperatureValue struct {
  Uuid        string `json:"uuid"`
  ApiToken    string `json:"apiToken"`
  Temperature int    `json:"temperature"`
}

type ModeValue struct {
  Uuid     string `json:"uuid"`
  ApiToken string `json:"apiToken"`
  Mode     int    `json:"mode"`
}

type FanModeValue struct {
  Uuid     string `json:"uuid"`
  ApiToken string `json:"apiToken"`
  FanMode  int    `json:"fanMode"`
}

type FanSpeedValue struct {
  Uuid     string `json:"uuid"`
  ApiToken string `json:"apiToken"`
  FanSpeed int    `json:"fanSpeed"`
}
