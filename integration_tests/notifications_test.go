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

var _ = Describe("Notifications", func() {
	var currentDate = time.Now()
	var notificationDeviceUUID = uuid.NewString()
	var notificationFeatureUUID = uuid.NewString()
	var mockedProfileAPIToken = "2ee7e6d0-c216-4548-bd78-fa3b04bb5fef"

	var ctx context.Context
	var logger *zap.SugaredLogger
	var router *gin.Engine
	var client *mongo.Client
	var collProfiles *mongo.Collection
	var collHomes *mongo.Collection
	var collDevices *mongo.Collection
	var httpMockServer *httptest.Server
	var notificationsResponseStatus int
	var notificationsResponseBody string

	var deviceSensor = models.Device{
		ID:           bson.NewObjectID(),
		Mac:          "AA:22:33:44:55:BB",
		Manufacturer: "test",
		Model:        "power-outage",
		UUID:         notificationDeviceUUID,
		Features: []models.Feature{{
			UUID:   notificationFeatureUUID,
			Type:   "sensor",
			Name:   "online",
			Enable: true,
			Order:  1,
			Unit:   "-",
		}},
		CreatedAt:  currentDate,
		ModifiedAt: currentDate,
	}

	getNotificationsHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/notifications/"+mockedProfileAPIToken {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(notificationsResponseStatus)
		_, _ = w.Write([]byte(notificationsResponseBody))
	})

	BeforeEach(func() {
		logger, router, client = initialization.MustStart()
		ctx = context.Background()
		defer logger.Sync()

		collProfiles = db.GetCollections(client).Profiles
		collHomes = db.GetCollections(client).Homes
		collDevices = db.GetCollections(client).Devices

		err := os.Setenv("LIMIT_TO_USER_EMAILS", "test@test.com")
		Expect(err).ShouldNot(HaveOccurred())

		notificationsResponseStatus = http.StatusOK
		notificationsResponseBody = getNotificationsJSONResponse(currentDate, deviceSensor.UUID, notificationFeatureUUID)

		mux := http.NewServeMux()
		mux.HandleFunc("/notifications/"+mockedProfileAPIToken, getNotificationsHandler)
		httpListener, errHTTP := net.Listen("tcp", "localhost:8089")
		logger.Infof("notifications_test - HTTP client listening at %s", httpListener.Addr().String())
		Expect(errHTTP).ShouldNot(HaveOccurred())
		httpMockServer = httptest.NewUnstartedServer(mux)
		httpMockServer.Listener.Close()
		httpMockServer.Listener = httpListener
		go func() {
			httpMockServer.Start()
		}()
	})

	AfterEach(func() {
		httpMockServer.Close()
		testuutils.DropAllCollections(ctx, collProfiles, collHomes, collDevices)
	})

	Context("calling notification api GET", func() {
		When("profile is authenticated", func() {
			It("should get logged profile notifications from online service", func() {
				jwtToken, cookieSession := testuutils.GetJwt(router)
				profileRes := testuutils.GetLoggedProfile(router, jwtToken, cookieSession)
				err := testuutils.SetAPITokenToProfile(ctx, collProfiles, profileRes.ID, mockedProfileAPIToken)
				Expect(err).ShouldNot(HaveOccurred())
				err = testuutils.InsertOne(ctx, collDevices, deviceSensor)
				Expect(err).ShouldNot(HaveOccurred())
				err = testuutils.AssignDeviceToProfile(ctx, collProfiles, profileRes.ID, deviceSensor.ID)
				Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, "/api/notifications", nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)

				Expect(recorder.Code).To(Equal(http.StatusOK))
				var response struct {
					Notifications []struct {
						ID                string          `json:"id"`
						SentAt            uint64          `json:"sentAt"`
						Title             string          `json:"title"`
						Body              string          `json:"body"`
						DeviceCount       uint64          `json:"deviceCount"`
						Devices           []models.Device `json:"devices"`
						Provider          string          `json:"provider"`
						ProviderMessageID string          `json:"providerMessageId"`
					} `json:"notifications"`
				}
				err = json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(response.Notifications).To(HaveLen(1))
				Expect(response.Notifications[0].ID).To(Equal("test-notification-a"))
				Expect(response.Notifications[0].DeviceCount).To(Equal(uint64(1)))
				Expect(response.Notifications[0].Devices).To(HaveLen(1))
				Expect(response.Notifications[0].Devices[0].ID).To(Equal(deviceSensor.ID))
				Expect(response.Notifications[0].Devices[0].UUID).To(Equal(deviceSensor.UUID))
				Expect(response.Notifications[0].Devices[0].Features).To(Equal(deviceSensor.Features))
			})
		})

		When("authorization is missing", func() {
			It("should return an error", func() {
				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, "/api/notifications", nil)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)

				Expect(recorder.Code).To(Equal(http.StatusUnauthorized))
				Expect(recorder.Body.String()).To(Equal(`{"error":"authorization header not found"}`))
			})
		})

		When("profile session points to a missing profile", func() {
			It("should return an error", func() {
				jwtToken, cookieSession := testuutils.GetJwt(router)
				profileRes := testuutils.GetLoggedProfile(router, jwtToken, cookieSession)
				_, err := collProfiles.DeleteOne(ctx, bson.M{"_id": profileRes.ID})
				Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, "/api/notifications", nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)

				Expect(recorder.Code).To(Equal(http.StatusUnauthorized))
				Expect(recorder.Body.String()).To(Equal(`{"error":"cannot find profile"}`))
			})
		})

		When("profile apiToken cannot be loaded", func() {
			It("should return an error", func() {
				jwtToken, cookieSession := testuutils.GetJwt(router)

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, "/api/notifications", nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)

				Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
				Expect(recorder.Body.String()).To(Equal(`{"error":"Cannot get notifications"}`))
			})
		})

		When("profile apiToken has an invalid format", func() {
			It("should return an error before calling online service", func() {
				jwtToken, cookieSession := testuutils.GetJwt(router)
				profileRes := testuutils.GetLoggedProfile(router, jwtToken, cookieSession)
				err := testuutils.SetAPITokenToProfile(ctx, collProfiles, profileRes.ID, "not-a-uuid")
				Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, "/api/notifications", nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)

				Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
				Expect(recorder.Body.String()).To(Equal(`{"error":"Cannot get notifications"}`))
			})
		})

		When("online service returns an error", func() {
			It("should return an error", func() {
				notificationsResponseStatus = http.StatusBadGateway
				notificationsResponseBody = `{"error":"online unavailable"}`
				jwtToken, cookieSession := testuutils.GetJwt(router)
				profileRes := testuutils.GetLoggedProfile(router, jwtToken, cookieSession)
				err := testuutils.SetAPITokenToProfile(ctx, collProfiles, profileRes.ID, mockedProfileAPIToken)
				Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, "/api/notifications", nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)

				Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
				Expect(recorder.Body.String()).To(Equal(`{"error":"Cannot get notifications"}`))
			})
		})

		When("online service returns invalid JSON", func() {
			It("should return an error", func() {
				notificationsResponseBody = `not-json`
				jwtToken, cookieSession := testuutils.GetJwt(router)
				profileRes := testuutils.GetLoggedProfile(router, jwtToken, cookieSession)
				err := testuutils.SetAPITokenToProfile(ctx, collProfiles, profileRes.ID, mockedProfileAPIToken)
				Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, "/api/notifications", nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)

				Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
				Expect(recorder.Body.String()).To(Equal(`{"error":"Cannot get notifications response"}`))
			})
		})

		When("matching notification device cannot be decoded", func() {
			It("should return an enrich notifications error", func() {
				jwtToken, cookieSession := testuutils.GetJwt(router)
				profileRes := testuutils.GetLoggedProfile(router, jwtToken, cookieSession)
				err := testuutils.SetAPITokenToProfile(ctx, collProfiles, profileRes.ID, mockedProfileAPIToken)
				Expect(err).ShouldNot(HaveOccurred())
				err = testuutils.InsertOne(ctx, collDevices, bson.M{
					"_id":        deviceSensor.ID,
					"uuid":       deviceSensor.UUID,
					"createdAt":  "not-a-date",
					"modifiedAt": currentDate,
				})
				Expect(err).ShouldNot(HaveOccurred())
				err = testuutils.AssignDeviceToProfile(ctx, collProfiles, profileRes.ID, deviceSensor.ID)
				Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, "/api/notifications", nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)

				Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
				Expect(recorder.Body.String()).To(Equal(`{"error":"Cannot get notifications response"}`))
			})
		})

		When("online service returns a device outside the logged profile", func() {
			It("should not enrich the notification with that device", func() {
				jwtToken, cookieSession := testuutils.GetJwt(router)
				profileRes := testuutils.GetLoggedProfile(router, jwtToken, cookieSession)
				err := testuutils.SetAPITokenToProfile(ctx, collProfiles, profileRes.ID, mockedProfileAPIToken)
				Expect(err).ShouldNot(HaveOccurred())
				err = testuutils.InsertOne(ctx, collDevices, deviceSensor)
				Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, "/api/notifications", nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)

				Expect(recorder.Code).To(Equal(http.StatusOK))
				var response struct {
					Notifications []struct {
						Devices []models.Device `json:"devices"`
					} `json:"notifications"`
				}
				err = json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(response.Notifications).To(HaveLen(1))
				Expect(response.Notifications[0].Devices).To(BeEmpty())
			})
		})
	})
})

func getNotificationsJSONResponse(sentAt time.Time, deviceUUID string, featureUUID string) string {
	return `{"notifications":[{"id":"test-notification-a","sentAt":` +
		fmt.Sprintf("%v", sentAt.UnixMilli()) +
		`,"title":"home anthill","body":"Device is offline","deviceCount":1,"devices":[{"deviceUuid":"` + deviceUUID +
		`","featureUuid":"` + featureUUID +
		`","createdAt":1710000000001,"modifiedAt":1710000000002}],"provider":"fcm","providerMessageId":"projects/home-anthill/messages/message-a"}]}`
}
