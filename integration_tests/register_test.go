package integration_tests

import (
	"api-server/api"
	"api-server/api/grpc/register"
	"api-server/initialization"
	"api-server/models"
	"api-server/test_utils"
	"bytes"
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
	"google.golang.org/grpc"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"time"
)

var dbGithubUserTestmock = models.Github{
	ID:        123456,
	Login:     "Test",
	Name:      "Test Test",
	Email:     "test@test.com",
	AvatarURL: "https://avatars.githubusercontent.com/u/123456?v=4",
}

type registerGrpcStub struct {
	register.UnimplementedRegistrationServer
	ctx    context.Context
	logger *zap.SugaredLogger
}

func newRegisterGrpc(ctx context.Context, logger *zap.SugaredLogger) *registerGrpcStub {
	return &registerGrpcStub{
		ctx:    ctx,
		logger: logger,
	}
}

func (handler *registerGrpcStub) Register(ctx context.Context, in *register.RegisterRequest) (*register.RegisterReply, error) {
	fmt.Printf("register_test - Register - received = %#v\n", in)
	return &register.RegisterReply{Status: strconv.FormatInt(http.StatusOK, 10), Message: "Inserted"}, nil
}

var _ = Describe("Register", func() {
	var ctx context.Context
	var logger *zap.SugaredLogger
	var router *gin.Engine
	var collProfiles *mongo.Collection
	var collHomes *mongo.Collection
	var collDevices *mongo.Collection
	var grpcMockServer *grpc.Server
	var httpMockServer *httptest.Server

	currentDate := time.Now()
	profile := models.Profile{
		ID:         primitive.NewObjectID(),
		Github:     dbGithubUserTestmock,
		ApiToken:   uuid.NewString(),
		Homes:      []primitive.ObjectID{}, // empty slice of ObjectIDs
		Devices:    []primitive.ObjectID{}, // empty slice of ObjectIDs
		CreatedAt:  currentDate,
		ModifiedAt: currentDate,
	}
	keepAliveHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"alive": true}]`))
	})
	registerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"id": 123412341234123412341234, "code": 200}]`))
	})

	BeforeEach(func() {
		logger, router, ctx, collProfiles, collHomes, collDevices = initialization.Start()
		defer logger.Sync()

		err := os.Setenv("SINGLE_USER_LOGIN_EMAIL", "test@test.com")
		Expect(err).ShouldNot(HaveOccurred())

		// --------- start a gRPC server ---------
		grpcMockServer = grpc.NewServer()
		registerGrpc := newRegisterGrpc(ctx, logger)
		register.RegisterRegistrationServer(grpcMockServer, registerGrpc)
		grpcListener, errGrpc := net.Listen("tcp", "localhost:50051")
		Expect(errGrpc).ShouldNot(HaveOccurred())
		logger.Infof("register_test - gRPC client listening at %s", grpcListener.Addr().String())
		go func() {
			errGrpc := grpcMockServer.Serve(grpcListener)
			Expect(errGrpc).ShouldNot(HaveOccurred())
		}()

		// --------- start an HTTP server ---------
		//registerResponse := `[{"id": 123412341234123412341234, "code": 200}]`
		mux := http.NewServeMux()
		mux.HandleFunc("/keepalive", keepAliveHandler)
		mux.HandleFunc("/sensors/register/temperature", registerHandler)
		mux.HandleFunc("/sensors/register/humidity", registerHandler)
		mux.HandleFunc("/sensors/register/light", registerHandler)
		mux.HandleFunc("/sensors/register/motion", registerHandler)
		mux.HandleFunc("/sensors/register/airquality", registerHandler)
		mux.HandleFunc("/sensors/register/airpressure", registerHandler)
		httpListener, errHttp := net.Listen("tcp", "localhost:8000")
		logger.Infof("register_test - HTTP client listening at %s", httpListener.Addr().String())
		Expect(errHttp).ShouldNot(HaveOccurred())
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
		grpcMockServer.Stop()
		httpMockServer.Close()
		test_utils.DropAllCollections(ctx, collProfiles, collHomes, collDevices)
	})

	Describe("calling register api", func() {
		When("registering a new device", func() {

			It("should return a success", func() {
				By("with an existing profile with a valid apiToken")
				err := test_utils.InsertOne(ctx, collProfiles, profile)
				Expect(err).ShouldNot(HaveOccurred())

				feature := api.FeatureReq{
					Type:   "controller",
					Name:   "test",
					Enable: true,
					Order:  1,
					Unit:   "-",
				}
				deviceRegisterReq := api.DeviceRegisterReq{
					Mac:          "11:22:33:44:55:66",
					Manufacturer: "test",
					Model:        "test-model",
					ApiToken:     profile.ApiToken,
					Features:     []api.FeatureReq{feature},
				}
				var buf bytes.Buffer
				err = json.NewEncoder(&buf).Encode(deviceRegisterReq)
				Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodPost, "/api/register", &buf)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusOK))
				var device models.Device
				err = json.Unmarshal(recorder.Body.Bytes(), &device)
				Expect(err).ShouldNot(HaveOccurred())

				Expect([]byte(device.ID.Hex())).To(HaveLen(24))
				Expect([]byte(device.UUID)).To(HaveLen(36))
				Expect(device.Mac).To(Equal(deviceRegisterReq.Mac))
				Expect(device.Manufacturer).To(Equal(deviceRegisterReq.Manufacturer))
				Expect(device.Model).To(Equal(deviceRegisterReq.Model))
				Expect(device.Features).To(HaveLen(len(deviceRegisterReq.Features)))
				for _, featureRes := range device.Features {
					Expect([]byte(featureRes.UUID)).To(HaveLen(36))
					Expect(featureRes.Type).To(Equal(feature.Type))
					Expect(featureRes.Name).To(Equal(feature.Name))
					Expect(featureRes.Enable).To(Equal(feature.Enable))
					Expect(featureRes.Order).To(Equal(feature.Order))
					Expect(featureRes.Unit).To(Equal(feature.Unit))
				}
			})

			It("should return a 409 if the device is already registered", func() {
				By("with an existing profile with a valid apiToken")
				err := test_utils.InsertOne(ctx, collProfiles, profile)
				Expect(err).ShouldNot(HaveOccurred())

				feature := api.FeatureReq{
					Type:   "controller",
					Name:   "test",
					Enable: true,
					Order:  1,
					Unit:   "-",
				}
				deviceRegisterReq := api.DeviceRegisterReq{
					Mac:          "11:22:33:44:55:66",
					Manufacturer: "test",
					Model:        "test-model",
					ApiToken:     profile.ApiToken,
					Features:     []api.FeatureReq{feature},
				}
				var buf bytes.Buffer
				err = json.NewEncoder(&buf).Encode(deviceRegisterReq)
				Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodPost, "/api/register", &buf)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusOK))
				var device models.Device
				err = json.Unmarshal(recorder.Body.Bytes(), &device)
				Expect(err).ShouldNot(HaveOccurred())

				Expect([]byte(device.ID.Hex())).To(HaveLen(24))
				Expect([]byte(device.UUID)).To(HaveLen(36))
				Expect(device.Mac).To(Equal(deviceRegisterReq.Mac))
				Expect(device.Manufacturer).To(Equal(deviceRegisterReq.Manufacturer))
				Expect(device.Model).To(Equal(deviceRegisterReq.Model))
				Expect(device.Features).To(HaveLen(len(deviceRegisterReq.Features)))
				for _, featureRes := range device.Features {
					Expect([]byte(featureRes.UUID)).To(HaveLen(36))
					Expect(featureRes.Type).To(Equal(feature.Type))
					Expect(featureRes.Name).To(Equal(feature.Name))
					Expect(featureRes.Enable).To(Equal(feature.Enable))
					Expect(featureRes.Order).To(Equal(feature.Order))
					Expect(featureRes.Unit).To(Equal(feature.Unit))
				}

				var buf2 bytes.Buffer
				err = json.NewEncoder(&buf2).Encode(deviceRegisterReq)
				Expect(err).ShouldNot(HaveOccurred())
				recorder = httptest.NewRecorder()
				req = httptest.NewRequest(http.MethodPost, "/api/register", &buf2)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusConflict))
				Expect(recorder.Body.String()).To(Equal(`{"message":"Already registered"}`))
			})
		})

		When("registering a new sensor", func() {
			It("should return a success", func() {
				By("with an existing profile with a valid apiToken")
				err := test_utils.InsertOne(ctx, collProfiles, profile)
				Expect(err).ShouldNot(HaveOccurred())

				feature := api.FeatureReq{
					Type:   "sensor",
					Name:   "temperature",
					Enable: true,
					Order:  1,
					Unit:   "°C",
				}
				sensorRegisterReq := api.DeviceRegisterReq{
					Mac:          "11:22:33:44:55:66",
					Manufacturer: "test",
					Model:        "test-model",
					ApiToken:     profile.ApiToken,
					Features:     []api.FeatureReq{feature},
				}
				var buf bytes.Buffer
				err = json.NewEncoder(&buf).Encode(sensorRegisterReq)
				Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodPost, "/api/register", &buf)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusOK))
				var device models.Device
				err = json.Unmarshal(recorder.Body.Bytes(), &device)
				Expect(err).ShouldNot(HaveOccurred())

				Expect([]byte(device.ID.Hex())).To(HaveLen(24))
				Expect([]byte(device.UUID)).To(HaveLen(36))
				Expect(device.Mac).To(Equal(sensorRegisterReq.Mac))
				Expect(device.Manufacturer).To(Equal(sensorRegisterReq.Manufacturer))
				Expect(device.Model).To(Equal(sensorRegisterReq.Model))
				Expect(device.Features).To(HaveLen(len(sensorRegisterReq.Features)))
				for _, featureRes := range device.Features {
					Expect([]byte(featureRes.UUID)).To(HaveLen(36))
					Expect(featureRes.Type).To(Equal(feature.Type))
					Expect(featureRes.Name).To(Equal(feature.Name))
					Expect(featureRes.Enable).To(Equal(feature.Enable))
					Expect(featureRes.Order).To(Equal(feature.Order))
					Expect(featureRes.Unit).To(Equal(feature.Unit))
				}
			})

			It("should return a 409 if the sensor is already registered", func() {
				By("with an existing profile with a valid apiToken")
				err := test_utils.InsertOne(ctx, collProfiles, profile)
				Expect(err).ShouldNot(HaveOccurred())

				feature := api.FeatureReq{
					Type:   "sensor",
					Name:   "temperature",
					Enable: true,
					Order:  1,
					Unit:   "°C",
				}
				sensorRegisterReq := api.DeviceRegisterReq{
					Mac:          "11:22:33:44:55:66",
					Manufacturer: "test",
					Model:        "test-model",
					ApiToken:     profile.ApiToken,
					Features:     []api.FeatureReq{feature},
				}
				var buf bytes.Buffer
				err = json.NewEncoder(&buf).Encode(sensorRegisterReq)
				Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodPost, "/api/register", &buf)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusOK))
				var device models.Device
				err = json.Unmarshal(recorder.Body.Bytes(), &device)
				Expect(err).ShouldNot(HaveOccurred())

				Expect([]byte(device.ID.Hex())).To(HaveLen(24))
				Expect([]byte(device.UUID)).To(HaveLen(36))
				Expect(device.Mac).To(Equal(sensorRegisterReq.Mac))
				Expect(device.Manufacturer).To(Equal(sensorRegisterReq.Manufacturer))
				Expect(device.Model).To(Equal(sensorRegisterReq.Model))
				Expect(device.Features).To(HaveLen(len(sensorRegisterReq.Features)))
				for _, featureRes := range device.Features {
					Expect([]byte(featureRes.UUID)).To(HaveLen(36))
					Expect(featureRes.Type).To(Equal(feature.Type))
					Expect(featureRes.Name).To(Equal(feature.Name))
					Expect(featureRes.Enable).To(Equal(feature.Enable))
					Expect(featureRes.Order).To(Equal(feature.Order))
					Expect(featureRes.Unit).To(Equal(feature.Unit))
				}

				var buf2 bytes.Buffer
				err = json.NewEncoder(&buf2).Encode(sensorRegisterReq)
				Expect(err).ShouldNot(HaveOccurred())
				recorder = httptest.NewRecorder()
				req = httptest.NewRequest(http.MethodPost, "/api/register", &buf2)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusConflict))
				Expect(recorder.Body.String()).To(Equal(`{"message":"Already registered"}`))
			})
		})

		When("you pass bad inputs", func() {
			It("should return an error, if body is missing", func() {
				err := test_utils.InsertOne(ctx, collProfiles, profile)
				Expect(err).ShouldNot(HaveOccurred())
				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodPost, "/api/register", nil)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
				Expect(recorder.Body.String()).To(Equal(`{"error":"invalid request payload"}`))
			})

			It("should return an error, if body is not valid", func() {
				err := test_utils.InsertOne(ctx, collProfiles, profile)
				Expect(err).ShouldNot(HaveOccurred())

				feature := api.FeatureReq{
					Type:   "unknowntype", // not valid, because must be either 'controller' or 'sensor'
					Name:   "",            // not valid, because 2 <= length <= 20
					Enable: true,
					Order:  0,  // not valid, because must be >= 1
					Unit:   "", // not valid, because 1 <= length <= 10
				}
				deviceRegisterReq := api.DeviceRegisterReq{
					Mac:          "1234", // not valid, because must be a MAC
					Manufacturer: "",     // not valid, because 3 <= length <= 50
					Model:        "",     // not valid, because 3 <= length <= 20
					ApiToken:     "1234", // not valid, because must be an UUIDv4
					Features:     []api.FeatureReq{feature},
				}
				var buf bytes.Buffer
				err = json.NewEncoder(&buf).Encode(deviceRegisterReq)
				Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodPost, "/api/register", &buf)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
				Expect(recorder.Body.String()).To(Equal(`{"error":"invalid request body, these fields are not valid: mac manufacturer model apitoken type name order unit"}`))
			})

			It("should return an error, if apiToken doesn't exist", func() {
				err := test_utils.InsertOne(ctx, collProfiles, profile)
				Expect(err).ShouldNot(HaveOccurred())

				unknownApiToken := uuid.NewString()
				feature := api.FeatureReq{
					Type:   "controller",
					Name:   "test",
					Enable: true,
					Order:  1,
					Unit:   "-",
				}
				deviceRegisterReq := api.DeviceRegisterReq{
					Mac:          "11:22:33:44:55:66",
					Manufacturer: "test",
					Model:        "test-model",
					ApiToken:     unknownApiToken,
					Features:     []api.FeatureReq{feature},
				}
				var buf bytes.Buffer
				err = json.NewEncoder(&buf).Encode(deviceRegisterReq)
				Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodPost, "/api/register", &buf)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
				Expect(recorder.Body.String()).To(Equal(`{"error":"cnnot register, profile token missing or not valid"}`))
			})
		})
	})
})
