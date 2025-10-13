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

	When("calling hasOnlineFeature", func() {
		It("should return a success if features contains a 'sensor' with name 'online' feature", func() {
			onlineFeature := models.Feature{
				UUID:   uuid.NewString(),
				Type:   "sensor",
				Name:   "online",
				Enable: true,
				Order:  1,
				Unit:   "-",
			}
			onlineSensor := []models.Feature{onlineFeature}
			hasOnlineFeature := HasOnlineFeature(onlineSensor)
			Expect(hasOnlineFeature).To(BeTrue())
		})
		It("should return a failure if features don't contain a 'sensor' with name 'online' feature", func() {
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
			hasOnlineFeature := HasOnlineFeature(featuresSensor)
			Expect(hasOnlineFeature).To(BeFalse())
		})
	})

	When("calling getOnlineFeature", func() {
		It("should return the feature object if features contains a 'sensor' with name 'online' feature", func() {
			onlineFeature := models.Feature{
				UUID:   uuid.NewString(),
				Type:   "sensor",
				Name:   "online",
				Enable: true,
				Order:  1,
				Unit:   "-",
			}
			sensorFeature1 := models.Feature{
				UUID:   uuid.NewString(),
				Type:   "sensor",
				Name:   "temperature",
				Enable: true,
				Order:  1,
				Unit:   "°C",
			}
			features := []models.Feature{sensorFeature1, onlineFeature}
			onlineFeatureFound := GetOnlineFeature(features)
			Expect(onlineFeatureFound.UUID).To(Equal(onlineFeature.UUID))
			Expect(onlineFeatureFound.Type).To(Equal(onlineFeature.Type))
			Expect(onlineFeatureFound.Name).To(Equal(onlineFeature.Name))
			Expect(onlineFeatureFound.Enable).To(Equal(onlineFeature.Enable))
			Expect(onlineFeatureFound.Order).To(Equal(onlineFeature.Order))
			Expect(onlineFeatureFound.Unit).To(Equal(onlineFeature.Unit))
		})
		It("should return nil if features don't contain a 'sensor' with name 'online' feature", func() {
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
			features := []models.Feature{sensorFeature1, sensorFeature2}
			onlineFeatureFound := GetOnlineFeature(features)
			Expect(onlineFeatureFound).To(BeNil())
		})
	})
})
