package utils

import "api-server/models"

func HasControllerFeature(features []models.Feature) bool {
	for _, feature := range features {
		if feature.Type == models.Controller {
			return true
		}
	}
	return false
}

func HasPowerOutageFeature(features []models.Feature) bool {
	for _, feature := range features {
		if feature.Type == models.Sensor && feature.Name == "poweroutage" {
			return true
		}
	}
	return false
}
