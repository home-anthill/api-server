package models

type OnOffValue struct {
  On bool `json:"on" validate:"required,boolean"`
}

type TemperatureValue struct {
  Temperature int `json:"temperature" validate:"required,min=-50,max=50"`
}

type ModeValue struct {
  Mode int `json:"mode" validate:"required,min=0,max=4"`
}

type FanModeValue struct {
  FanMode int `json:"fanMode" validate:"required,min=0,max=3"`
}

type FanSpeedValue struct {
  FanSpeed int `json:"fanSpeed" validate:"required,min=0,max=3"`
}
