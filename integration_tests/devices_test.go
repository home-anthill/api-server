package integration_tests

import (
	"api-server/initialization"
	"api-server/models"
	"api-server/testuutils"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"net/http"
	"net/http/httptest"
	"os"
	"time"
)

var _ = Describe("Devices", func() {
	var ctx context.Context
	var logger *zap.SugaredLogger
	var router *gin.Engine
	var collProfiles *mongo.Collection
	var collHomes *mongo.Collection
	var collDevices *mongo.Collection

	var currDate = time.Now()
	var deviceController = models.Device{
		ID:           primitive.NewObjectID(),
		Mac:          "11:22:33:44:55:66",
		Manufacturer: "test",
		Model:        "test",
		UUID:         uuid.NewString(),
		Features: []models.Feature{{
			UUID:   uuid.NewString(),
			Type:   "controller",
			Name:   "ac-beko",
			Enable: true,
			Order:  1,
			Unit:   "-",
		}},
		CreatedAt:  currDate,
		ModifiedAt: currDate,
	}
	var deviceSensor = models.Device{
		ID:           primitive.NewObjectID(),
		Mac:          "AA:22:33:44:55:BB",
		Manufacturer: "test2",
		Model:        "test2",
		UUID:         uuid.NewString(),
		Features: []models.Feature{{
			UUID:   uuid.NewString(),
			Type:   "sensor",
			Name:   "temperature",
			Enable: true,
			Order:  1,
			Unit:   "°C",
		}, {
			UUID:   uuid.NewString(),
			Type:   "sensor",
			Name:   "light",
			Enable: true,
			Order:  1,
			Unit:   "lux",
		}},
		CreatedAt:  currDate,
		ModifiedAt: currDate,
	}
	var home = models.Home{
		ID:       primitive.NewObjectID(),
		Name:     "home1",
		Location: "location1",
		Rooms: []models.Room{{
			ID:         primitive.NewObjectID(),
			Name:       "room1",
			Floor:      1,
			CreatedAt:  currDate,
			ModifiedAt: currDate,
			Devices:    []primitive.ObjectID{},
		}},
		CreatedAt:  currDate,
		ModifiedAt: currDate,
	}

	BeforeEach(func() {
		logger, router, ctx, collProfiles, collHomes, collDevices = initialization.Start()
		defer logger.Sync()

		err := os.Setenv("SINGLE_USER_LOGIN_EMAIL", "test@test.com")
		Expect(err).ShouldNot(HaveOccurred())
	})

	AfterEach(func() {
		testuutils.DropAllCollections(ctx, collProfiles, collHomes, collDevices)
	})

	Context("calling devices api GET", func() {
		BeforeEach(func() {
			err := testuutils.InsertOne(ctx, collDevices, deviceController)
			Expect(err).ShouldNot(HaveOccurred())
			err = testuutils.InsertOne(ctx, collDevices, deviceSensor)
			Expect(err).ShouldNot(HaveOccurred())
		})

		When("profile owns a device", func() {
			It("should get a list of devices", func() {
				jwtToken, cookieSession := testuutils.GetJwt(router)
				profileRes := testuutils.GetLoggedProfile(router, jwtToken, cookieSession)

				err := testuutils.AssignDeviceToProfile(ctx, collProfiles, profileRes.ID, deviceController.ID)
				Expect(err).ShouldNot(HaveOccurred())
				err = testuutils.AssignDeviceToProfile(ctx, collProfiles, profileRes.ID, deviceSensor.ID)
				Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, "/api/devices", nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusOK))
				var devices []models.Device
				err = json.Unmarshal(recorder.Body.Bytes(), &devices)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(devices).To(HaveLen(2))
			})
		})
	})

	Context("calling devices api DELETE", func() {
		BeforeEach(func() {
			err := testuutils.InsertOne(ctx, collDevices, deviceController)
			Expect(err).ShouldNot(HaveOccurred())
			err = testuutils.InsertOne(ctx, collDevices, deviceSensor)
			Expect(err).ShouldNot(HaveOccurred())
			err = testuutils.InsertOne(ctx, collHomes, home)
			Expect(err).ShouldNot(HaveOccurred())
		})

		When("profile owns 2 devices", func() {
			It("should remove the first one successfully", func() {
				jwtToken, cookieSession := testuutils.GetJwt(router)
				profileRes := testuutils.GetLoggedProfile(router, jwtToken, cookieSession)

				err := testuutils.AssignDeviceToProfile(ctx, collProfiles, profileRes.ID, deviceController.ID)
				Expect(err).ShouldNot(HaveOccurred())
				err = testuutils.AssignDeviceToProfile(ctx, collProfiles, profileRes.ID, deviceSensor.ID)
				Expect(err).ShouldNot(HaveOccurred())
				err = testuutils.AssignHomeToProfile(ctx, collProfiles, profileRes.ID, home.ID)
				Expect(err).ShouldNot(HaveOccurred())

				err = testuutils.AssignDeviceToHomeAndRoom(ctx, collHomes, home.ID, home.Rooms[0].ID, deviceController.ID)
				Expect(err).ShouldNot(HaveOccurred())

				devices, err := testuutils.FindAll[models.Device](ctx, collDevices)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(devices).To(HaveLen(2))

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodDelete, "/api/devices/"+deviceController.ID.Hex(), nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusOK))
				Expect(recorder.Body.String()).To(Equal(`{"message":"device has been deleted"}`))

				devices, err = testuutils.FindAll[models.Device](ctx, collDevices)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(devices).To(HaveLen(1))
				Expect(devices[0].ID).To(Equal(deviceSensor.ID))
				Expect(devices[0].Model).To(Equal(deviceSensor.Model))
				Expect(devices[0].Manufacturer).To(Equal(deviceSensor.Manufacturer))
				Expect(devices[0].Mac).To(Equal(deviceSensor.Mac))
				Expect(devices[0].UUID).To(Equal(deviceSensor.UUID))
				Expect(devices[0].Features).To(Equal(deviceSensor.Features))
			})
		})

		When("you pass bad inputs", func() {
			It("should return an error, because of bad deviceId", func() {
				jwtToken, cookieSession := testuutils.GetJwt(router)
				badDeviceID := "bad_device_id"

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodDelete, "/api/devices/"+badDeviceID, nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
				Expect(recorder.Body.String()).To(Equal(`{"error":"wrong format of device id"}`))
			})

			It("should return an error, because device is not owned by profile", func() {
				jwtToken, cookieSession := testuutils.GetJwt(router)

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodDelete, "/api/devices/"+deviceController.ID.Hex(), nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
				Expect(recorder.Body.String()).To(Equal(`{"error":"cannot delete device, because it is not in your profile"}`))
			})
		})
	})
})
