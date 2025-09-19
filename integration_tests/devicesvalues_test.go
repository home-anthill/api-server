package integration_tests

import (
	"api-server/api/grpc/device"
	"api-server/db"
	"api-server/initialization"
	"api-server/models"
	"api-server/testuutils"
	"bytes"
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
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var currentDate = time.Now()

var deviceUUID = uuid.NewString()

// mocked GET controller API values
var controllerAcFeatureUUID = uuid.NewString()
var controllerValue = float32(22.34)

// mocked GET sensor API values
var temperatureFeatureUUID = uuid.NewString()
var lightFeatureUUID = uuid.NewString()
var humidityFeatureUUID = uuid.NewString()
var airpressureFeatureUUID = uuid.NewString()
var motionFeatureUUID = uuid.NewString()
var airqualityFeatureUUID = uuid.NewString()
var temperatureSensorValue = float32(22.12)
var lightSensorValue = float32(17.1)
var humiditySensorValue = float32(55.12)
var airpressureSensorValue = float32(100.71)
var motionSensorValue = float32(1.0)
var airqualitySensorValue = float32(2.0)

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

// Attention: this function stub name must match the one defined in .proto file
func (handler *deviceGrpcStub) GetValue(ctx context.Context, in *device.GetValueRequest) (*device.GetValueResponse, error) {
	fmt.Printf("gRPC stub - GetValue - received = %#v\n", in)
	return &device.GetValueResponse{
		// we create this partial mock response for all controller values for simplicity
		// so we don't distinguish between different feature names
		Value:      controllerValue,
		CreatedAt:  currentDate.UnixMilli(),
		ModifiedAt: currentDate.UnixMilli(),
	}, nil
}

// Attention: this function stub name must match the one defined in .proto file
func (handler *deviceGrpcStub) SetValue(ctx context.Context, in *device.SetValueRequest) (*device.SetValueResponse, error) {
	fmt.Printf("gRPC stub - SetValue - received = %#v\n", in)
	return &device.SetValueResponse{
		Status:  "200",
		Message: "Updated",
	}, nil
}

var _ = Describe("DevicesValues", func() {
	var ctx context.Context
	var logger *zap.SugaredLogger
	var router *gin.Engine
	var client *mongo.Client
	var collProfiles *mongo.Collection
	var collHomes *mongo.Collection
	var collDevices *mongo.Collection
	var grpcMockServer *grpc.Server
	var httpMockServer *httptest.Server

	var deviceController = models.Device{
		ID:           primitive.NewObjectID(),
		Mac:          "11:22:33:44:55:66",
		Manufacturer: "test",
		Model:        "test",
		UUID:         deviceUUID,
		Features: []models.Feature{{
			UUID:   controllerAcFeatureUUID,
			Type:   "controller",
			Name:   "ac-beko",
			Enable: true,
			Order:  1,
			Unit:   "-",
		}},
		CreatedAt:  currentDate,
		ModifiedAt: currentDate,
	}
	// hybrid device like a thermostat
	// (temperature sensor + airconditioner)
	var deviceHybrid = models.Device{
		ID:           primitive.NewObjectID(),
		Mac:          "CC:22:33:44:55:66",
		Manufacturer: "test",
		Model:        "test",
		UUID:         deviceUUID,
		Features: []models.Feature{{
			UUID:   controllerAcFeatureUUID,
			Type:   "controller",
			Name:   "ac-lg",
			Enable: true,
			Order:  1,
			Unit:   "-",
		}, {
			UUID:   temperatureFeatureUUID,
			Type:   "sensor",
			Name:   "temperature",
			Enable: true,
			Order:  2,
			Unit:   "°C",
		}},
		CreatedAt:  currentDate,
		ModifiedAt: currentDate,
	}
	var deviceSensor = models.Device{
		ID:           primitive.NewObjectID(),
		Mac:          "AA:22:33:44:55:BB",
		Manufacturer: "test2",
		Model:        "test2",
		UUID:         deviceUUID,
		Features: []models.Feature{{
			UUID:   temperatureFeatureUUID,
			Type:   "sensor",
			Name:   "temperature",
			Enable: true,
			Order:  1,
			Unit:   "°C",
		}, {
			UUID:   lightFeatureUUID,
			Type:   "sensor",
			Name:   "light",
			Enable: true,
			Order:  2,
			Unit:   "lux",
		}, {
			UUID:   humidityFeatureUUID,
			Type:   "sensor",
			Name:   "humidity",
			Enable: true,
			Order:  3,
			Unit:   "%",
		}, {
			UUID:   airpressureFeatureUUID,
			Type:   "sensor",
			Name:   "airpressure",
			Enable: true,
			Order:  4,
			Unit:   "lux",
		}, {
			UUID:   motionFeatureUUID,
			Type:   "sensor",
			Name:   "motion",
			Enable: true,
			Order:  5,
			Unit:   "-",
		}, {
			UUID:   airqualityFeatureUUID,
			Type:   "sensor",
			Name:   "airquality",
			Enable: true,
			Order:  6,
			Unit:   "-",
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
		_, _ = w.Write([]byte(getSensorJSONResponse(temperatureSensorValue, currentDate, currentDate)))
	})
	getSensorLightHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(getSensorJSONResponse(lightSensorValue, currentDate, currentDate)))
	})
	getSensorHumidityHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(getSensorJSONResponse(humiditySensorValue, currentDate, currentDate)))
	})
	getSensorAirpressureHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(getSensorJSONResponse(airpressureSensorValue, currentDate, currentDate)))
	})
	getSensorMotionHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(getSensorJSONResponse(motionSensorValue, currentDate, currentDate)))
	})
	getSensorAirqualityHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(getSensorJSONResponse(airqualitySensorValue, currentDate, currentDate)))
	})

	BeforeEach(func() {
		logger, router, ctx, client = initialization.Start()
		defer logger.Sync()

		collProfiles = db.GetCollections(client).Profiles
		collHomes = db.GetCollections(client).Homes
		collDevices = db.GetCollections(client).Devices

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
		mux.HandleFunc("/sensors/"+deviceUUID+"/temperature", getSensorTemperatureHandler)
		mux.HandleFunc("/sensors/"+deviceUUID+"/light", getSensorLightHandler)
		mux.HandleFunc("/sensors/"+deviceUUID+"/humidity", getSensorHumidityHandler)
		mux.HandleFunc("/sensors/"+deviceUUID+"/airpressure", getSensorAirpressureHandler)
		mux.HandleFunc("/sensors/"+deviceUUID+"/motion", getSensorMotionHandler)
		mux.HandleFunc("/sensors/"+deviceUUID+"/airquality", getSensorAirqualityHandler)
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
		grpcMockServer.Stop()
		httpMockServer.Close()
		testuutils.DropAllCollections(ctx, collProfiles, collHomes, collDevices)
	})

	Context("calling devicesvalues api GET", func() {
		BeforeEach(func() {
			err := testuutils.InsertOne(ctx, collDevices, deviceController)
			Expect(err).ShouldNot(HaveOccurred())
			err = testuutils.InsertOne(ctx, collDevices, deviceHybrid)
			Expect(err).ShouldNot(HaveOccurred())
			err = testuutils.InsertOne(ctx, collDevices, deviceSensor)
			Expect(err).ShouldNot(HaveOccurred())
		})

		When("profile owns a controller device", func() {
			It("should get device value of a specific feature", func() {
				jwtToken, cookieSession := testuutils.GetJwt(router)
				profileRes := testuutils.GetLoggedProfile(router, jwtToken, cookieSession)

				err := testuutils.AssignDeviceToProfile(ctx, collProfiles, profileRes.ID, deviceController.ID)
				Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, "/api/devices/"+deviceController.ID.Hex()+"/values", nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusOK))
				var deviceStates []models.DeviceFeatureState
				err = json.Unmarshal(recorder.Body.Bytes(), &deviceStates)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(deviceStates).To(HaveLen(1))
				Expect(deviceStates[0].Value).To(Equal(controllerValue))
				Expect(deviceStates[0].CreatedAt).To(Equal(currentDate.UnixMilli()))
				Expect(deviceStates[0].ModifiedAt).To(Equal(currentDate.UnixMilli()))
			})
		})
		//
		When("profile owns a sensor", func() {
			It("should get a list of values", func() {
				jwtToken, cookieSession := testuutils.GetJwt(router)
				profileRes := testuutils.GetLoggedProfile(router, jwtToken, cookieSession)

				err := testuutils.AssignDeviceToProfile(ctx, collProfiles, profileRes.ID, deviceSensor.ID)
				Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, "/api/devices/"+deviceSensor.ID.Hex()+"/values", nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusOK))
				var deviceStates []models.DeviceFeatureState
				err = json.Unmarshal(recorder.Body.Bytes(), &deviceStates)
				Expect(err).ShouldNot(HaveOccurred())
				// order is the same of deviceSensor.Features
				Expect(deviceStates).To(HaveLen(6))
				Expect(deviceStates[0].FeatureUUID).To(Equal(temperatureFeatureUUID))
				Expect(deviceStates[0].Value).To(Equal(temperatureSensorValue))
				Expect(deviceStates[0].CreatedAt).To(Equal(currentDate.UnixMilli()))
				Expect(deviceStates[0].ModifiedAt).To(Equal(currentDate.UnixMilli()))
				Expect(deviceStates[1].FeatureUUID).To(Equal(lightFeatureUUID))
				Expect(deviceStates[1].Value).To(Equal(lightSensorValue))
				Expect(deviceStates[1].CreatedAt).To(Equal(currentDate.UnixMilli()))
				Expect(deviceStates[1].ModifiedAt).To(Equal(currentDate.UnixMilli()))
				Expect(deviceStates[2].FeatureUUID).To(Equal(humidityFeatureUUID))
				Expect(deviceStates[2].Value).To(Equal(humiditySensorValue))
				Expect(deviceStates[2].CreatedAt).To(Equal(currentDate.UnixMilli()))
				Expect(deviceStates[2].ModifiedAt).To(Equal(currentDate.UnixMilli()))
				Expect(deviceStates[3].FeatureUUID).To(Equal(airpressureFeatureUUID))
				Expect(deviceStates[3].Value).To(Equal(airpressureSensorValue))
				Expect(deviceStates[3].CreatedAt).To(Equal(currentDate.UnixMilli()))
				Expect(deviceStates[3].ModifiedAt).To(Equal(currentDate.UnixMilli()))
				Expect(deviceStates[4].FeatureUUID).To(Equal(motionFeatureUUID))
				Expect(deviceStates[4].Value).To(Equal(motionSensorValue))
				Expect(deviceStates[4].CreatedAt).To(Equal(currentDate.UnixMilli()))
				Expect(deviceStates[4].ModifiedAt).To(Equal(currentDate.UnixMilli()))
				Expect(deviceStates[5].FeatureUUID).To(Equal(airqualityFeatureUUID))
				Expect(deviceStates[5].Value).To(Equal(airqualitySensorValue))
				Expect(deviceStates[5].CreatedAt).To(Equal(currentDate.UnixMilli()))
				Expect(deviceStates[5].ModifiedAt).To(Equal(currentDate.UnixMilli()))
			})
		})

		When("profile owns a hybrid device (controller + sensor)", func() {
			It("should get a list of values", func() {
				jwtToken, cookieSession := testuutils.GetJwt(router)
				profileRes := testuutils.GetLoggedProfile(router, jwtToken, cookieSession)

				err := testuutils.AssignDeviceToProfile(ctx, collProfiles, profileRes.ID, deviceHybrid.ID)
				Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, "/api/devices/"+deviceHybrid.ID.Hex()+"/values", nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusOK))
				var deviceStates []models.DeviceFeatureState
				err = json.Unmarshal(recorder.Body.Bytes(), &deviceStates)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(deviceStates).To(HaveLen(2))
				// to simplify testing setup we use a single mocked response for all controller values
				// for both setpoint and tolerance will return the same value in thi test
				Expect(deviceStates[0].FeatureUUID).To(Equal(controllerAcFeatureUUID))
				Expect(deviceStates[0].Value).To(Equal(controllerValue))
				Expect(deviceStates[0].CreatedAt).To(Equal(currentDate.UnixMilli()))
				Expect(deviceStates[0].ModifiedAt).To(Equal(currentDate.UnixMilli()))
				Expect(deviceStates[1].FeatureUUID).To(Equal(temperatureFeatureUUID))
				Expect(deviceStates[1].Value).To(Equal(temperatureSensorValue))
				Expect(deviceStates[1].CreatedAt).To(Equal(currentDate.UnixMilli()))
				Expect(deviceStates[1].ModifiedAt).To(Equal(currentDate.UnixMilli()))

			})
		})

		When("you pass bad inputs", func() {
			It("should return an error, because ...", func() {
				jwtToken, cookieSession := testuutils.GetJwt(router)

				badDeviceID := "bad_device_id"

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, "/api/devices/"+badDeviceID+"/values", nil)
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
				jwtToken, cookieSession := testuutils.GetJwt(router)
				profileRes := testuutils.GetLoggedProfile(router, jwtToken, cookieSession)

				unexistingDeviceID := primitive.NewObjectID()

				err := testuutils.AssignDeviceToProfile(ctx, collProfiles, profileRes.ID, unexistingDeviceID)
				Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, "/api/devices/"+unexistingDeviceID.Hex()+"/values", nil)
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
			err := testuutils.InsertOne(ctx, collDevices, deviceController)
			Expect(err).ShouldNot(HaveOccurred())
		})

		When("profile owns a controller device", func() {
			It("should set a new value of a feature of that device", func() {
				jwtToken, cookieSession := testuutils.GetJwt(router)
				profileRes := testuutils.GetLoggedProfile(router, jwtToken, cookieSession)

				err := testuutils.AssignDeviceToProfile(ctx, collProfiles, profileRes.ID, deviceController.ID)
				Expect(err).ShouldNot(HaveOccurred())

				devState := models.DeviceFeatureState{
					FeatureUUID: deviceController.Features[0].UUID,
					Type:        models.Controller,
					Name:        deviceController.Features[0].Name,
					Value:       float32(22.45),
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
				jwtToken, cookieSession := testuutils.GetJwt(router)

				badDeviceID := "bad_device_id"

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodPost, "/api/devices/"+badDeviceID+"/values", nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
				Expect(recorder.Body.String()).To(Equal(`{"error":"wrong format of the path param 'id'"}`))
			})

			It("should return an error, because body is missing", func() {
				jwtToken, cookieSession := testuutils.GetJwt(router)
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
				jwtToken, cookieSession := testuutils.GetJwt(router)
				profileRes := testuutils.GetLoggedProfile(router, jwtToken, cookieSession)

				err := testuutils.AssignDeviceToProfile(ctx, collProfiles, profileRes.ID, deviceController.ID)
				Expect(err).ShouldNot(HaveOccurred())

				devState := models.DeviceFeatureState{
					FeatureUUID: deviceController.Features[0].UUID,
					// type is a required field, but here it's missing
					// name is a required field, but here it's missing
					Value: float32(22.45),
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
				Expect(recorder.Body.String()).To(Equal(`{"error":"invalid request body, these fields are not valid: type name"}`))
			})
		})

		When("profile don't own any device", func() {
			It("should return an error, because device is not owned by profile", func() {
				jwtToken, cookieSession := testuutils.GetJwt(router)

				devState := models.DeviceFeatureState{
					FeatureUUID: deviceController.Features[0].UUID,
					Type:        models.Controller,
					Name:        deviceController.Features[0].Name,
					Value:       float32(22.45),
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
				jwtToken, cookieSession := testuutils.GetJwt(router)
				profileRes := testuutils.GetLoggedProfile(router, jwtToken, cookieSession)

				unexistingDeviceID := primitive.NewObjectID()
				err := testuutils.AssignDeviceToProfile(ctx, collProfiles, profileRes.ID, unexistingDeviceID)
				Expect(err).ShouldNot(HaveOccurred())

				devState := models.DeviceFeatureState{
					FeatureUUID: deviceController.Features[0].UUID,
					Type:        models.Controller,
					Name:        deviceController.Features[0].Name,
					Value:       float32(22.45),
				}
				var deviceState bytes.Buffer
				err = json.NewEncoder(&deviceState).Encode(devState)
				Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodPost, "/api/devices/"+unexistingDeviceID.Hex()+"/values", &deviceState)
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

func getSensorJSONResponse(value float32, currDate time.Time, modDate time.Time) string {
	return `{"value": ` + fmt.Sprintf("%f", value) + `, "createdAt": ` + fmt.Sprintf("%v", currDate.UnixMilli()) + `, "modifiedAt": ` + fmt.Sprintf("%v", modDate.UnixMilli()) + `}`
}
