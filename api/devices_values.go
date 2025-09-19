package api

import (
	device3 "api-server/api/grpc/device"
	"api-server/customerrors"
	"api-server/db"
	"api-server/models"
	"api-server/utils"
	"encoding/json"
	"net/http"
	"os"
	"reflect"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// DevicesValues struct
type DevicesValues struct {
	client            *mongo.Client
	collDevices       *mongo.Collection
	collProfiles      *mongo.Collection
	collHomes         *mongo.Collection
	ctx               context.Context
	logger            *zap.SugaredLogger
	grpcTarget        string
	sensorGetValueURL string
	validate          *validator.Validate
}

// NewDevicesValues function
func NewDevicesValues(ctx context.Context, logger *zap.SugaredLogger, client *mongo.Client, validate *validator.Validate) *DevicesValues {
	grpcURL := os.Getenv("GRPC_URL")
	sensorServerURL := os.Getenv("HTTP_SENSOR_SERVER") + ":" + os.Getenv("HTTP_SENSOR_PORT")
	sensorGetValueURL := sensorServerURL + os.Getenv("HTTP_SENSOR_GETVALUE_API")

	return &DevicesValues{
		client:            client,
		collDevices:       db.GetCollections(client).Devices,
		collProfiles:      db.GetCollections(client).Profiles,
		collHomes:         db.GetCollections(client).Homes,
		ctx:               ctx,
		logger:            logger,
		grpcTarget:        grpcURL,
		sensorGetValueURL: sensorGetValueURL,
		validate:          validate,
	}
}

// ------------------------------ Public methods ------------------------------

// GetValuesDevice function
func (handler *DevicesValues) GetValuesDevice(c *gin.Context) {
	handler.logger.Info("REST - GET - GetValuesDevice called")

	objectID, errID := primitive.ObjectIDFromHex(c.Param("id"))
	if errID != nil {
		handler.logger.Error("REST - GET - GetValuesDevice - wrong format of the path param 'id'")
		c.JSON(http.StatusBadRequest, gin.H{"error": "wrong format of the path param 'id'"})
		return
	}

	// retrieve current profile object from database (using session profile as input)
	session := sessions.Default(c)
	profile, err := utils.GetLoggedProfile(handler.ctx, &session, handler.collProfiles)
	if err != nil {
		handler.logger.Error("REST - GET - GetValuesDevice - cannot find profile in session")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "cannot find profile in session"})
		return
	}

	// check if device is in profile (device owned by profile)

	if !utils.Contains(profile.Devices, objectID) {
		handler.logger.Error("REST - GET - GetValuesDevice - this device is not in your profile")
		c.JSON(http.StatusBadRequest, gin.H{"error": "this device is not in your profile"})
		return
	}
	// get device from db
	device, err := handler.getDevice(objectID)
	if err != nil {
		handler.logger.Error("REST - GET - GetValuesDevice - cannot find device")
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot find device"})
		return
	}

	var deviceFeatureStates []models.DeviceFeatureState
	for _, feature := range device.Features {
		handler.logger.Debugf("REST - GET - GetValuesDevice - feature = %v", feature)
		if feature.Type == models.Controller {
			// Set up a connection to the server.
			conn, err := grpc.NewClient(handler.grpcTarget, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				handler.logger.Errorf("REST - GET - GetValuesDevice - cannot establish gRPC connection, err = %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot get values - connection error"})
				return
			}
			defer conn.Close() // FIXME Possible resource leak, 'defer' is called in the 'for' loop
			client := device3.NewDeviceClient(conn)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel() // FIXME Possible resource leak, 'defer' is called in the 'for' loop

			response, errSend := client.GetValue(ctx, &device3.GetValueRequest{
				Id:          device.ID.Hex(),
				FeatureUuid: feature.UUID,
				FeatureName: feature.Name,
				Mac:         device.Mac,
				ApiToken:    profile.APIToken,
			})

			if errSend != nil {
				handler.logger.Errorf("REST - GET - GetValuesDevice - cannot get values via gRPC, err = %v", errSend)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot get values"})
				return
			}

			deviceFeatureStates = append(deviceFeatureStates, models.DeviceFeatureState{
				FeatureUUID: feature.UUID,
				Type:        feature.Type,
				Name:        feature.Name,
				Value:       response.Value,
				CreatedAt:   response.CreatedAt,
				ModifiedAt:  response.ModifiedAt,
			})
		} else {
			path := handler.sensorGetValueURL + device.UUID + "/" + feature.Name
			_, result, err := utils.Get(path)
			if err != nil {
				handler.logger.Errorf("REST - GetValuesDevice - cannot get sensor value from remote service = %#v", err)
				// TODO manage errors
				// return custom_errors.Wrap(http.StatusInternalServerError, err, "Cannot register sensor device feature "+feature.Name)
			}

			sensorFeatureValue := models.DeviceFeatureState{}
			err = json.Unmarshal([]byte(result), &sensorFeatureValue)
			if err != nil {
				handler.logger.Errorf("REST - GetValuesDevice - cannot unmarshal JSON response from sensor value remote service = %#v", err)
				// TODO manage errors
			}
			// add to the object with the value also other information
			// to associate the value to the specific feature
			sensorFeatureValue.FeatureUUID = feature.UUID
			sensorFeatureValue.Type = feature.Type
			sensorFeatureValue.Name = feature.Name

			handler.logger.Debugf("REST - GetValuesDevice - sensor value for feature = %s is = %#v\n", feature.Name, sensorFeatureValue)
			deviceFeatureStates = append(deviceFeatureStates, sensorFeatureValue)
		}
	}
	c.JSON(http.StatusOK, deviceFeatureStates)
}

// PostValueDevice function
func (handler *DevicesValues) PostValueDevice(c *gin.Context) {
	handler.logger.Info("REST - POST - PostValueDevice called")

	objectID, errID := primitive.ObjectIDFromHex(c.Param("id"))
	if errID != nil {
		handler.logger.Errorf("REST - GET - PostValueDevice - wrong format of the path param 'id', err %#v", errID)
		c.JSON(http.StatusBadRequest, gin.H{"error": "wrong format of the path param 'id'"})
		return
	}

	var featureState models.DeviceFeatureState
	if err := c.ShouldBindJSON(&featureState); err != nil {
		handler.logger.Errorf("REST - POST - PostValueDevice - invalid request payload, err %#v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	err := handler.validate.Struct(featureState)
	if err != nil {
		handler.logger.Errorf("REST - POST - PostValueDevice - request body is not valid, err %#v", err)
		var errFields = utils.GetErrorMessage(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body, these fields are not valid:" + errFields})
		return
	}

	// retrieve current profile object from database (using session profile as input)
	session := sessions.Default(c)
	profile, err := utils.GetLoggedProfile(handler.ctx, &session, handler.collProfiles)
	if err != nil {
		handler.logger.Errorf("REST - GET - PostValueDevice - cannot find profile in session, err %#v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "cannot find profile in session"})
		return
	}

	// check if device is in profile (device owned by profile)
	if !utils.Contains(profile.Devices, objectID) {
		handler.logger.Error("REST - POST - PostValueDevice - this is not your device")
		c.JSON(http.StatusBadRequest, gin.H{"error": "this device is not in your profile"})
		return
	}
	// get device from db
	device, err := handler.getDevice(objectID)
	if err != nil {
		handler.logger.Errorf("REST - POST - PostValueDevice - cannot find device, err %#v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot find device"})
		return
	}
	// send via gRPC
	err = handler.sendViaGrpc(&device, &featureState, profile.APIToken)
	if err != nil {
		handler.logger.Errorf("REST - POST - PostValueDevice - cannot set value via gRPC, err %#v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot set value"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "set value success"})
}

// ------------------------------ Private methods ------------------------------

func (handler *DevicesValues) sendViaGrpc(device *models.Device, featureState *models.DeviceFeatureState, apiToken string) error {
	handler.logger.Infof("gRPC - sendViaGrpc - Called with value = %#v and apiToken = %s", featureState, apiToken)

	// Set up a connection to the gRPC server.
	securityDialOption, isSecure, err := utils.BuildSecurityDialOption()
	if err != nil {
		handler.logger.Errorf("gRPC - sendViaGrpc - Cannot build security dial option object!, err %#v", err)
		return customerrors.Wrap(http.StatusInternalServerError, err, "Cannot create securityDialOption to prepare the gRPC connection")
	}
	if isSecure {
		handler.logger.Info("gRPC - sendViaGrpc - GRPC secure enabled!")
	} else {
		handler.logger.Info("gRPC - sendViaGrpc - GRPC secure NOT enabled!")
	}

	conn, err := grpc.NewClient(handler.grpcTarget, securityDialOption)
	if err != nil {
		handler.logger.Errorf("gRPC - sendViaGrpc - cannot connect via gRPC, err %#v", err)
		return customerrors.GrpcSendError{
			Status:  customerrors.ConnectionError,
			Message: "Cannot connect to api-devices",
		}
	}
	defer conn.Close()
	client := device3.NewDeviceClient(conn)

	// -------------------------------------------------------
	// I reach this point only if I can connect to gRPC SERVER
	// -------------------------------------------------------
	handler.logger.Info("gRPC - sendViaGrpc - gRPC server connected")

	clientDeadline := time.Now().Add(time.Duration(200) * time.Millisecond)
	contextBg, cancelBg := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelBg()
	ctx, cancel := context.WithDeadline(contextBg, clientDeadline)
	defer cancel()
	handler.logger.Infof("gRPC - sendViaGrpc - getType(value) %s", getType(featureState))

	response, errSend := client.SetValue(ctx, &device3.SetValueRequest{
		Id:          device.ID.Hex(),
		FeatureUuid: featureState.FeatureUUID,
		FeatureName: featureState.Name,
		Mac:         device.Mac,
		ApiToken:    apiToken,
		Value:       featureState.Value,
	})
	handler.logger.Debugf("gRPC - sendViaGrpc - Device set value status %s", response.GetStatus())
	handler.logger.Debugf("gRPC - sendViaGrpc - Device set value message %s", response.GetMessage())
	return errSend
}

func (handler *DevicesValues) getDevice(deviceID primitive.ObjectID) (models.Device, error) {
	handler.logger.Debug("getDevice - searching device with objectId: ", deviceID)
	var device models.Device
	err := handler.collDevices.FindOne(handler.ctx, bson.M{
		"_id": deviceID,
	}).Decode(&device)
	handler.logger.Debug("Device found: ", device)
	return device, err
}

func getType(value interface{}) string {
	t := reflect.TypeOf(value)
	if t.Kind() == reflect.Ptr {
		return "*" + t.Elem().Name()
	}
	return t.Name()
}
