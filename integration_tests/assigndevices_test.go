package integration_tests

import (
	"api-server/api"
	"api-server/initialization"
	"api-server/models"
	"api-server/test_utils"
	"bytes"
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

var _ = Describe("AssignDevices", func() {
	var ctx context.Context
	var logger *zap.SugaredLogger
	var router *gin.Engine
	var collProfiles *mongo.Collection
	var collHomes *mongo.Collection
	var collDevices *mongo.Collection

	var currDate = time.Now()
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
		}, {
			ID:         primitive.NewObjectID(),
			Name:       "room2",
			Floor:      2,
			CreatedAt:  currDate,
			ModifiedAt: currDate,
			Devices:    []primitive.ObjectID{},
		}},
		CreatedAt:  currDate,
		ModifiedAt: currDate,
	}
	var home2 = models.Home{
		ID:       primitive.NewObjectID(),
		Name:     "home2",
		Location: "location2",
		Rooms: []models.Room{{
			ID:         primitive.NewObjectID(),
			Name:       "room1",
			Floor:      3,
			CreatedAt:  currDate,
			ModifiedAt: currDate,
			Devices:    []primitive.ObjectID{},
		}},
		CreatedAt:  currDate,
		ModifiedAt: currDate,
	}
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

	BeforeEach(func() {
		logger, router, ctx, collProfiles, collHomes, collDevices = initialization.Start()
		defer logger.Sync()

		err := os.Setenv("SINGLE_USER_LOGIN_EMAIL", "test@test.com")
		Expect(err).ShouldNot(HaveOccurred())
	})

	AfterEach(func() {
		test_utils.DropAllCollections(ctx, collProfiles, collHomes, collDevices)
	})

	Context("calling assignDevice api PUT", func() {
		BeforeEach(func() {
			err := test_utils.InsertOne(ctx, collHomes, home)
			Expect(err).ShouldNot(HaveOccurred())
			err = test_utils.InsertOne(ctx, collHomes, home2)
			Expect(err).ShouldNot(HaveOccurred())
			err = test_utils.InsertOne(ctx, collDevices, deviceController)
			Expect(err).ShouldNot(HaveOccurred())
			err = test_utils.InsertOne(ctx, collDevices, deviceSensor)
			Expect(err).ShouldNot(HaveOccurred())
		})

		When("device is still not assigned to any home+room", func() {
			It("should assign device to home+room successfully", func() {
				jwtToken, cookieSession := test_utils.GetJwt(router)
				profileRes := test_utils.GetLoggedProfile(router, jwtToken, cookieSession)

				err := test_utils.AssignDeviceToProfile(ctx, collProfiles, profileRes.ID, deviceController.ID)
				Expect(err).ShouldNot(HaveOccurred())
				err = test_utils.AssignHomeToProfile(ctx, collProfiles, profileRes.ID, home.ID)
				Expect(err).ShouldNot(HaveOccurred())

				assignDeviceReq := api.AssignDeviceReq{
					HomeId: home.ID.Hex(),
					RoomId: home.Rooms[0].ID.Hex(),
				}
				var assignDeviceBug bytes.Buffer
				err = json.NewEncoder(&assignDeviceBug).Encode(assignDeviceReq)
				Expect(err).ShouldNot(HaveOccurred())
				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodPut, "/api/devices/"+deviceController.ID.Hex(), &assignDeviceBug)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusOK))
				Expect(recorder.Body.String()).To(Equal(`{"message":"device has been assigned to room"}`))

				homeFromDb, err := test_utils.FindOneById[models.Home](ctx, collHomes, home.ID)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(homeFromDb.Rooms[0].Devices).To(HaveLen(1))
				Expect(homeFromDb.Rooms[0].Devices[0]).To(Equal(deviceController.ID))
			})

			It("should not assign device, because of bad deviceId", func() {
				jwtToken, cookieSession := test_utils.GetJwt(router)
				profileRes := test_utils.GetLoggedProfile(router, jwtToken, cookieSession)

				err := test_utils.AssignDeviceToProfile(ctx, collProfiles, profileRes.ID, deviceController.ID)
				Expect(err).ShouldNot(HaveOccurred())
				err = test_utils.AssignHomeToProfile(ctx, collProfiles, profileRes.ID, home.ID)
				Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodPut, "/api/devices/"+"bad_device_id", nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
				Expect(recorder.Body.String()).To(Equal(`{"error":"wrong format of device 'id' path param"}`))
			})

			It("should not assign device, because body format is not valid", func() {
				jwtToken, cookieSession := test_utils.GetJwt(router)
				profileRes := test_utils.GetLoggedProfile(router, jwtToken, cookieSession)

				err := test_utils.AssignDeviceToProfile(ctx, collProfiles, profileRes.ID, deviceController.ID)
				Expect(err).ShouldNot(HaveOccurred())
				err = test_utils.AssignHomeToProfile(ctx, collProfiles, profileRes.ID, home.ID)
				Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodPut, "/api/devices/"+deviceController.ID.Hex(), nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
				Expect(recorder.Body.String()).To(Equal(`{"error":"Invalid request payload"}`))
			})

			It("should not assign device, because of body validation errors", func() {
				jwtToken, cookieSession := test_utils.GetJwt(router)
				profileRes := test_utils.GetLoggedProfile(router, jwtToken, cookieSession)

				err := test_utils.AssignDeviceToProfile(ctx, collProfiles, profileRes.ID, deviceController.ID)
				Expect(err).ShouldNot(HaveOccurred())
				err = test_utils.AssignHomeToProfile(ctx, collProfiles, profileRes.ID, home.ID)
				Expect(err).ShouldNot(HaveOccurred())

				assignDeviceReq := api.AssignDeviceReq{
					HomeId: "",
					RoomId: "",
				}
				var assignDeviceBug bytes.Buffer
				err = json.NewEncoder(&assignDeviceBug).Encode(assignDeviceReq)
				Expect(err).ShouldNot(HaveOccurred())
				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodPut, "/api/devices/"+deviceController.ID.Hex(), &assignDeviceBug)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
				logger.Infof("recorder.Body.String() %s", recorder.Body.String())
				Expect(recorder.Body.String()).To(Equal(`{"error":"invalid request body, these fields are not valid: homeid roomid"}`))
			})

			It("should not assign device, because of wrong format of homeid and roomid", func() {
				jwtToken, cookieSession := test_utils.GetJwt(router)
				profileRes := test_utils.GetLoggedProfile(router, jwtToken, cookieSession)

				err := test_utils.AssignDeviceToProfile(ctx, collProfiles, profileRes.ID, deviceController.ID)
				Expect(err).ShouldNot(HaveOccurred())
				err = test_utils.AssignHomeToProfile(ctx, collProfiles, profileRes.ID, home.ID)
				Expect(err).ShouldNot(HaveOccurred())

				assignDeviceReq := api.AssignDeviceReq{
					HomeId: "bad_format_id",
					RoomId: "bad_format_id",
				}
				var assignDeviceBug bytes.Buffer
				err = json.NewEncoder(&assignDeviceBug).Encode(assignDeviceReq)
				Expect(err).ShouldNot(HaveOccurred())
				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodPut, "/api/devices/"+deviceController.ID.Hex(), &assignDeviceBug)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
				logger.Infof("recorder.Body.String() %s", recorder.Body.String())
				Expect(recorder.Body.String()).To(Equal(`{"error":"wrong format of one of the values in body"}`))
			})

			It("should not assign device, because profile don't own that device", func() {
				jwtToken, cookieSession := test_utils.GetJwt(router)
				profileRes := test_utils.GetLoggedProfile(router, jwtToken, cookieSession)

				err := test_utils.AssignHomeToProfile(ctx, collProfiles, profileRes.ID, home.ID)
				Expect(err).ShouldNot(HaveOccurred())

				assignDeviceReq := api.AssignDeviceReq{
					HomeId: home.ID.Hex(),
					RoomId: home.Rooms[0].ID.Hex(),
				}
				var assignDeviceBug bytes.Buffer
				err = json.NewEncoder(&assignDeviceBug).Encode(assignDeviceReq)
				Expect(err).ShouldNot(HaveOccurred())
				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodPut, "/api/devices/"+deviceController.ID.Hex(), &assignDeviceBug)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusUnauthorized))
				Expect(recorder.Body.String()).To(Equal(`{"error":"you are not the owner of this device id = ` + deviceController.ID.Hex() + `"}`))
			})

			It("should not assign device, because profile don't own home", func() {
				jwtToken, cookieSession := test_utils.GetJwt(router)
				profileRes := test_utils.GetLoggedProfile(router, jwtToken, cookieSession)

				err := test_utils.AssignDeviceToProfile(ctx, collProfiles, profileRes.ID, deviceController.ID)
				Expect(err).ShouldNot(HaveOccurred())

				assignDeviceReq := api.AssignDeviceReq{
					HomeId: home.ID.Hex(),
					RoomId: home.Rooms[0].ID.Hex(),
				}
				var assignDeviceBug bytes.Buffer
				err = json.NewEncoder(&assignDeviceBug).Encode(assignDeviceReq)
				Expect(err).ShouldNot(HaveOccurred())
				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodPut, "/api/devices/"+deviceController.ID.Hex(), &assignDeviceBug)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusUnauthorized))
				Expect(recorder.Body.String()).To(Equal(`{"error":"you are not the owner of home id = ` + home.ID.Hex() + `"}`))
			})

			It("should not assign device, because homeId is not a home in db", func() {
				jwtToken, cookieSession := test_utils.GetJwt(router)
				profileRes := test_utils.GetLoggedProfile(router, jwtToken, cookieSession)

				unknownHomeId, _ := primitive.ObjectIDFromHex("640499d9973e9ad5f58d3135")

				err := test_utils.AssignDeviceToProfile(ctx, collProfiles, profileRes.ID, deviceController.ID)
				Expect(err).ShouldNot(HaveOccurred())
				err = test_utils.AssignHomeToProfile(ctx, collProfiles, profileRes.ID, unknownHomeId)
				Expect(err).ShouldNot(HaveOccurred())

				assignDeviceReq := api.AssignDeviceReq{
					HomeId: unknownHomeId.Hex(), // unknown ObjectID
					RoomId: home.Rooms[0].ID.Hex(),
				}
				var assignDeviceBug bytes.Buffer
				err = json.NewEncoder(&assignDeviceBug).Encode(assignDeviceReq)
				Expect(err).ShouldNot(HaveOccurred())
				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodPut, "/api/devices/"+deviceController.ID.Hex(), &assignDeviceBug)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusNotFound))
				Expect(recorder.Body.String()).To(Equal(`{"error":"Cannot find home id = ` + unknownHomeId.Hex() + `"}`))
			})

			It("should not assign device, because roomId is not a room of home", func() {
				jwtToken, cookieSession := test_utils.GetJwt(router)
				profileRes := test_utils.GetLoggedProfile(router, jwtToken, cookieSession)

				unknownRoomId, _ := primitive.ObjectIDFromHex("640499d9973e9ad5f58d3135")

				err := test_utils.AssignDeviceToProfile(ctx, collProfiles, profileRes.ID, deviceController.ID)
				Expect(err).ShouldNot(HaveOccurred())
				err = test_utils.AssignHomeToProfile(ctx, collProfiles, profileRes.ID, home.ID)
				Expect(err).ShouldNot(HaveOccurred())

				assignDeviceReq := api.AssignDeviceReq{
					HomeId: home.ID.Hex(),
					RoomId: unknownRoomId.Hex(), // unknown ObjectID
				}
				var assignDeviceBug bytes.Buffer
				err = json.NewEncoder(&assignDeviceBug).Encode(assignDeviceReq)
				Expect(err).ShouldNot(HaveOccurred())
				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodPut, "/api/devices/"+deviceController.ID.Hex(), &assignDeviceBug)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusNotFound))
				Expect(recorder.Body.String()).To(Equal(`{"error":"Cannot find room id = ` + unknownRoomId.Hex() + `"}`))
			})
		})

		When("device is already assigned to another room of the same home", func() {
			It("should assign device to home+room and clean previous assignment successfully", func() {
				jwtToken, cookieSession := test_utils.GetJwt(router)
				profileRes := test_utils.GetLoggedProfile(router, jwtToken, cookieSession)

				err := test_utils.AssignDeviceToProfile(ctx, collProfiles, profileRes.ID, deviceController.ID)
				Expect(err).ShouldNot(HaveOccurred())
				err = test_utils.AssignDeviceToProfile(ctx, collProfiles, profileRes.ID, deviceSensor.ID)
				Expect(err).ShouldNot(HaveOccurred())
				err = test_utils.AssignHomeToProfile(ctx, collProfiles, profileRes.ID, home.ID)
				Expect(err).ShouldNot(HaveOccurred())

				// assign `deviceController` to `room2`
				err = test_utils.AssignDeviceToHomeAndRoom(ctx, collHomes, home.ID, home.Rooms[1].ID, deviceController.ID)
				Expect(err).ShouldNot(HaveOccurred())
				homeFromDb, err := test_utils.FindOneById[models.Home](ctx, collHomes, home.ID)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(homeFromDb.Rooms[1].Devices).To(HaveLen(1))
				Expect(homeFromDb.Rooms[1].Devices[0]).To(Equal(deviceController.ID))

				// assign `deviceController` to `room1`
				assignDeviceReq := api.AssignDeviceReq{
					HomeId: home.ID.Hex(),
					RoomId: home.Rooms[0].ID.Hex(),
				}
				var assignDeviceBug bytes.Buffer
				err = json.NewEncoder(&assignDeviceBug).Encode(assignDeviceReq)
				Expect(err).ShouldNot(HaveOccurred())
				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodPut, "/api/devices/"+deviceController.ID.Hex(), &assignDeviceBug)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusOK))
				Expect(recorder.Body.String()).To(Equal(`{"message":"device has been assigned to room"}`))

				// `deviceController` should be assigned only to `room1` and not also to `room2`
				homeFromDb, err = test_utils.FindOneById[models.Home](ctx, collHomes, home.ID)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(homeFromDb.Rooms[0].Devices).To(HaveLen(1))
				Expect(homeFromDb.Rooms[0].Devices[0]).To(Equal(deviceController.ID))
			})
		})

		When("device is already assigned to another home of the same profile", func() {
			It("should assign device to home+room and clean previous assignment successfully", func() {
				jwtToken, cookieSession := test_utils.GetJwt(router)
				profileRes := test_utils.GetLoggedProfile(router, jwtToken, cookieSession)

				err := test_utils.AssignDeviceToProfile(ctx, collProfiles, profileRes.ID, deviceController.ID)
				Expect(err).ShouldNot(HaveOccurred())
				err = test_utils.AssignHomeToProfile(ctx, collProfiles, profileRes.ID, home.ID)
				Expect(err).ShouldNot(HaveOccurred())
				err = test_utils.AssignHomeToProfile(ctx, collProfiles, profileRes.ID, home2.ID)
				Expect(err).ShouldNot(HaveOccurred())

				// assign `deviceController` and `deviceSensor` to `room1` of `home2`
				err = test_utils.AssignDeviceToHomeAndRoom(ctx, collHomes, home2.ID, home2.Rooms[0].ID, deviceController.ID)
				Expect(err).ShouldNot(HaveOccurred())
				err = test_utils.AssignDeviceToHomeAndRoom(ctx, collHomes, home2.ID, home2.Rooms[0].ID, deviceSensor.ID)
				Expect(err).ShouldNot(HaveOccurred())
				homeFromDb, err := test_utils.FindOneById[models.Home](ctx, collHomes, home.ID)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(homeFromDb.Rooms).To(HaveLen(2))
				Expect(homeFromDb.Rooms[0].Devices).To(HaveLen(0))
				Expect(homeFromDb.Rooms[1].Devices).To(HaveLen(0))
				home2FromDb, err := test_utils.FindOneById[models.Home](ctx, collHomes, home2.ID)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(home2FromDb.Rooms).To(HaveLen(1))
				Expect(home2FromDb.Rooms[0].Devices).To(HaveLen(2))
				Expect(home2FromDb.Rooms[0].Devices).To(ContainElements(deviceController.ID, deviceSensor.ID))

				// assign `deviceController` to `room1` of `home1`
				assignDeviceReq := api.AssignDeviceReq{
					HomeId: home.ID.Hex(),
					RoomId: home.Rooms[0].ID.Hex(),
				}
				var assignDeviceBug bytes.Buffer
				err = json.NewEncoder(&assignDeviceBug).Encode(assignDeviceReq)
				Expect(err).ShouldNot(HaveOccurred())
				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodPut, "/api/devices/"+deviceController.ID.Hex(), &assignDeviceBug)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusOK))
				Expect(recorder.Body.String()).To(Equal(`{"message":"device has been assigned to room"}`))

				// `deviceController` should be assigned only to `room1` of `home1` and not also to `room1` of `home2`
				homeFromDb, err = test_utils.FindOneById[models.Home](ctx, collHomes, home.ID)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(homeFromDb.Rooms[0].Devices).To(HaveLen(1))
				Expect(homeFromDb.Rooms[0].Devices[0]).To(Equal(deviceController.ID))
				home2FromDb, err = test_utils.FindOneById[models.Home](ctx, collHomes, home2.ID)
				Expect(err).ShouldNot(HaveOccurred())
				// `deviceController` should be removed from `home2`, but `deviceSensor` should still be there
				Expect(home2FromDb.Rooms[0].Devices).To(HaveLen(1))
				Expect(home2FromDb.Rooms[0].Devices[0]).To(Equal(deviceSensor.ID))

			})
		})

		// this case shouldn't happen in real scenarios, however I want to be sure that
		// this API won't change other profiles.
		When("device is already assigned to another home of another profile (weird situation)", func() {
			It("should assign device to home+room without cleaning other profiles successfully", func() {
				jwtToken, cookieSession := test_utils.GetJwt(router)
				profileRes := test_utils.GetLoggedProfile(router, jwtToken, cookieSession)

				// there is no need to add another profile to cover this scenario, because I can add other stuff in db
				// without assigning to the profile (in this case `home2`)
				err := test_utils.AssignDeviceToProfile(ctx, collProfiles, profileRes.ID, deviceController.ID)
				Expect(err).ShouldNot(HaveOccurred())
				err = test_utils.AssignHomeToProfile(ctx, collProfiles, profileRes.ID, home.ID)
				Expect(err).ShouldNot(HaveOccurred())

				// assign `deviceController` and `deviceSensor` to `room1` of `home2`
				err = test_utils.AssignDeviceToHomeAndRoom(ctx, collHomes, home2.ID, home2.Rooms[0].ID, deviceController.ID)
				Expect(err).ShouldNot(HaveOccurred())
				err = test_utils.AssignDeviceToHomeAndRoom(ctx, collHomes, home2.ID, home2.Rooms[0].ID, deviceSensor.ID)
				Expect(err).ShouldNot(HaveOccurred())
				homeFromDb, err := test_utils.FindOneById[models.Home](ctx, collHomes, home.ID)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(homeFromDb.Rooms).To(HaveLen(2))
				Expect(homeFromDb.Rooms[0].Devices).To(HaveLen(0))
				Expect(homeFromDb.Rooms[1].Devices).To(HaveLen(0))
				home2FromDb, err := test_utils.FindOneById[models.Home](ctx, collHomes, home2.ID)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(home2FromDb.Rooms).To(HaveLen(1))
				Expect(home2FromDb.Rooms[0].Devices).To(HaveLen(2))
				Expect(home2FromDb.Rooms[0].Devices).To(ContainElements(deviceController.ID, deviceSensor.ID))

				// assign `deviceController` to `room1` of `home1`
				assignDeviceReq := api.AssignDeviceReq{
					HomeId: home.ID.Hex(),
					RoomId: home.Rooms[0].ID.Hex(),
				}
				var assignDeviceBug bytes.Buffer
				err = json.NewEncoder(&assignDeviceBug).Encode(assignDeviceReq)
				Expect(err).ShouldNot(HaveOccurred())
				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodPut, "/api/devices/"+deviceController.ID.Hex(), &assignDeviceBug)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusOK))
				Expect(recorder.Body.String()).To(Equal(`{"message":"device has been assigned to room"}`))

				// `deviceController` should be assigned to `room1` of `home1` as requested
				homeFromDb, err = test_utils.FindOneById[models.Home](ctx, collHomes, home.ID)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(homeFromDb.Rooms[0].Devices).To(HaveLen(1))
				Expect(homeFromDb.Rooms[0].Devices[0]).To(Equal(deviceController.ID))
				home2FromDb, err = test_utils.FindOneById[models.Home](ctx, collHomes, home2.ID)
				Expect(err).ShouldNot(HaveOccurred())
				// both `deviceController` and `deviceSensor` should still be there!
				// This situation is bad, because `deviceController` should not be assigned to 2 players
				// However, I don't want that this API will modify other players, so I prefer to leave `deviceController` on both players
				// In production, this situation isn't possible, but I decided to test this situation to prevent unexpected behaviour.
				Expect(home2FromDb.Rooms[0].Devices).To(HaveLen(2))
				Expect(home2FromDb.Rooms[0].Devices).To(ContainElements(deviceController.ID, deviceSensor.ID))

			})
		})
	})
})
