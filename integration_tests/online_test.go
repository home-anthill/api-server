package integration_tests

import (
	"api-server/db"
	"api-server/initialization"
	"api-server/models"
	"api-server/testuutils"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"
)

var _ = Describe("Online", func() {
	var currentDate = time.Now()
	var onlineDeviceUUID = uuid.NewString()
	var onlineFeatureUUID = uuid.NewString()
	var onlineDeviceUUID2 = uuid.NewString()
	var onlineFeatureUUID2 = uuid.NewString()
	var mockedProfileAPIToken = "2ee7e6d0-c216-4548-bd78-fa3b04bb5fef"

	var ctx context.Context
	var logger *zap.SugaredLogger
	var router *gin.Engine
	var client *mongo.Client
	var collProfiles *mongo.Collection
	var collHomes *mongo.Collection
	var collDevices *mongo.Collection
	var httpMockServer *httptest.Server
	var oldHTTPOnlineServer string
	var oldHTTPOnlinePort string
	var onlineResponseStatus int
	var onlineResponseBody string

	var deviceSensor = models.Device{
		ID:           bson.NewObjectID(),
		Mac:          "AA:22:33:44:55:BB",
		Manufacturer: "test",
		Model:        "environment-sensor",
		UUID:         onlineDeviceUUID,
		Features: []models.Feature{{
			UUID:   onlineFeatureUUID,
			Type:   "sensor",
			Name:   "online",
			Enable: true,
			Order:  1,
			Unit:   "-",
		}, {
			UUID:   uuid.NewString(),
			Type:   "sensor",
			Name:   "temperature",
			Enable: true,
			Order:  2,
			Unit:   "C",
		}},
		CreatedAt:  currentDate,
		ModifiedAt: currentDate,
	}

	var deviceSensorNoOnline = models.Device{
		ID:           bson.NewObjectID(),
		Mac:          "FF:22:33:44:55:CC",
		Manufacturer: "test",
		Model:        "pir",
		UUID:         uuid.NewString(),
		Features: []models.Feature{{
			UUID:   uuid.NewString(),
			Type:   "sensor",
			Name:   "motion",
			Enable: true,
			Order:  1,
			Unit:   "-",
		}},
		CreatedAt:  currentDate,
		ModifiedAt: currentDate,
	}

	var deviceSensor2 = models.Device{
		ID:           bson.NewObjectID(),
		Mac:          "CC:22:33:44:55:DD",
		Manufacturer: "test",
		Model:        "air-quality-sensor",
		UUID:         onlineDeviceUUID2,
		Features: []models.Feature{{
			UUID:   onlineFeatureUUID2,
			Type:   "sensor",
			Name:   "online",
			Enable: true,
			Order:  1,
			Unit:   "-",
		}, {
			UUID:   uuid.NewString(),
			Type:   "sensor",
			Name:   "humidity",
			Enable: true,
			Order:  2,
			Unit:   "%",
		}},
		CreatedAt:  currentDate,
		ModifiedAt: currentDate,
	}

	getSensorOnlineHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(onlineResponseStatus)
		_, _ = w.Write([]byte(onlineResponseBody))
	})

	BeforeEach(func() {
		onlineResponseStatus = http.StatusOK
		onlineResponseBody = getOnlineJSONResponse(mockedProfileAPIToken, currentDate, currentDate)

		// --------- start an HTTP server ---------
		mux := http.NewServeMux()
		mux.HandleFunc("/online/"+onlineDeviceUUID+"/features/"+onlineFeatureUUID, getSensorOnlineHandler)
		mux.HandleFunc("/online/"+onlineDeviceUUID2+"/features/"+onlineFeatureUUID2, getSensorOnlineHandler)
		httpListener, errHTTP := net.Listen("tcp", "127.0.0.1:0")
		Expect(errHTTP).ShouldNot(HaveOccurred())

		host, port, err := net.SplitHostPort(httpListener.Addr().String())
		Expect(err).ShouldNot(HaveOccurred())
		oldHTTPOnlineServer = os.Getenv("HTTP_ALARM_SERVER")
		oldHTTPOnlinePort = os.Getenv("HTTP_ALARM_PORT")
		err = os.Setenv("HTTP_ALARM_SERVER", "http://"+host)
		Expect(err).ShouldNot(HaveOccurred())
		err = os.Setenv("HTTP_ALARM_PORT", port)
		Expect(err).ShouldNot(HaveOccurred())

		logger, router, client = initialization.MustStart()
		ctx = context.Background()
		defer logger.Sync()

		collProfiles = db.GetCollections(client).Profiles
		collHomes = db.GetCollections(client).Homes
		collDevices = db.GetCollections(client).Devices

		err = os.Setenv("LIMIT_TO_USER_EMAILS", "test@test.com")
		Expect(err).ShouldNot(HaveOccurred())

		logger.Infof("online_test - HTTP client listening at %s", httpListener.Addr().String())
		httpMockServer = httptest.NewUnstartedServer(mux)
		// NewUnstartedServer creates an httpListener, so we need to Close that
		// httpListener and replace it with the one we created.
		httpMockServer.Listener.Close()
		httpMockServer.Listener = httpListener
		go func() {
			httpMockServer.Start()
		}()
	})

	AfterEach(func() {
		if httpMockServer != nil {
			httpMockServer.Close()
		}
		if ctx != nil {
			testuutils.DropAllCollections(ctx, collProfiles, collHomes, collDevices)
		}
		err := os.Setenv("HTTP_ALARM_SERVER", oldHTTPOnlineServer)
		Expect(err).ShouldNot(HaveOccurred())
		err = os.Setenv("HTTP_ALARM_PORT", oldHTTPOnlinePort)
		Expect(err).ShouldNot(HaveOccurred())
	})

	Context("calling online api GET", func() {
		BeforeEach(func() {
			err := testuutils.InsertOne(ctx, collDevices, deviceSensor)
			Expect(err).ShouldNot(HaveOccurred())
			err = testuutils.InsertOne(ctx, collDevices, deviceSensorNoOnline)
			Expect(err).ShouldNot(HaveOccurred())
			err = testuutils.InsertOne(ctx, collDevices, deviceSensor2)
			Expect(err).ShouldNot(HaveOccurred())
		})

		When("profile owns a sensor with online feature", func() {
			It("should get online", func() {
				jwtToken, cookieSession := testuutils.GetJwt(router)
				profileRes := testuutils.GetLoggedProfile(router, jwtToken, cookieSession)
				// set mocked APIToken to the logged profile
				err := testuutils.SetAPITokenToProfile(ctx, collProfiles, profileRes.ID, mockedProfileAPIToken)
				Expect(err).ShouldNot(HaveOccurred())
				// assign mocked sensor device to the logged profile
				err = testuutils.AssignDeviceToProfile(ctx, collProfiles, profileRes.ID, deviceSensor.ID)
				Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, "/api/online/"+deviceSensor.ID.Hex(), nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusOK))
				var online *models.Online
				err = json.Unmarshal(recorder.Body.Bytes(), &online)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(online.CreatedAt.UnixMilli()).To(Equal(currentDate.UnixMilli()))
				Expect(online.ModifiedAt.UnixMilli()).To(Equal(currentDate.UnixMilli()))
			})
		})

		When("profile owns online-capable devices", func() {
			It("should get online statuses with devices and features", func() {
				jwtToken, cookieSession := testuutils.GetJwt(router)
				profileRes := testuutils.GetLoggedProfile(router, jwtToken, cookieSession)
				err := testuutils.SetAPITokenToProfile(ctx, collProfiles, profileRes.ID, mockedProfileAPIToken)
				Expect(err).ShouldNot(HaveOccurred())
				err = testuutils.AssignDeviceToProfile(ctx, collProfiles, profileRes.ID, deviceSensor.ID)
				Expect(err).ShouldNot(HaveOccurred())
				err = testuutils.AssignDeviceToProfile(ctx, collProfiles, profileRes.ID, deviceSensorNoOnline.ID)
				Expect(err).ShouldNot(HaveOccurred())
				err = testuutils.AssignDeviceToProfile(ctx, collProfiles, profileRes.ID, deviceSensor2.ID)
				Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, "/api/online", nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusOK))

				var onlineStatuses []models.OnlineDeviceStatus
				err = json.Unmarshal(recorder.Body.Bytes(), &onlineStatuses)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(onlineStatuses).To(HaveLen(2))
				Expect(onlineStatuses[0].CreatedAt.UnixMilli()).To(Equal(currentDate.UnixMilli()))
				Expect(onlineStatuses[0].ModifiedAt.UnixMilli()).To(Equal(currentDate.UnixMilli()))
				Expect(onlineStatuses[0].CurrentTime.UnixMilli()).To(Equal(currentDate.UnixMilli()))
				Expect([]string{onlineStatuses[0].Device.UUID, onlineStatuses[1].Device.UUID}).To(ConsistOf(deviceSensor.UUID, deviceSensor2.UUID))
				Expect([]string{onlineStatuses[0].Feature.UUID, onlineStatuses[1].Feature.UUID}).To(ConsistOf(onlineFeatureUUID, onlineFeatureUUID2))
			})
		})

		When("profile does not own any device", func() {
			It("should get an empty profile online statuses response", func() {
				jwtToken, cookieSession := testuutils.GetJwt(router)

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, "/api/online", nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusOK))

				var onlineStatuses []models.OnlineDeviceStatus
				err := json.Unmarshal(recorder.Body.Bytes(), &onlineStatuses)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(onlineStatuses).To(BeEmpty())
			})
		})

		When("profile owns only devices without online feature", func() {
			It("should get an empty profile online statuses response", func() {
				jwtToken, cookieSession := testuutils.GetJwt(router)
				profileRes := testuutils.GetLoggedProfile(router, jwtToken, cookieSession)
				err := testuutils.AssignDeviceToProfile(ctx, collProfiles, profileRes.ID, deviceSensorNoOnline.ID)
				Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, "/api/online", nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusOK))

				var onlineStatuses []models.OnlineDeviceStatus
				err = json.Unmarshal(recorder.Body.Bytes(), &onlineStatuses)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(onlineStatuses).To(BeEmpty())
			})
		})

		When("profile owns a sensor without online feature", func() {
			It("should return an error, because sensor hasn't online feature", func() {
				jwtToken, cookieSession := testuutils.GetJwt(router)
				profileRes := testuutils.GetLoggedProfile(router, jwtToken, cookieSession)
				// set mocked APIToken to the logged profile
				err := testuutils.SetAPITokenToProfile(ctx, collProfiles, profileRes.ID, mockedProfileAPIToken)
				Expect(err).ShouldNot(HaveOccurred())
				// assign mocked sensor device to the logged profile
				err = testuutils.AssignDeviceToProfile(ctx, collProfiles, profileRes.ID, deviceSensorNoOnline.ID)
				Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, "/api/online/"+deviceSensorNoOnline.ID.Hex(), nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
				Expect(recorder.Body.String()).To(Equal(`{"error":"cannot find online feature in this device"}`))
			})
		})

		When("you pass bad inputs", func() {
			It("should return an error, because ...", func() {
				jwtToken, cookieSession := testuutils.GetJwt(router)

				badDeviceID := "bad_device_id"

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, "/api/online/"+badDeviceID, nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
				Expect(recorder.Body.String()).To(Equal(`{"error":"wrong format of the path param 'id'"}`))
			})
		})

		When("profile don't own any device", func() {
			It("should return an error, because you can get only devices owned by profile", func() {
				jwtToken, cookieSession := testuutils.GetJwt(router)
				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, "/api/online/"+deviceSensor.ID.Hex(), nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
				Expect(recorder.Body.String()).To(Equal(`{"error":"this device is not in your profile"}`))
			})
		})

		When("profile owns a device not in 'devices' collection", func() {
			It("should return an error, because device doesn't exist", func() {
				jwtToken, cookieSession := testuutils.GetJwt(router)
				profileRes := testuutils.GetLoggedProfile(router, jwtToken, cookieSession)

				unexistingDeviceID := bson.NewObjectID()

				err := testuutils.AssignDeviceToProfile(ctx, collProfiles, profileRes.ID, unexistingDeviceID)
				Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, "/api/online/"+unexistingDeviceID.Hex(), nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
				Expect(recorder.Body.String()).To(Equal(`{"error":"cannot find device"}`))
			})
		})

		When("profile owns an online sensor with an invalid UUID", func() {
			It("should return an error before calling the alarm service", func() {
				jwtToken, cookieSession := testuutils.GetJwt(router)
				profileRes := testuutils.GetLoggedProfile(router, jwtToken, cookieSession)

				deviceBadUUID := cloneOnlineTestDevice(deviceSensor)
				deviceBadUUID.ID = bson.NewObjectID()
				deviceBadUUID.UUID = "not-a-uuid"
				err := testuutils.InsertOne(ctx, collDevices, deviceBadUUID)
				Expect(err).ShouldNot(HaveOccurred())
				err = testuutils.AssignDeviceToProfile(ctx, collProfiles, profileRes.ID, deviceBadUUID.ID)
				Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, "/api/online/"+deviceBadUUID.ID.Hex(), nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
				Expect(recorder.Body.String()).To(Equal(`{"error":"Cannot get online"}`))
			})
		})

		When("profile owns an online-capable device with an invalid feature UUID", func() {
			It("should return an error before calling the alarm service", func() {
				jwtToken, cookieSession := testuutils.GetJwt(router)
				profileRes := testuutils.GetLoggedProfile(router, jwtToken, cookieSession)

				deviceBadFeatureUUID := cloneOnlineTestDevice(deviceSensor)
				deviceBadFeatureUUID.ID = bson.NewObjectID()
				deviceBadFeatureUUID.Mac = "DD:22:33:44:55:EE"
				deviceBadFeatureUUID.Features[0].UUID = "not-a-uuid"
				err := testuutils.InsertOne(ctx, collDevices, deviceBadFeatureUUID)
				Expect(err).ShouldNot(HaveOccurred())
				err = testuutils.AssignDeviceToProfile(ctx, collProfiles, profileRes.ID, deviceBadFeatureUUID.ID)
				Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, "/api/online/"+deviceBadFeatureUUID.ID.Hex(), nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
				Expect(recorder.Body.String()).To(Equal(`{"error":"Cannot get online"}`))
			})
		})

		When("alarm service returns an error for a single device lookup", func() {
			It("should return a remote online error", func() {
				onlineResponseStatus = http.StatusBadGateway
				onlineResponseBody = `{"error":"online unavailable"}`

				jwtToken, cookieSession := testuutils.GetJwt(router)
				profileRes := testuutils.GetLoggedProfile(router, jwtToken, cookieSession)
				err := testuutils.AssignDeviceToProfile(ctx, collProfiles, profileRes.ID, deviceSensor.ID)
				Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, "/api/online/"+deviceSensor.ID.Hex(), nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
				Expect(recorder.Body.String()).To(Equal(`{"error":"Cannot get online"}`))
			})
		})

		When("alarm service returns invalid JSON for a single device lookup", func() {
			It("should return an online response parsing error", func() {
				onlineResponseBody = `not-json`

				jwtToken, cookieSession := testuutils.GetJwt(router)
				profileRes := testuutils.GetLoggedProfile(router, jwtToken, cookieSession)
				err := testuutils.AssignDeviceToProfile(ctx, collProfiles, profileRes.ID, deviceSensor.ID)
				Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, "/api/online/"+deviceSensor.ID.Hex(), nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
				Expect(recorder.Body.String()).To(Equal(`{"error":"Cannot get online response"}`))
			})
		})

		When("profile online lookup includes a device with an invalid UUID", func() {
			It("should return an error before calling the alarm service", func() {
				jwtToken, cookieSession := testuutils.GetJwt(router)
				profileRes := testuutils.GetLoggedProfile(router, jwtToken, cookieSession)

				deviceBadUUID := cloneOnlineTestDevice(deviceSensor)
				deviceBadUUID.ID = bson.NewObjectID()
				deviceBadUUID.Mac = "EE:22:33:44:55:FF"
				deviceBadUUID.UUID = "not-a-uuid"
				err := testuutils.InsertOne(ctx, collDevices, deviceBadUUID)
				Expect(err).ShouldNot(HaveOccurred())
				err = testuutils.AssignDeviceToProfile(ctx, collProfiles, profileRes.ID, deviceBadUUID.ID)
				Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, "/api/online", nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
				Expect(recorder.Body.String()).To(Equal(`{"error":"Cannot get online"}`))
			})
		})

		When("alarm service returns an error for profile online lookup", func() {
			It("should return a remote online error", func() {
				onlineResponseStatus = http.StatusBadGateway
				onlineResponseBody = `{"error":"online unavailable"}`

				jwtToken, cookieSession := testuutils.GetJwt(router)
				profileRes := testuutils.GetLoggedProfile(router, jwtToken, cookieSession)
				err := testuutils.AssignDeviceToProfile(ctx, collProfiles, profileRes.ID, deviceSensor.ID)
				Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, "/api/online", nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
				Expect(recorder.Body.String()).To(Equal(`{"error":"Cannot get online"}`))
			})
		})

		When("alarm service returns invalid JSON for profile online lookup", func() {
			It("should return a remote online error", func() {
				onlineResponseBody = `not-json`

				jwtToken, cookieSession := testuutils.GetJwt(router)
				profileRes := testuutils.GetLoggedProfile(router, jwtToken, cookieSession)
				err := testuutils.AssignDeviceToProfile(ctx, collProfiles, profileRes.ID, deviceSensor.ID)
				Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, "/api/online", nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
				Expect(recorder.Body.String()).To(Equal(`{"error":"Cannot get online"}`))
			})
		})
	})
})

func getOnlineJSONResponse(apiToken string, createDate time.Time, modDate time.Time) string {
	return `{"apiToken": "` + apiToken + `", "createdAt": ` + fmt.Sprintf("%v", createDate.UnixMilli()) + `, "modifiedAt": ` + fmt.Sprintf("%v", modDate.UnixMilli()) + `, "currentTime": ` + fmt.Sprintf("%v", modDate.UnixMilli()) + `}`
}

func cloneOnlineTestDevice(device models.Device) models.Device {
	device.Features = append([]models.Feature(nil), device.Features...)
	return device
}
