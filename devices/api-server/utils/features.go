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
