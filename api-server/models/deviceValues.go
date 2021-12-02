package models

type StatusValue struct {
	On          bool `json:"on"`
	Temperature int  `json:"temperature"`
	Mode        int  `json:"mode"`
	FanMode     int  `json:"fanMode"`
	FanSpeed    int  `json:"fanSpeed"`
}

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
	FanMode int `json:"fanMode"`
}

type FanSpeedValue struct {
	FanSpeed int `json:"fanSpeed"`
}
