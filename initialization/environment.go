package initialization

import (
	"fmt"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"os"
	"regexp"
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
	currentWorkDirectory, _ := os.Getwd()
	rootPath := projectName.Find([]byte(currentWorkDirectory))
	envFilePath := string(rootPath) + `/.env`
	err := godotenv.Load(envFilePath)
	return envFilePath, err
}

func printEnv(logger *zap.SugaredLogger) {
	if os.Getenv("JWT_PASSWORD") == "" {
		panic(fmt.Errorf("'JWT_PASSWORD' environment variable is mandatory"))
	}

	logger.Info("ENVIRONMENT = " + os.Getenv("ENV"))
	logger.Info("MONGODB_URL = " + os.Getenv("MONGODB_URL"))
	logger.Info("HTTP_SERVER = " + os.Getenv("HTTP_SERVER"))
	logger.Info("HTTP_PORT = " + os.Getenv("HTTP_PORT"))
	logger.Info("OAUTH2_CALLBACK = " + os.Getenv("OAUTH2_CALLBACK"))
	logger.Info("OAUTH2_CLIENTID = " + os.Getenv("OAUTH2_CLIENTID"))
	logger.Info("OAUTH2_SECRETID = " + os.Getenv("OAUTH2_SECRETID"))
	logger.Info("OAUTH2_APP_CALLBACK = " + os.Getenv("OAUTH2_APP_CALLBACK"))
	logger.Info("OAUTH2_APP_CLIENTID = " + os.Getenv("OAUTH2_APP_CLIENTID"))
	logger.Info("OAUTH2_APP_SECRETID = " + os.Getenv("OAUTH2_APP_SECRETID"))
	logger.Info("HTTP_CORS = " + os.Getenv("HTTP_CORS"))
	logger.Info("HTTP_SENSOR_SERVER = " + os.Getenv("HTTP_SENSOR_SERVER"))
	logger.Info("HTTP_SENSOR_PORT = " + os.Getenv("HTTP_SENSOR_PORT"))
	logger.Info("HTTP_SENSOR_GETVALUE_API = " + os.Getenv("HTTP_SENSOR_GETVALUE_API"))
	logger.Info("HTTP_SENSOR_REGISTER_API = " + os.Getenv("HTTP_SENSOR_REGISTER_API"))
	logger.Info("HTTP_SENSOR_KEEPALIVE_API = " + os.Getenv("HTTP_SENSOR_KEEPALIVE_API"))
	logger.Info("HTTP_ONLINE_SERVER = " + os.Getenv("HTTP_ONLINE_SERVER"))
	logger.Info("HTTP_ONLINE_PORT = " + os.Getenv("HTTP_ONLINE_PORT"))
	logger.Info("HTTP_ONLINE_API = " + os.Getenv("HTTP_ONLINE_API"))
	logger.Info("HTTP_ONLINE_FCMTOKEN_API = " + os.Getenv("HTTP_ONLINE_FCMTOKEN_API"))
	logger.Info("HTTP_ONLINE_KEEPALIVE_API = " + os.Getenv("HTTP_ONLINE_KEEPALIVE_API"))
	logger.Info("GRPC_URL = " + os.Getenv("GRPC_URL"))
	logger.Info("GRPC_TLS = " + os.Getenv("GRPC_TLS"))
	logger.Info("CERT_FOLDER_PATH = " + os.Getenv("CERT_FOLDER_PATH"))
	logger.Info("SINGLE_USER_LOGIN_EMAIL = " + os.Getenv("SINGLE_USER_LOGIN_EMAIL"))
	logger.Info("JWT_PASSWORD = " + os.Getenv("JWT_PASSWORD"))
	logger.Info("COOKIE_SECRET = " + os.Getenv("COOKIE_SECRET"))
	logger.Info("INTERNAL_CLUSTER_PATH = " + os.Getenv("INTERNAL_CLUSTER_PATH"))
}
