package api

import (
	pb "api-server/api/grpc/device"
	"api-server/customerrors"
	"api-server/db"
	"api-server/models"
	"api-server/utils"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// DevicesValues handles reading and writing feature values for devices.
type DevicesValues struct {
	client            *mongo.Client
	collDevices       *mongo.Collection
	collProfiles      *mongo.Collection
	collHomes         *mongo.Collection
	logger            *zap.SugaredLogger
	grpcTarget        string
	sensorGetValueURL string
	validate          *validator.Validate
}

// NewDevicesValues constructs a DevicesValues handler with the given dependencies.
func NewDevicesValues(logger *zap.SugaredLogger, client *mongo.Client, validate *validator.Validate) *DevicesValues {
	grpcURL := os.Getenv("GRPC_URL")
	sensorServerURL := os.Getenv("HTTP_SENSOR_SERVER") + ":" + os.Getenv("HTTP_SENSOR_PORT")
	sensorGetValueURL := sensorServerURL + os.Getenv("HTTP_SENSOR_GETVALUE_API")

	return &DevicesValues{
		client:            client,
		collDevices:       db.GetCollections(client).Devices,
		collProfiles:      db.GetCollections(client).Profiles,
		collHomes:         db.GetCollections(client).Homes,
		logger:            logger,
		grpcTarget:        grpcURL,
		sensorGetValueURL: sensorGetValueURL,
		validate:          validate,
	}
}

// ------------------------------ Public methods ------------------------------

// GetValuesDevice function
func (dv *DevicesValues) GetValuesDevice(c *gin.Context) {
	dv.logger.Info("REST - GET - GetValuesDevice called")

	objectID, errID := bson.ObjectIDFromHex(c.Param("id"))
	if errID != nil {
		dv.logger.Error("REST - GET - GetValuesDevice - wrong format of the path param 'id'")
		c.JSON(http.StatusBadRequest, gin.H{"error": "wrong format of the path param 'id'"})
		return
	}

	// retrieve current profile object from database (using session profile as input)
	session := sessions.Default(c)
	profile, err := utils.GetLoggedProfile(c.Request.Context(), &session, dv.collProfiles)
	if err != nil {
		dv.logger.Error("REST - GET - GetValuesDevice - cannot find profile in session")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "cannot find profile in session"})
		return
	}

	// check if device is in profile (device owned by profile)

	if !utils.Contains(profile.Devices, objectID) {
		dv.logger.Error("REST - GET - GetValuesDevice - this device is not in your profile")
		c.JSON(http.StatusBadRequest, gin.H{"error": "this device is not in your profile"})
		return
	}
	// get device from db
	device, err := dv.getDevice(c.Request.Context(), objectID)
	if err != nil {
		dv.logger.Error("REST - GET - GetValuesDevice - cannot find device")
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot find device"})
		return
	}

	var deviceFeatureStates []models.DeviceFeatureState
	for _, feature := range device.Features {
		dv.logger.Debugf("REST - GET - GetValuesDevice - feature = %v", feature)
		if feature.Type == models.Controller {
			state, err := dv.getControllerValue(&device, &feature, profile.APIToken)
			if err != nil {
				dv.logger.Errorf("REST - GET - GetValuesDevice - cannot get values via gRPC, err = %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot get values"})
				return
			}
			deviceFeatureStates = append(deviceFeatureStates, *state)
		} else {
			if !utils.IsValidUUID(device.UUID) || !utils.IsValidUUID(feature.UUID) {
				dv.logger.Errorf("REST - GetValuesDevice - invalid UUID format: device=%s feature=%s", device.UUID, feature.UUID)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot get sensor value"})
				return
			}
			path := dv.sensorGetValueURL + device.UUID + "/features/" + feature.UUID + "/" + feature.Name
			dv.logger.Debugf("REST - GetValuesDevice - path = %s\n", path)
			_, result, err := utils.Get(path)
			if err != nil {
				dv.logger.Errorf("REST - GetValuesDevice - cannot get sensor value from remote service = %#v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot get sensor value"})
				return
			}

			sensorFeatureValue := models.DeviceFeatureState{}
			err = json.Unmarshal([]byte(result), &sensorFeatureValue)
			if err != nil {
				dv.logger.Errorf("REST - GetValuesDevice - cannot unmarshal JSON response from sensor value remote service = %#v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot get sensor value"})
				return
			}
			// add to the object with the value also other information
			// to associate the value to the specific feature
			sensorFeatureValue.FeatureUUID = feature.UUID
			sensorFeatureValue.Type = feature.Type
			sensorFeatureValue.Name = feature.Name

			dv.logger.Debugf("REST - GetValuesDevice - sensor value for feature = %s is = %#v\n", feature.Name, sensorFeatureValue)
			deviceFeatureStates = append(deviceFeatureStates, sensorFeatureValue)
		}
	}
	c.JSON(http.StatusOK, deviceFeatureStates)
}

// PostValuesDevice function
func (dv *DevicesValues) PostValuesDevice(c *gin.Context) {
	dv.logger.Info("REST - POST - PostValuesDevice called")

	objectID, errID := bson.ObjectIDFromHex(c.Param("id"))
	if errID != nil {
		dv.logger.Errorf("REST - GET - PostValuesDevice - wrong format of the path param 'id', err %#v", errID)
		c.JSON(http.StatusBadRequest, gin.H{"error": "wrong format of the path param 'id'"})
		return
	}

	var featureStates []models.DeviceFeatureState
	if err := c.ShouldBindJSON(&featureStates); err != nil {
		dv.logger.Errorf("REST - POST - PostValuesDevice - invalid request payload, err %#v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	for _, fs := range featureStates {
		err := dv.validate.Struct(fs)
		if err != nil {
			dv.logger.Errorf("REST - POST - PostValuesDevice - request body is not valid, err %#v", err)
			var errFields = utils.GetErrorMessage(err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body, these fields are not valid:" + errFields})
			return
		}
	}

	// retrieve current profile object from database (using session profile as input)
	session := sessions.Default(c)
	profile, err := utils.GetLoggedProfile(c.Request.Context(), &session, dv.collProfiles)
	if err != nil {
		dv.logger.Errorf("REST - GET - PostValuesDevice - cannot find profile in session, err %#v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "cannot find profile in session"})
		return
	}

	// check if device is in profile (device owned by profile)
	if !utils.Contains(profile.Devices, objectID) {
		dv.logger.Error("REST - POST - PostValuesDevice - this is not your device")
		c.JSON(http.StatusBadRequest, gin.H{"error": "this device is not in your profile"})
		return
	}
	// get device from db
	device, err := dv.getDevice(c.Request.Context(), objectID)
	if err != nil {
		dv.logger.Errorf("REST - POST - PostValuesDevice - cannot find device, err %#v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot find device"})
		return
	}
	// send via gRPC
	err = dv.sendViaGrpc(&device, featureStates, profile.APIToken)
	if err != nil {
		dv.logger.Errorf("REST - POST - PostValuesDevice - cannot set values via gRPC, err %#v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot set value"})
		return
	}

	dv.logger.Infow("AUDIT - device values set",
		"profileID", profile.ID.Hex(),
		"deviceID", objectID.Hex(),
		"clientIP", c.ClientIP(),
	)
	c.JSON(http.StatusOK, gin.H{"message": "set values success"})
}

// ------------------------------ Private methods ------------------------------

// getControllerValue calls gRPC to get a single controller feature value.
// Having this as a separate function ensures that defer runs at the end of
// each call, avoiding the resource leak that occurs with defer inside a for-loop.
func (dv *DevicesValues) getControllerValue(device *models.Device, feature *models.Feature, apiToken string) (*models.DeviceFeatureState, error) {
	conn, err := grpc.NewClient(dv.grpcTarget, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := pb.NewDeviceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	response, err := client.GetValue(ctx, &pb.GetValueRequest{
		Id:          device.ID.Hex(),
		DeviceUuid:  device.UUID,
		FeatureUuid: feature.UUID,
		FeatureName: feature.Name,
		Mac:         device.Mac,
		ApiToken:    apiToken,
	})
	if err != nil {
		return nil, err
	}

	return &models.DeviceFeatureState{
		FeatureUUID: feature.UUID,
		Name:        feature.Name,
		Type:        feature.Type,
		Value:       response.Value,
		CreatedAt:   response.CreatedAt,
		ModifiedAt:  response.ModifiedAt,
	}, nil
}

func (dv *DevicesValues) sendViaGrpc(device *models.Device, featureStates []models.DeviceFeatureState, apiToken string) error {
	dv.logger.Infof("gRPC - sendViaGrpc - Called with featureStates = %#v", featureStates)

	// Set up a connection to the gRPC server.
	securityDialOption, isSecure, err := utils.BuildSecurityDialOption()
	if err != nil {
		dv.logger.Errorf("gRPC - sendViaGrpc - Cannot build security dial option object!, err %#v", err)
		return customerrors.Wrap(http.StatusInternalServerError, err, "Cannot create securityDialOption to prepare the gRPC connection")
	}
	if isSecure {
		dv.logger.Info("gRPC - sendViaGrpc - GRPC secure enabled!")
	} else {
		dv.logger.Info("gRPC - sendViaGrpc - GRPC secure NOT enabled!")
	}

	conn, err := grpc.NewClient(dv.grpcTarget, securityDialOption)
	if err != nil {
		dv.logger.Errorf("gRPC - sendViaGrpc - cannot connect via gRPC, err %#v", err)
		return customerrors.GrpcSendError{
			Status:  customerrors.ConnectionError,
			Message: "Cannot connect to api-devices",
		}
	}
	defer conn.Close()
	client := pb.NewDeviceClient(conn)

	// -------------------------------------------------------
	// I reach this point only if I can connect to gRPC SERVER
	// -------------------------------------------------------
	dv.logger.Info("gRPC - sendViaGrpc - gRPC server connected")

	const grpcClientDeadline = 200 * time.Millisecond
	const grpcContextTimeout = 5 * time.Second
	clientDeadline := time.Now().Add(grpcClientDeadline)
	contextBg, cancelBg := context.WithTimeout(context.Background(), grpcContextTimeout)
	defer cancelBg()
	ctx, cancel := context.WithDeadline(contextBg, clientDeadline)
	defer cancel()

	requests := utils.MapSlice(featureStates, func(featureState models.DeviceFeatureState) *pb.SetValueRequest {
		return &pb.SetValueRequest{
			FeatureUuid: featureState.FeatureUUID,
			FeatureName: featureState.Name,
			Value:       featureState.Value,
		}
	})
	dv.logger.Debugf("gRPC - sendViaGrpc - requests request = %#v", requests)

	response, errSend := client.SetValues(ctx, &pb.SetValuesRequest{
		Id:            device.ID.Hex(),
		DeviceUuid:    device.UUID,
		Mac:           device.Mac,
		ApiToken:      apiToken,
		FeatureValues: requests,
	})
	if errSend != nil {
		return errSend
	}

	dv.logger.Debugf("gRPC - sendViaGrpc - Device set value response %#v", response)
	return nil
}

func (dv *DevicesValues) getDevice(ctx context.Context, deviceID bson.ObjectID) (models.Device, error) {
	dv.logger.Debug("getDevice - searching device with objectId: ", deviceID)
	var device models.Device
	err := dv.collDevices.FindOne(ctx, bson.M{
		"_id": deviceID,
	}).Decode(&device)
	dv.logger.Debug("Device found: ", device)
	return device, err
}
