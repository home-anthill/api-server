package models

type Register struct {
	//swagger:ignore
	Mac          string `json:"mac"`
	Name         string `json:"name"`
	Manufacturer string `json:"manufacturer"`
	Model        string `json:"model"`
}
