package models

type Values struct {
  Uuid        string `json:"uuid"`
  ApiToken    string `json:"apiToken"`
  On          bool   `json:"on"`
  Temperature int    `json:"temperature"`
  Mode        int    `json:"mode"`
  FanSpeed    int    `json:"fanSpeed"`
}
