package integration_tests

import (
	"api-server/api"
	"api-server/db"
	"api-server/initialization"
	"api-server/models"
	"api-server/testuutils"
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
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"time"
)

var githubUserTestmock = models.GitHub{
	ID:        123456,
	Login:     "Test",
	Name:      "Test Test",
	Email:     "test@test.com",
	AvatarURL: "https://avatars.githubusercontent.com/u/123456?v=4",
}

var _ = Describe("FCMToken", func() {
	var ctx context.Context
	var logger *zap.SugaredLogger
	var router *gin.Engine
	var client *mongo.Client
	var collProfiles *mongo.Collection
	var collHomes *mongo.Collection
	var collDevices *mongo.Collection
	var httpMockServer *httptest.Server

	currentDate := time.Now()
	profile := models.Profile{
		ID:         primitive.NewObjectID(),
		Github:     githubUserTestmock,
		APIToken:   uuid.NewString(),
		Homes:      []primitive.ObjectID{}, // empty slice of ObjectIDs
		Devices:    []primitive.ObjectID{}, // empty slice of ObjectIDs
		CreatedAt:  currentDate,
		ModifiedAt: currentDate,
	}
	keepAliveHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"alive": true}]`))
	})
	fcmTokenHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"code": 200}]`))
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
		mux.HandleFunc("/keepalive", keepAliveHandler)
		mux.HandleFunc("/fcmtoken", fcmTokenHandler)
		httpListener, errHTTP := net.Listen("tcp", "localhost:8000")
		logger.Infof("fcmtoken_test - HTTP client listening at %s", httpListener.Addr().String())
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

	Describe("calling fcmtoken api", func() {
		When("initializing a new smartphone with Firebase Cloud Messaging token", func() {

			It("should return a success", func() {
				By("with an existing profile with a valid apiToken")
				err := testuutils.InsertOne(ctx, collProfiles, profile)
				Expect(err).ShouldNot(HaveOccurred())

				initFCMTokenReq := api.InitFCMTokenReq{
					APIToken: profile.APIToken,
					FCMToken: "dTknleBlRLqEoWBMjiIr80:APA91bEs_Tf8dkrZ_eb872Ok--Up34Luqp1S4WZwzTGr6X1ag4PO4ksHwFFifNqTb1lhATzNcaVqDZ01kP35a0caOa6Akw4oCzYh0ElqL8msgjtZw2phLEk",
				}
				var buf bytes.Buffer
				err = json.NewEncoder(&buf).Encode(initFCMTokenReq)
				Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodPost, "/api/fcmtoken", &buf)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusOK))
				Expect(recorder.Body.String()).To(Equal(`{"message":"FCMToken assigned to APIToken"}`))
			})
		})

		When("you pass bad inputs", func() {
			It("should return an error, if body is missing", func() {
				err := testuutils.InsertOne(ctx, collProfiles, profile)
				Expect(err).ShouldNot(HaveOccurred())
				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodPost, "/api/fcmtoken", nil)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
				Expect(recorder.Body.String()).To(Equal(`{"error":"invalid request payload"}`))
			})

			It("should return an error, if body is not valid", func() {
				err := testuutils.InsertOne(ctx, collProfiles, profile)
				Expect(err).ShouldNot(HaveOccurred())

				initFCMTokenReq := api.InitFCMTokenReq{
					APIToken: "1234", // not valid, because must be an UUIDv4
					FCMToken: "abcd",
				}
				var buf bytes.Buffer
				err = json.NewEncoder(&buf).Encode(initFCMTokenReq)
				Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodPost, "/api/fcmtoken", &buf)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
				Expect(recorder.Body.String()).To(Equal(`{"error":"invalid request body, these fields are not valid: apitoken"}`))
			})

			It("should return an error, if apiToken doesn't exist", func() {
				err := testuutils.InsertOne(ctx, collProfiles, profile)
				Expect(err).ShouldNot(HaveOccurred())

				unknownAPIToken := uuid.NewString()
				initFCMTokenReq := api.InitFCMTokenReq{
					APIToken: unknownAPIToken,
					FCMToken: "abcd",
				}
				var buf bytes.Buffer
				err = json.NewEncoder(&buf).Encode(initFCMTokenReq)
				Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodPost, "/api/fcmtoken", &buf)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
				Expect(recorder.Body.String()).To(Equal(`{"error":"cannot initialize FCM Token, profile token missing or not valid"}`))
			})
		})
	})
})
