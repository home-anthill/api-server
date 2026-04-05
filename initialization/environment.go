package initialization

import (
	"errors"
	"fmt"
	"os"
	"regexp"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

const projectDirName = "api-server"

// InitEnv function
func InitEnv(logger *zap.SugaredLogger) {
	// Load .env file and print variables
	envFile, err := readEnv()
	logger.Debugf("BuildConfig - envFile = %s", envFile)
	if err != nil {
		logger.Error("BuildConfig - failed to load the env file")
		panic("InitEnv - failed to load the env file at ./" + envFile)
	}
	printEnv(logger)
}

func readEnv() (string, error) {
	// solution taken from https://stackoverflow.com/a/68347834/3590376
	projectName := regexp.MustCompile(`^(.*` + projectDirName + `)`)
	currentWorkDirectory, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("cannot get current working directory: %w", err)
	}
	rootPath := projectName.Find([]byte(currentWorkDirectory))
	envFilePath := string(rootPath) + `/.env`
	err = godotenv.Load(envFilePath)
	return envFilePath, err
}

func printEnv(logger *zap.SugaredLogger) {
	if os.Getenv("JWT_PASSWORD") == "" {
		panic(errors.New("'JWT_PASSWORD' environment variable is mandatory"))
	}
	if os.Getenv("JWT_REFRESH_PASSWORD") == "" {
		panic(errors.New("'JWT_REFRESH_PASSWORD' environment variable is mandatory"))
	}

	logger.Infof("ENVIRONMENT = %s", os.Getenv("ENV"))
	logger.Infof("LOG_FOLDER = %s", os.Getenv("LOG_FOLDER"))
	logger.Infof("MONGODB_URL = %s", os.Getenv("MONGODB_URL"))
	logger.Infof("HTTP_SERVER = %s", os.Getenv("HTTP_SERVER"))
	logger.Infof("HTTP_PORT = %s", os.Getenv("HTTP_PORT"))
	logger.Infof("OAUTH2_CALLBACK = %s", os.Getenv("OAUTH2_CALLBACK"))
	logger.Infof("OAUTH2_CLIENTID = %s", os.Getenv("OAUTH2_CLIENTID"))
	logger.Infof("OAUTH2_APP_CALLBACK = %s", os.Getenv("OAUTH2_APP_CALLBACK"))
	logger.Infof("OAUTH2_APP_CLIENTID = %s", os.Getenv("OAUTH2_APP_CLIENTID"))
	logger.Infof("HTTP_CORS = %s", os.Getenv("HTTP_CORS"))
	logger.Infof("HTTP_SENSOR_SERVER = %s", os.Getenv("HTTP_SENSOR_SERVER"))
	logger.Infof("HTTP_SENSOR_PORT = %s", os.Getenv("HTTP_SENSOR_PORT"))
	logger.Infof("HTTP_SENSOR_GETVALUE_API = %s", os.Getenv("HTTP_SENSOR_GETVALUE_API"))
	logger.Infof("HTTP_SENSOR_REGISTER_API = %s", os.Getenv("HTTP_SENSOR_REGISTER_API"))
	logger.Infof("HTTP_SENSOR_KEEPALIVE_API = %s", os.Getenv("HTTP_SENSOR_KEEPALIVE_API"))
	logger.Infof("HTTP_ONLINE_SERVER = %s", os.Getenv("HTTP_ONLINE_SERVER"))
	logger.Infof("HTTP_ONLINE_PORT = %s", os.Getenv("HTTP_ONLINE_PORT"))
	logger.Infof("HTTP_ONLINE_API = %s", os.Getenv("HTTP_ONLINE_API"))
	logger.Infof("HTTP_ONLINE_FCMTOKEN_API = %s", os.Getenv("HTTP_ONLINE_FCMTOKEN_API"))
	logger.Infof("HTTP_ONLINE_KEEPALIVE_API = %s", os.Getenv("HTTP_ONLINE_KEEPALIVE_API"))
	logger.Infof("GRPC_URL = %s", os.Getenv("GRPC_URL"))
	logger.Infof("GRPC_TLS = %s", os.Getenv("GRPC_TLS"))
	logger.Infof("CERT_FOLDER_PATH = %s", os.Getenv("CERT_FOLDER_PATH"))
	logger.Infof("SINGLE_USER_LOGIN_EMAIL = %s", os.Getenv("SINGLE_USER_LOGIN_EMAIL"))
	logger.Infof("INTERNAL_CLUSTER_PATH = %s", os.Getenv("INTERNAL_CLUSTER_PATH"))
}
