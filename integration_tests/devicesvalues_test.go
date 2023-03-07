package integration_tests

import (
	"api-server/api/grpc/device"
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
	"time"
)

var sensorUuid = uuid.NewString()
var temperatureFeatureUuid = uuid.NewString()
var lightFeatureUuid = uuid.NewString()
var temperatureSensorValue float64 = 22
var lightSensorValue float64 = 17

type deviceGrpcStub struct {
	device.UnimplementedDeviceServer
	ctx    context.Context
	logger *zap.SugaredLogger
}

func newDeviceGrpc(ctx context.Context, logger *zap.SugaredLogger) *deviceGrpcStub {
	return &deviceGrpcStub{
		ctx:    ctx,
		logger: logger,
	}
}

func (handler *deviceGrpcStub) GetStatus(ctx context.Context, in *device.StatusRequest) (*device.StatusResponse, error) {
	fmt.Printf("devicesvalues_test - GetStatus - received = %#v\n", in)
	return &device.StatusResponse{
		On:          true,
		Temperature: 22,
		Mode:        1,
		FanSpeed:    2,
	}, nil
}

func (handler *deviceGrpcStub) SetValues(ctx context.Context, in *device.ValuesRequest) (*device.ValuesResponse, error) {
	fmt.Printf("devicesvalues_test - SetValues - received = %#v\n", in)
	return &device.ValuesResponse{
		Status:  "200",
		Message: "Updated",
	}, nil
}

var _ = Describe("Devices", func() {
	var ctx context.Context
	var logger *zap.SugaredLogger
	var router *gin.Engine
	var collProfiles *mongo.Collection
	var collHomes *mongo.Collection
	var collDevices *mongo.Collection
	var grpcMockServer *grpc.Server
	var httpMockServer *httptest.Server

	var currentDate = time.Now()
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
		CreatedAt:  currentDate,
		ModifiedAt: currentDate,
	}
	var deviceSensor = models.Device{
		ID:           primitive.NewObjectID(),
		Mac:          "AA:22:33:44:55:BB",
		Manufacturer: "test2",
		Model:        "test2",
		UUID:         sensorUuid,
		Features: []models.Feature{{
			UUID:   temperatureFeatureUuid,
			Type:   "sensor",
			Name:   "temperature",
			Enable: true,
			Order:  1,
			Unit:   "°C",
		}, {
			UUID:   lightFeatureUuid,
			Type:   "sensor",
			Name:   "light",
			Enable: true,
			Order:  1,
			Unit:   "lux",
		}},
		CreatedAt:  currentDate,
		ModifiedAt: currentDate,
	}

	keepAliveHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"alive": true}]`))
	})
	getSensorTemperatureHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		value := fmt.Sprintf("%f", temperatureSensorValue)
		_, _ = w.Write([]byte(`{"value": ` + value + `}`))
	})
	getSensorLightHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		value := fmt.Sprintf("%f", lightSensorValue)
		_, _ = w.Write([]byte(`{"value": ` + value + `}`))
	})

	BeforeEach(func() {
		logger, router, ctx, collProfiles, collHomes, collDevices = initialization.Start()
		defer logger.Sync()

		err := os.Setenv("SINGLE_USER_LOGIN_EMAIL", "test@test.com")
		Expect(err).ShouldNot(HaveOccurred())

		// --------- start a gRPC server ---------
		grpcMockServer = grpc.NewServer()
		deviceGrpc := newDeviceGrpc(ctx, logger)
		device.RegisterDeviceServer(grpcMockServer, deviceGrpc)
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
		mux.HandleFunc("/sensors/"+sensorUuid+"/temperature", getSensorTemperatureHandler)
		mux.HandleFunc("/sensors/"+sensorUuid+"/light", getSensorLightHandler)
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

	Context("calling devicesvalues api GET", func() {
		BeforeEach(func() {
			err := test_utils.InsertOne(ctx, collDevices, deviceController)
			Expect(err).ShouldNot(HaveOccurred())
			err = test_utils.InsertOne(ctx, collDevices, deviceSensor)
			Expect(err).ShouldNot(HaveOccurred())
		})

		When("profile owns a controller device", func() {
			It("should get a list of devices", func() {
				jwtToken, cookieSession := test_utils.GetJwt(router)
				profileRes := test_utils.GetLoggedProfile(router, jwtToken, cookieSession)

				err := test_utils.AssignDeviceToProfile(ctx, collProfiles, profileRes.ID, deviceController.ID)
				Expect(err).ShouldNot(HaveOccurred())
				err = test_utils.AssignDeviceToProfile(ctx, collProfiles, profileRes.ID, deviceSensor.ID)
				Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, "/api/devices/"+deviceController.ID.Hex()+"/values", nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusOK))
				var deviceState models.DeviceState
				err = json.Unmarshal(recorder.Body.Bytes(), &deviceState)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(deviceState.On).To(Equal(true))
				Expect(deviceState.Temperature).To(Equal(22))
				Expect(deviceState.Mode).To(Equal(1))
				Expect(deviceState.FanSpeed).To(Equal(2))
			})
		})

		When("profile owns a sensor", func() {
			It("should get a list of devices", func() {
				jwtToken, cookieSession := test_utils.GetJwt(router)
				profileRes := test_utils.GetLoggedProfile(router, jwtToken, cookieSession)

				err := test_utils.AssignDeviceToProfile(ctx, collProfiles, profileRes.ID, deviceController.ID)
				Expect(err).ShouldNot(HaveOccurred())
				err = test_utils.AssignDeviceToProfile(ctx, collProfiles, profileRes.ID, deviceSensor.ID)
				Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, "/api/devices/"+deviceSensor.ID.Hex()+"/values", nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusOK))
				var deviceStates []models.SensorValue
				err = json.Unmarshal(recorder.Body.Bytes(), &deviceStates)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(deviceStates[0].UUID).To(Equal(temperatureFeatureUuid))
				Expect(deviceStates[0].Value).To(Equal(temperatureSensorValue))
				Expect(deviceStates[1].UUID).To(Equal(lightFeatureUuid))
				Expect(deviceStates[1].Value).To(Equal(lightSensorValue))
			})
		})

		When("you pass bad inputs", func() {
			It("should return an error, because ...", func() {
				jwtToken, cookieSession := test_utils.GetJwt(router)

				badDeviceId := "bad_device_id"

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, "/api/devices/"+badDeviceId+"/values", nil)
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
				jwtToken, cookieSession := test_utils.GetJwt(router)
				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, "/api/devices/"+deviceController.ID.Hex()+"/values", nil)
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
				jwtToken, cookieSession := test_utils.GetJwt(router)
				profileRes := test_utils.GetLoggedProfile(router, jwtToken, cookieSession)

				unexistingDeviceId := primitive.NewObjectID()

				err := test_utils.AssignDeviceToProfile(ctx, collProfiles, profileRes.ID, unexistingDeviceId)
				Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, "/api/devices/"+unexistingDeviceId.Hex()+"/values", nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
				Expect(recorder.Body.String()).To(Equal(`{"error":"cannot find device"}`))
			})
		})
	})

	Context("calling devicesvalues api POST", func() {
		BeforeEach(func() {
			err := test_utils.InsertOne(ctx, collDevices, deviceController)
			Expect(err).ShouldNot(HaveOccurred())
		})

		When("profile owns a controller device", func() {
			It("should set a new value of that device", func() {
				jwtToken, cookieSession := test_utils.GetJwt(router)
				profileRes := test_utils.GetLoggedProfile(router, jwtToken, cookieSession)

				err := test_utils.AssignDeviceToProfile(ctx, collProfiles, profileRes.ID, deviceController.ID)
				Expect(err).ShouldNot(HaveOccurred())

				devState := models.DeviceState{
					On:          true,
					Temperature: 28,
					Mode:        2,
					FanSpeed:    1,
				}
				var deviceState bytes.Buffer
				err = json.NewEncoder(&deviceState).Encode(devState)
				Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodPost, "/api/devices/"+deviceController.ID.Hex()+"/values", &deviceState)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusOK))
				Expect(recorder.Body.String()).To(Equal(`{"message":"set value success"}`))
			})
		})

		When("you pass bad inputs", func() {
			It("should return an error, because ...", func() {
				jwtToken, cookieSession := test_utils.GetJwt(router)

				badDeviceId := "bad_device_id"

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodPost, "/api/devices/"+badDeviceId+"/values", nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
				Expect(recorder.Body.String()).To(Equal(`{"error":"wrong format of the path param 'id'"}`))
			})

			It("should return an error, because body is missing", func() {
				jwtToken, cookieSession := test_utils.GetJwt(router)
				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodPost, "/api/devices/"+deviceController.ID.Hex()+"/values", nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
				Expect(recorder.Body.String()).To(Equal(`{"error":"invalid request payload"}`))
			})

			It("should return an error, because body is not valid", func() {
				jwtToken, cookieSession := test_utils.GetJwt(router)
				profileRes := test_utils.GetLoggedProfile(router, jwtToken, cookieSession)

				err := test_utils.AssignDeviceToProfile(ctx, collProfiles, profileRes.ID, deviceController.ID)
				Expect(err).ShouldNot(HaveOccurred())

				devState := models.DeviceState{
					On:          true,
					Temperature: 10, // invalid, because it must be >= 17 and <= 30
					Mode:        0,  // invalid, because it must be >= 1 and <= 5
					FanSpeed:    0,  // invalid, because it must be >= 1 and <= 5
				}
				var deviceState bytes.Buffer
				err = json.NewEncoder(&deviceState).Encode(devState)
				Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodPost, "/api/devices/"+deviceController.ID.Hex()+"/values", &deviceState)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
				Expect(recorder.Body.String()).To(Equal(`{"error":"invalid request body, these fields are not valid: temperature mode fanspeed"}`))
			})
		})

		When("profile don't own any device", func() {
			It("should return an error, because device is not owned by profile", func() {
				jwtToken, cookieSession := test_utils.GetJwt(router)

				devState := models.DeviceState{
					On:          true,
					Temperature: 27,
					Mode:        2,
					FanSpeed:    2,
				}
				var deviceState bytes.Buffer
				err := json.NewEncoder(&deviceState).Encode(devState)
				Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodPost, "/api/devices/"+deviceController.ID.Hex()+"/values", &deviceState)
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
				jwtToken, cookieSession := test_utils.GetJwt(router)
				profileRes := test_utils.GetLoggedProfile(router, jwtToken, cookieSession)

				unexistingDeviceId := primitive.NewObjectID()
				err := test_utils.AssignDeviceToProfile(ctx, collProfiles, profileRes.ID, unexistingDeviceId)
				Expect(err).ShouldNot(HaveOccurred())

				devState := models.DeviceState{
					On:          true,
					Temperature: 27,
					Mode:        2,
					FanSpeed:    2,
				}
				var deviceState bytes.Buffer
				err = json.NewEncoder(&deviceState).Encode(devState)
				Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodPost, "/api/devices/"+unexistingDeviceId.Hex()+"/values", &deviceState)
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
