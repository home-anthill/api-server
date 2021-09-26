package models

type OnOffValue struct {
	UUID         string `json:"uuid"`
	ProfileToken string `json:"profileToken"`
	On           bool   `json:"on"`
}

type TemperatureValue struct {
	UUID         string `json:"uuid"`
	ProfileToken string `json:"profileToken"`
	Temperature  int    `json:"temperature"`
}

type ModeValue struct {
	UUID         string `json:"uuid"`
	ProfileToken string `json:"profileToken"`
	Mode         int    `json:"temperature"`
}

type FanValue struct {
	UUID         string `json:"uuid"`
	ProfileToken string `json:"profileToken"`
	Fan          int    `json:"fan"`
}

type SwingValue struct {
	UUID         string `json:"uuid"`
	ProfileToken string `json:"profileToken"`
	Swing        bool   `json:"swing"`
}
