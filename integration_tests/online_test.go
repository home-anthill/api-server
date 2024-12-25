package integration_tests

import (
	"api-server/db"
	"api-server/initialization"
	"api-server/models"
	"api-server/testuutils"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"time"
)

var onlineCurrentDate = time.Now()
var onlineUUID = uuid.NewString()
var mockedProfileAPIToken = "2ee7e6d0-c216-4548-bd78-fa3b04bb5fef"

var _ = Describe("Online", func() {
	var ctx context.Context
	var logger *zap.SugaredLogger
	var router *gin.Engine
	var client *mongo.Client
	var collProfiles *mongo.Collection
	var collHomes *mongo.Collection
	var collDevices *mongo.Collection
	var httpMockServer *httptest.Server

	var deviceSensor = models.Device{
		ID:           primitive.NewObjectID(),
		Mac:          "AA:22:33:44:55:BB",
		Manufacturer: "test",
		Model:        "poweroutage",
		UUID:         onlineUUID,
		Features: []models.Feature{{
			UUID:   uuid.NewString(),
			Type:   "sensor",
			Name:   "poweroutage",
			Enable: true,
			Order:  1,
			Unit:   "",
		}},
		CreatedAt:  onlineCurrentDate,
		ModifiedAt: onlineCurrentDate,
	}

	getSensorOnlineHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(getOnlineJSONResponse(mockedProfileAPIToken, onlineCurrentDate, onlineCurrentDate)))
	})

	BeforeEach(func() {
		logger, router, ctx, client = initialization.Start()
		defer logger.Sync()

		collProfiles = db.GetCollections(client).Profiles
		collHomes = db.GetCollections(client).Homes
		collDevices = db.GetCollections(client).Devices

		err := os.Setenv("SINGLE_USER_LOGIN_EMAIL", "test@test.com")
		Expect(err).ShouldNot(HaveOccurred())

		// --------- start an HTTP server ---------
		//registerResponse := `[{"id": 123412341234123412341234, "code": 200}]`
		mux := http.NewServeMux()
		mux.HandleFunc("/online/"+onlineUUID, getSensorOnlineHandler)
		httpListener, errHTTP := net.Listen("tcp", "localhost:8000")
		logger.Infof("register_test - HTTP client listening at %s", httpListener.Addr().String())
		Expect(errHTTP).ShouldNot(HaveOccurred())
		httpMockServer = httptest.NewUnstartedServer(mux)
		// NewUnstartedServer creates a httpListener, so we need to Close that
		// httpListener and replace it with the one we created.
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

	Context("calling online api GET", func() {
		BeforeEach(func() {
			err := testuutils.InsertOne(ctx, collDevices, deviceSensor)
			Expect(err).ShouldNot(HaveOccurred())
		})

		When("profile owns a sensor", func() {
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

				Expect(online.UUID).To(Equal(onlineUUID))
				Expect(online.APIToken).To(Equal(mockedProfileAPIToken))
				Expect(online.CreatedAt.UnixMilli()).To(Equal(onlineCurrentDate.UnixMilli()))
				Expect(online.ModifiedAt.UnixMilli()).To(Equal(onlineCurrentDate.UnixMilli()))
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

				unexistingDeviceID := primitive.NewObjectID()

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
	})
})

func getOnlineJSONResponse(apiToken string, createDate time.Time, modDate time.Time) string {
	return `{"apiToken": "` + apiToken + `", "createdAt": ` + fmt.Sprintf("%v", createDate.UnixMilli()) + `, "modifiedAt": ` + fmt.Sprintf("%v", modDate.UnixMilli()) + `}`
}
