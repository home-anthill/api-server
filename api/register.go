package api

import (
	"api-server/api/grpc/register"
	"api-server/customerrors"
	"api-server/db"
	"api-server/models"
	"api-server/utils"
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"io"
	"net/http"
	"os"
	"time"
)

// FeatureReq struct
type FeatureReq struct {
	Type   models.Type `json:"type" validate:"required,oneof='controller' 'sensor'"`
	Name   string      `json:"name" validate:"required,min=2,max=20"`
	Enable bool        `json:"enable" validate:"required,boolean"`
	Order  int         `json:"order" validate:"required,gte=1"`
	Unit   string      `json:"unit" validate:"required,min=1,max=10"`
}

// DeviceRegisterReq struct
type DeviceRegisterReq struct {
	Mac          string       `json:"mac" validate:"required,mac"`
	Manufacturer string       `json:"manufacturer" validate:"required,min=3,max=50"`
	Model        string       `json:"model" validate:"required,min=3,max=20"`
	APIToken     string       `json:"apiToken" validate:"required,uuid4"`
	Features     []FeatureReq `json:"features" validate:"required,dive"`
}

// SensorRegisterReq struct
type SensorRegisterReq struct {
	UUID           string `json:"uuid"`
	Mac            string `json:"mac"`
	Manufacturer   string `json:"manufacturer"`
	Model          string `json:"model"`
	ProfileOwnerID string `json:"profileOwnerId"`
	APIToken       string `json:"apiToken"`
}

// Register struct
type Register struct {
	client             *mongo.Client
	collDevices        *mongo.Collection
	collProfiles       *mongo.Collection
	ctx                context.Context
	logger             *zap.SugaredLogger
	grpcTarget         string
	keepAliveSensorURL string
	registerSensorURL  string
	validate           *validator.Validate
}

// ControllerDeviceValidation function
func ControllerDeviceValidation(sl validator.StructLevel) {
	deviceReq, _ := sl.Current().Interface().(DeviceRegisterReq)
	var isController = false
	for _, feature := range deviceReq.Features {
		if feature.Type == models.Controller {
			isController = true
		}
	}
	if isController && len(deviceReq.Features) > 1 {
		sl.ReportError(deviceReq.Features, "features", "Features", "features", "")
	}
}

// NewRegister function
func NewRegister(ctx context.Context, logger *zap.SugaredLogger, client *mongo.Client, validate *validator.Validate) *Register {
	grpcURL := os.Getenv("GRPC_URL")
	sensorServerURL := os.Getenv("HTTP_SENSOR_SERVER") + ":" + os.Getenv("HTTP_SENSOR_PORT")
	keepAliveSensorURL := sensorServerURL + os.Getenv("HTTP_SENSOR_KEEPALIVE_API")
	registerSensorURL := sensorServerURL + os.Getenv("HTTP_SENSOR_REGISTER_API")

	// add custom validators
	// validator to check that a controller device can have only one feature
	validate.RegisterStructValidation(ControllerDeviceValidation, DeviceRegisterReq{})

	return &Register{
		client:             client,
		collDevices:        db.GetCollections(client).Devices,
		collProfiles:       db.GetCollections(client).Profiles,
		ctx:                ctx,
		logger:             logger,
		grpcTarget:         grpcURL,
		keepAliveSensorURL: keepAliveSensorURL,
		registerSensorURL:  registerSensorURL,
		validate:           validate,
	}
}

// PostRegister function
func (handler *Register) PostRegister(c *gin.Context) {
	handler.logger.Info("REST - PostRegister called")

	var registerBody DeviceRegisterReq
	if err := c.ShouldBindJSON(&registerBody); err != nil {
		handler.logger.Errorf("REST - PostRegister - Cannot bind request body. Err = %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	err := handler.validate.Struct(registerBody)
	if err != nil {
		handler.logger.Errorf("REST - PostRegister - request body is not valid, err %#v", err)
		var errFields = utils.GetErrorMessage(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body, these fields are not valid:" + errFields})
		return
	}

	// search if profile token exists and retrieve profile
	var profileFound models.Profile
	errProfile := handler.collProfiles.FindOne(handler.ctx, bson.M{
		"apiToken": registerBody.APIToken,
	}).Decode(&profileFound)
	if errProfile != nil {
		handler.logger.Errorf("REST - PostRegister - Cannot find profile with that apiToken. Err = %v\n", errProfile)
		c.JSON(http.StatusBadRequest, gin.H{"error": "cnnot register, profile token missing or not valid"})
		return
	}

	// search and skip db add if device already exists
	var device models.Device
	err = handler.collDevices.FindOne(handler.ctx, bson.M{
		"mac": registerBody.Mac,
	}).Decode(&device)
	if err == nil {
		handler.logger.Info("REST - PostRegister - Device already registered")
		// if err == nil => device found in db (already exists)
		// skip register process returning "already registered"
		c.JSON(http.StatusConflict, gin.H{"message": "Already registered"})
		return
	}

	isController := false
	for _, feature := range registerBody.Features {
		if feature.Type == models.Controller {
			isController = true
		}
	}

	handler.logger.Debugf("REST - PostRegister - registering device, isController = %v", isController)

	insertDate := time.Now()
	device = models.Device{}
	device.ID = primitive.NewObjectID()
	device.UUID = uuid.NewString()
	device.Mac = registerBody.Mac
	device.Manufacturer = registerBody.Manufacturer
	device.Model = registerBody.Model
	device.CreatedAt = insertDate
	device.ModifiedAt = insertDate

	device.Features = utils.MapSlice(registerBody.Features, func(fReq FeatureReq) models.Feature {
		return models.Feature{
			UUID:   uuid.NewString(),
			Type:   fReq.Type,
			Name:   fReq.Name,
			Enable: fReq.Enable,
			Order:  fReq.Order,
			Unit:   fReq.Unit,
		}
	})

	// if it's a controller device => call gRPC
	// otherwise REST call to sensor service
	if isController {
		status, message, err := handler.registerControllerViaGRPC(&device, &profileFound)
		if err != nil {
			handler.logger.Errorf("REST - PostRegister - cannot register controller device via gRPC. Err %v\n", err)
			if re, ok := err.(*customerrors.ErrorWrapper); ok {
				handler.logger.Errorf("REST - PostRegister - cannot register device with status = %d, message = %s\n", re.Code, re.Message)
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot register controller device"})
			return
		}
		handler.logger.Infof("REST - PostRegister - controller device registered with status = %s, message = %s\n", status, message)
	} else {
		err := handler.registerSensorViaHTTP(&device, &profileFound)
		if err != nil {
			handler.logger.Errorf("REST - PostRegister - cannot register sensor device via HTTP. Err %v\n", err)
			if re, ok := err.(*customerrors.ErrorWrapper); ok {
				handler.logger.Errorf("REST - PostRegister - cannot register device with status = %d, message = %s\n", re.Code, re.Message)
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot register sensor device"})
			return
		}
		handler.logger.Info("REST - PostRegister - sensor device registered successfully")
	}

	handler.logger.Debugf("REST - PostRegister - registered device = %#v", device)
	c.JSON(http.StatusOK, device)
}

func (handler *Register) registerSensorViaHTTP(device *models.Device, profileFound *models.Profile) error {
	// check if service is available calling keep-alive
	// TODO remove this in a production code
	_, _, keepAliveErr := handler.keepAliveSensorService(handler.keepAliveSensorURL)
	if keepAliveErr != nil {
		return customerrors.Wrap(http.StatusInternalServerError, keepAliveErr, "Cannot call keepAlive of remote register service")
	}

	// Insert device into api-server database
	errInsDb := handler.insertDevice(device, profileFound)
	if errInsDb != nil {
		return customerrors.Wrap(http.StatusInternalServerError, errInsDb, "Cannot insert the new device")
	}

	// do the real call to the remote registration service
	payload := SensorRegisterReq{
		UUID:           device.UUID,
		Mac:            device.Mac,
		Manufacturer:   device.Manufacturer,
		Model:          device.Model,
		ProfileOwnerID: profileFound.ID.Hex(),
		APIToken:       profileFound.APIToken,
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return customerrors.Wrap(http.StatusInternalServerError, err, "Cannot create payload to register sensor service")
	}

	for _, feature := range device.Features {
		_, _, err := handler.registerSensor(handler.registerSensorURL+feature.Name, payloadJSON)
		if err != nil {
			return customerrors.Wrap(http.StatusInternalServerError, err, "Cannot register sensor device feature "+feature.Name)
		}
		//handler.logger.Debugf("REST - PostRegister - sensor device registered with status= %d, body= %s\n", statusCode, respBody)
	}
	return nil
}

func (handler *Register) registerControllerViaGRPC(device *models.Device, profileFound *models.Profile) (string, string, error) {
	handler.logger.Info("gRPC - registerControllerViaGRPC - Sending register via gRPC...")
	// Set up a connection to the gRPC server.
	securityDialOption, isSecure, err := utils.BuildSecurityDialOption()
	if err != nil {
		return "", "", customerrors.Wrap(http.StatusInternalServerError, err, "Cannot create securityDialOption to prepare the gRPC connection")
	}
	if isSecure {
		handler.logger.Debug("registerControllerViaGRPC - GRPC secure enabled!")
	} else {
		handler.logger.Info("registerControllerViaGRPC - GRPC secure NOT enabled!")
	}

	handler.logger.Debugf("gRPC - registerControllerViaGRPC - handler.grpcTarget = %s", handler.grpcTarget)

	conn, err := grpc.NewClient(handler.grpcTarget, securityDialOption)
	if err != nil {
		handler.logger.Error("gRPC - registerControllerViaGRPC - cannot connect via gRPC", err)
		return "", "", customerrors.GrpcSendError{
			Status:  customerrors.ConnectionError,
			Message: "Cannot connect to api-devices",
		}
	}
	defer conn.Close()
	client := register.NewRegistrationClient(conn)

	// -------------------------------------------------------
	// I reach this point only if I can connect to gRPC SERVER
	// -------------------------------------------------------

	// Insert device into api-server database
	errInsDb := handler.insertDevice(device, profileFound)
	if errInsDb != nil {
		return "", "", customerrors.Wrap(http.StatusInternalServerError, errInsDb, "Cannot insert the new device")
	}

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	r, err := client.Register(ctx, &register.RegisterRequest{
		Id:             device.ID.Hex(),
		Uuid:           device.UUID,
		Mac:            device.Mac,
		Name:           device.Features[0].Name,
		Manufacturer:   device.Manufacturer,
		Model:          device.Model,
		ProfileOwnerId: profileFound.ID.Hex(),
		ApiToken:       profileFound.APIToken,
	})
	if err != nil {
		return "", "", customerrors.Wrap(http.StatusInternalServerError, err, "Cannot invoke Register via gRPC")
	}
	return r.GetStatus(), r.GetMessage(), nil
}

func (handler *Register) keepAliveSensorService(urlKeepAlive string) (int, string, error) {
	response, err := http.Get(urlKeepAlive)
	if err != nil {
		return -1, "", customerrors.Wrap(http.StatusInternalServerError, err, "Cannot call keepAlive of the remote register service via HTTP")
	}
	defer response.Body.Close()
	body, _ := io.ReadAll(response.Body)
	return response.StatusCode, string(body), nil
}

func (handler *Register) registerSensor(urlRegister string, payloadJSON []byte) (int, string, error) {
	var payloadBody = bytes.NewBuffer(payloadJSON)
	response, err := http.Post(urlRegister, "application/json", payloadBody)
	if err != nil {
		return -1, "", customerrors.Wrap(http.StatusInternalServerError, err, "Cannot register to sensor service via HTTP")
	}
	defer response.Body.Close()
	body, _ := io.ReadAll(response.Body)
	return response.StatusCode, string(body), nil
}

func (handler *Register) insertDevice(device *models.Device, profile *models.Profile) error {
	// start-session
	dbSession, err := handler.client.StartSession()
	if err != nil {
		handler.logger.Errorf("insertDevice - cannot start a db session, err = %#v", err)
		return customerrors.Wrap(http.StatusInternalServerError, err, "unknown error while trying to register a device")
	}
	// Defers ending the session after the transaction is committed or ended
	defer dbSession.EndSession(context.TODO())

	_, errTrans := dbSession.WithTransaction(context.TODO(), func(sessionCtx mongo.SessionContext) (interface{}, error) {
		// Official `mongo-driver` documentation state: "callback may be run
		// multiple times during WithTransaction due to retry attempts, so it must be idempotent."

		// Insert device
		_, errInsert := handler.collDevices.InsertOne(sessionCtx, device)
		if errInsert != nil {
			return nil, customerrors.Wrap(http.StatusInternalServerError, errInsert, "Cannot insert the new device")
		}
		// push device.ID to profile.devices into api-server database
		_, errUpd := handler.collProfiles.UpdateOne(
			sessionCtx,
			bson.M{"_id": profile.ID},
			bson.M{"$push": bson.M{"devices": device.ID}},
		)
		if errUpd != nil {
			return nil, customerrors.Wrap(http.StatusInternalServerError, errUpd, "Cannot update profile with the new device")
		}
		return nil, nil
	}, options.Transaction().SetWriteConcern(writeconcern.Majority()))

	if errTrans != nil {
		handler.logger.Errorf("insertDevice - insert device in transaction, errTrans = %#v", errTrans)
	}
	return errTrans
}
