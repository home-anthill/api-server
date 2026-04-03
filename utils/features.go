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

func HasOnlineFeature(features []models.Feature) bool {
	for _, feature := range features {
		if feature.Type == models.Sensor && feature.Name == "online" {
			return true
		}
	}
	return false
}

func GetOnlineFeature(features []models.Feature) *models.Feature {
	for i := range features {
		if features[i].Type == models.Sensor && features[i].Name == "online" {
			return &features[i]
		}
	}
	return nil
}
