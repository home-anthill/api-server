package integration_tests

import (
	devicepb "api-server/api/grpc/device"
	"api-server/db"
	"api-server/initialization"
	"api-server/models"
	"api-server/testuutils"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
)

type deleteDeviceGrpcStub struct {
	devicepb.UnimplementedDeviceServer
	controllerDeleteCalled     *atomic.Bool
	controllerDeleteShouldFail *atomic.Bool
}

func (handler *deleteDeviceGrpcStub) DeleteValue(ctx context.Context, in *devicepb.DeleteValueRequest) (*devicepb.SetValueResponse, error) {
	handler.controllerDeleteCalled.Store(true)
	if handler.controllerDeleteShouldFail.Load() {
		return nil, grpcstatus.Error(codes.Unavailable, "controller cleanup unavailable")
	}
	return &devicepb.SetValueResponse{Status: "200", Message: "Deleted"}, nil
}

var _ = Describe("Devices", func() {
	var ctx context.Context
	var logger *zap.SugaredLogger
	var router *gin.Engine
	var client *mongo.Client
	var collProfiles *mongo.Collection
	var collHomes *mongo.Collection
	var collDevices *mongo.Collection
	var httpOnlineMockServer *httptest.Server
	var httpSensorMockServer *httptest.Server
	var grpcMockServer *grpc.Server
	var notificationPreferenceBody string
	var onlineDeleteCalled atomic.Bool
	var sensorDeleteCalled atomic.Bool
	var controllerDeleteCalled atomic.Bool
	var onlineDeleteShouldFail atomic.Bool
	var sensorDeleteShouldFail atomic.Bool
	var controllerDeleteShouldFail atomic.Bool

	var currDate = time.Now()
	var deviceController = models.Device{
		ID:           bson.NewObjectID(),
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
		ID:           bson.NewObjectID(),
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
	var deviceOnlineSensor = models.Device{
		ID:           bson.NewObjectID(),
		Mac:          "AA:22:33:44:55:FF",
		Manufacturer: "test3",
		Model:        "online",
		UUID:         uuid.NewString(),
		Features: []models.Feature{{
			UUID:   uuid.NewString(),
			Type:   "sensor",
			Name:   "online",
			Enable: true,
			Order:  1,
			Unit:   "-",
		}},
		CreatedAt:  currDate,
		ModifiedAt: currDate,
	}
	var home = models.Home{
		ID:       bson.NewObjectID(),
		Name:     "home1",
		Location: "location1",
		Rooms: []models.Room{{
			ID:         bson.NewObjectID(),
			Name:       "room1",
			Floor:      1,
			CreatedAt:  currDate,
			ModifiedAt: currDate,
			Devices:    []bson.ObjectID{},
		}},
		CreatedAt:  currDate,
		ModifiedAt: currDate,
	}

	deleteOnlineSensorOnlineHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		onlineDeleteCalled.Store(true)
		if onlineDeleteShouldFail.Load() {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error":"online cleanup unavailable"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	})
	deleteOnlineSensorSensorHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		sensorDeleteCalled.Store(true)
		if sensorDeleteShouldFail.Load() {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error":"sensor cleanup unavailable"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	})
	updateOnlineFeatureNotificationHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(r.Body)
		notificationPreferenceBody = buf.String()
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	})

	BeforeEach(func() {
		grpcListener, errGrpc := net.Listen("tcp", "127.0.0.1:0")
		Expect(errGrpc).ShouldNot(HaveOccurred())
		err := os.Setenv("GRPC_URL", grpcListener.Addr().String())
		Expect(err).ShouldNot(HaveOccurred())

		logger, router, client = initialization.MustStart()
		ctx = context.Background()
		defer logger.Sync()

		collProfiles = db.GetCollections(client).Profiles
		collHomes = db.GetCollections(client).Homes
		collDevices = db.GetCollections(client).Devices

		err = os.Setenv("LIMIT_TO_USER_EMAILS", "test@test.com")
		Expect(err).ShouldNot(HaveOccurred())
		onlineDeleteCalled.Store(false)
		sensorDeleteCalled.Store(false)
		controllerDeleteCalled.Store(false)
		onlineDeleteShouldFail.Store(false)
		sensorDeleteShouldFail.Store(false)
		controllerDeleteShouldFail.Store(false)

		grpcMockServer = grpc.NewServer()
		devicepb.RegisterDeviceServer(grpcMockServer, &deleteDeviceGrpcStub{
			controllerDeleteCalled:     &controllerDeleteCalled,
			controllerDeleteShouldFail: &controllerDeleteShouldFail,
		})
		go func() {
			errGrpc := grpcMockServer.Serve(grpcListener)
			if errGrpc != nil && !errors.Is(errGrpc, grpc.ErrServerStopped) {
				panic(errGrpc)
			}
		}()

		// --------- start an online HTTP server ---------
		onlineMux := http.NewServeMux()
		onlineMux.HandleFunc(
			"/online/"+deviceOnlineSensor.UUID+"/features/"+deviceOnlineSensor.Features[0].UUID,
			deleteOnlineSensorOnlineHandler,
		)
		onlineMux.HandleFunc(
			"/alarms/"+deviceOnlineSensor.UUID+"/features/"+deviceOnlineSensor.Features[0].UUID+"/notifications",
			updateOnlineFeatureNotificationHandler,
		)
		httpListener, errHTTP := net.Listen("tcp", "localhost:8089")
		logger.Infof("online_test - HTTP client listening at %s", httpListener.Addr().String())
		Expect(errHTTP).ShouldNot(HaveOccurred())
		httpOnlineMockServer = httptest.NewUnstartedServer(onlineMux)
		// NewUnstartedServer creates an httpListener, so we need to Close that
		// httpListener and replace it with the one we created.
		httpOnlineMockServer.Listener.Close()
		httpOnlineMockServer.Listener = httpListener
		go func() {
			httpOnlineMockServer.Start()
		}()

		// --------- start a sensor HTTP server ---------
		sensorMux := http.NewServeMux()
		sensorMux.HandleFunc(
			"/sensors/"+deviceOnlineSensor.UUID+"/features/"+deviceOnlineSensor.Features[0].UUID,
			deleteOnlineSensorSensorHandler,
		)
		sensorHTTPListener, errHTTP := net.Listen("tcp", "localhost:8000")
		logger.Infof("sensor_test - HTTP client listening at %s", sensorHTTPListener.Addr().String())
		Expect(errHTTP).ShouldNot(HaveOccurred())
		httpSensorMockServer = httptest.NewUnstartedServer(sensorMux)
		httpSensorMockServer.Listener.Close()
		httpSensorMockServer.Listener = sensorHTTPListener
		go func() {
			httpSensorMockServer.Start()
		}()
	})

	AfterEach(func() {
		grpcMockServer.Stop()
		httpOnlineMockServer.Close()
		httpSensorMockServer.Close()
		notificationPreferenceBody = ""
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

		When("profile session points to a missing profile", func() {
			It("should return an error", func() {
				jwtToken, cookieSession := testuutils.GetJwt(router)
				profileRes := testuutils.GetLoggedProfile(router, jwtToken, cookieSession)

				_, err := collProfiles.DeleteOne(ctx, bson.M{"_id": profileRes.ID})
				Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, "/api/devices", nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
				Expect(recorder.Body.String()).To(Equal(`{"error":"cannot find profile"}`))
			})
		})
	})

	Context("calling devices api DELETE", func() {
		BeforeEach(func() {
			err := testuutils.InsertOne(ctx, collDevices, deviceController)
			Expect(err).ShouldNot(HaveOccurred())
			err = testuutils.InsertOne(ctx, collDevices, deviceOnlineSensor)
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
				Expect(controllerDeleteCalled.Load()).To(BeTrue())
				Expect(sensorDeleteCalled.Load()).To(BeFalse())
				Expect(onlineDeleteCalled.Load()).To(BeFalse())
			})

			It("should fail and keep the device when controller cleanup fails", func() {
				jwtToken, cookieSession := testuutils.GetJwt(router)
				profileRes := testuutils.GetLoggedProfile(router, jwtToken, cookieSession)

				err := testuutils.AssignDeviceToProfile(ctx, collProfiles, profileRes.ID, deviceController.ID)
				Expect(err).ShouldNot(HaveOccurred())
				err = testuutils.AssignHomeToProfile(ctx, collProfiles, profileRes.ID, home.ID)
				Expect(err).ShouldNot(HaveOccurred())

				controllerDeleteShouldFail.Store(true)

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodDelete, "/api/devices/"+deviceController.ID.Hex(), nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
				Expect(recorder.Body.String()).To(Equal(`{"error":"cannot cleanup device state"}`))

				devices, err := testuutils.FindAll[models.Device](ctx, collDevices)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(devices).To(HaveLen(2))
				Expect(controllerDeleteCalled.Load()).To(BeTrue())
			})
		})

		When("profile owns a online sensor", func() {
			It("should remove the online sensor", func() {
				jwtToken, cookieSession := testuutils.GetJwt(router)
				profileRes := testuutils.GetLoggedProfile(router, jwtToken, cookieSession)

				err := testuutils.AssignDeviceToProfile(ctx, collProfiles, profileRes.ID, deviceOnlineSensor.ID)
				Expect(err).ShouldNot(HaveOccurred())
				err = testuutils.AssignHomeToProfile(ctx, collProfiles, profileRes.ID, home.ID)
				Expect(err).ShouldNot(HaveOccurred())

				err = testuutils.AssignDeviceToHomeAndRoom(ctx, collHomes, home.ID, home.Rooms[0].ID, deviceOnlineSensor.ID)
				Expect(err).ShouldNot(HaveOccurred())

				devices, err := testuutils.FindAll[models.Device](ctx, collDevices)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(devices).To(HaveLen(2))

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodDelete, "/api/devices/"+deviceOnlineSensor.ID.Hex(), nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusOK))
				Expect(recorder.Body.String()).To(Equal(`{"message":"device has been deleted"}`))

				devices, err = testuutils.FindAll[models.Device](ctx, collDevices)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(devices).To(HaveLen(1))
				Expect(sensorDeleteCalled.Load()).To(BeTrue())
				Expect(onlineDeleteCalled.Load()).To(BeTrue())
				Expect(controllerDeleteCalled.Load()).To(BeFalse())
			})

			It("should fail and keep the device when sensor cleanup fails", func() {
				jwtToken, cookieSession := testuutils.GetJwt(router)
				profileRes := testuutils.GetLoggedProfile(router, jwtToken, cookieSession)

				err := testuutils.AssignDeviceToProfile(ctx, collProfiles, profileRes.ID, deviceOnlineSensor.ID)
				Expect(err).ShouldNot(HaveOccurred())
				err = testuutils.AssignHomeToProfile(ctx, collProfiles, profileRes.ID, home.ID)
				Expect(err).ShouldNot(HaveOccurred())

				sensorDeleteShouldFail.Store(true)

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodDelete, "/api/devices/"+deviceOnlineSensor.ID.Hex(), nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
				Expect(recorder.Body.String()).To(Equal(`{"error":"cannot cleanup device state"}`))

				devices, err := testuutils.FindAll[models.Device](ctx, collDevices)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(devices).To(HaveLen(2))
				Expect(sensorDeleteCalled.Load()).To(BeTrue())
				Expect(onlineDeleteCalled.Load()).To(BeFalse())
			})

			It("should fail and keep the device when online cleanup fails", func() {
				jwtToken, cookieSession := testuutils.GetJwt(router)
				profileRes := testuutils.GetLoggedProfile(router, jwtToken, cookieSession)

				err := testuutils.AssignDeviceToProfile(ctx, collProfiles, profileRes.ID, deviceOnlineSensor.ID)
				Expect(err).ShouldNot(HaveOccurred())
				err = testuutils.AssignHomeToProfile(ctx, collProfiles, profileRes.ID, home.ID)
				Expect(err).ShouldNot(HaveOccurred())

				onlineDeleteShouldFail.Store(true)

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodDelete, "/api/devices/"+deviceOnlineSensor.ID.Hex(), nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
				Expect(recorder.Body.String()).To(Equal(`{"error":"cannot cleanup device state"}`))

				devices, err := testuutils.FindAll[models.Device](ctx, collDevices)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(devices).To(HaveLen(2))
				Expect(sensorDeleteCalled.Load()).To(BeTrue())
				Expect(onlineDeleteCalled.Load()).To(BeTrue())
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

			It("should return an error, because device is owned but missing from db", func() {
				jwtToken, cookieSession := testuutils.GetJwt(router)
				profileRes := testuutils.GetLoggedProfile(router, jwtToken, cookieSession)
				missingDeviceID := bson.NewObjectID()

				err := testuutils.AssignDeviceToProfile(ctx, collProfiles, profileRes.ID, missingDeviceID)
				Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodDelete, "/api/devices/"+missingDeviceID.Hex(), nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
				Expect(recorder.Body.String()).To(Equal(`{"error":"cannot find device"}`))
			})
		})
	})

	Context("calling device feature notifications api PUT", func() {
		BeforeEach(func() {
			err := testuutils.InsertOne(ctx, collDevices, deviceOnlineSensor)
			Expect(err).ShouldNot(HaveOccurred())
		})

		When("profile owns the device", func() {
			It("should silence feature notifications", func() {
				jwtToken, cookieSession := testuutils.GetJwt(router)
				profileRes := testuutils.GetLoggedProfile(router, jwtToken, cookieSession)

				err := testuutils.AssignDeviceToProfile(ctx, collProfiles, profileRes.ID, deviceOnlineSensor.ID)
				Expect(err).ShouldNot(HaveOccurred())

				body := bytes.NewBufferString(`{"notificationSilenced":true}`)
				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(
					http.MethodPut,
					"/api/devices/"+deviceOnlineSensor.ID.Hex()+"/features/"+deviceOnlineSensor.Features[0].UUID+"/notifications",
					body,
				)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)

				Expect(recorder.Code).To(Equal(http.StatusOK))
				Expect(recorder.Body.String()).To(Equal(`{"message":"feature notification updated"}`))
				Expect(notificationPreferenceBody).To(Equal(`{"notificationSilenced":true}`))

				deviceFromDb, err := testuutils.FindOneById[models.Device](ctx, collDevices, deviceOnlineSensor.ID)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(deviceFromDb.Features[0].NotificationSilenced).To(BeTrue())
			})
		})
	})
})
