package utils

import (
	"api-server/models"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("using feature utils", func() {
	When("calling hasControllerFeature", func() {
		It("should return a success if features contains a 'controller' feature", func() {
			controllerFeature := models.Feature{
				UUID:   uuid.NewString(),
				Type:   "controller",
				Name:   "ac-beko",
				Enable: true,
				Order:  1,
				Unit:   "-",
			}
			featuresController := []models.Feature{controllerFeature}
			isController := HasControllerFeature(featuresController)
			Expect(isController).To(BeTrue())
		})
		It("should return a failure if features don't contain a 'controller' feature", func() {
			sensorFeature1 := models.Feature{
				UUID:   uuid.NewString(),
				Type:   "sensor",
				Name:   "temperature",
				Enable: true,
				Order:  1,
				Unit:   "°C",
			}
			sensorFeature2 := models.Feature{
				UUID:   uuid.NewString(),
				Type:   "sensor",
				Name:   "humidity",
				Enable: true,
				Order:  1,
				Unit:   "%",
			}
			featuresSensor := []models.Feature{sensorFeature1, sensorFeature2}
			isController := HasControllerFeature(featuresSensor)
			Expect(isController).To(BeFalse())
		})
	})
})
