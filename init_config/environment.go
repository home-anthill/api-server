package init_config

import (
	"fmt"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"os"
	"regexp"
)

const projectDirName = "api-server"

func InitEnv() (string, error) {
	// solution taken from https://stackoverflow.com/a/68347834/3590376
	projectName := regexp.MustCompile(`^(.*` + projectDirName + `)`)
	currentWorkDirectory, _ := os.Getwd()
	rootPath := projectName.Find([]byte(currentWorkDirectory))
	envFilePath := string(rootPath) + `/.env`
	err := godotenv.Load(envFilePath)
	return envFilePath, err
}

func PrintEnv(logger *zap.SugaredLogger) {
	if os.Getenv("JWT_PASSWORD") == "" {
		panic(fmt.Errorf("'JWT_PASSWORD' environment variable is mandatory"))
	}

	fmt.Println("ENVIRONMENT = " + os.Getenv("ENV"))
	fmt.Println("MONGODB_URL = " + os.Getenv("MONGODB_URL"))
	fmt.Println("HTTP_SERVER = " + os.Getenv("HTTP_SERVER"))
	fmt.Println("HTTP_PORT = " + os.Getenv("HTTP_PORT"))
	fmt.Println("HTTP_TLS = " + os.Getenv("HTTP_TLS"))
	fmt.Println("HTTP_CERT_FILE = " + os.Getenv("HTTP_CERT_FILE"))
	fmt.Println("HTTP_KEY_FILE = " + os.Getenv("HTTP_KEY_FILE"))
	fmt.Println("HTTP_CORS = " + os.Getenv("HTTP_CORS"))
	fmt.Println("HTTP_SENSOR_SERVER = " + os.Getenv("HTTP_SENSOR_SERVER"))
	fmt.Println("HTTP_SENSOR_PORT = " + os.Getenv("HTTP_SENSOR_PORT"))
	fmt.Println("HTTP_SENSOR_GETVALUE_API = " + os.Getenv("HTTP_SENSOR_GETVALUE_API"))
	fmt.Println("HTTP_SENSOR_REGISTER_API = " + os.Getenv("HTTP_SENSOR_REGISTER_API"))
	fmt.Println("HTTP_SENSOR_KEEPALIVE_API = " + os.Getenv("HTTP_SENSOR_KEEPALIVE_API"))
	fmt.Println("GRPC_URL = " + os.Getenv("GRPC_URL"))
	fmt.Println("GRPC_TLS = " + os.Getenv("GRPC_TLS"))
	fmt.Println("CERT_FOLDER_PATH = " + os.Getenv("CERT_FOLDER_PATH"))
	fmt.Println("SINGLE_USER_LOGIN_EMAIL = " + os.Getenv("SINGLE_USER_LOGIN_EMAIL"))
	fmt.Println("JWT_PASSWORD = " + os.Getenv("JWT_PASSWORD"))
	fmt.Println("COOKIE_SECRET = " + os.Getenv("COOKIE_SECRET"))
	fmt.Println("OAUTH2_CLIENTID = " + os.Getenv("OAUTH2_CLIENTID"))
	fmt.Println("OAUTH2_SECRETID = " + os.Getenv("OAUTH2_SECRETID"))
	fmt.Println("INTERNAL_CLUSTER_PATH = " + os.Getenv("INTERNAL_CLUSTER_PATH"))

	logger.Info("ENVIRONMENT = " + os.Getenv("ENV"))
	logger.Info("MONGODB_URL = " + os.Getenv("MONGODB_URL"))
	logger.Info("HTTP_SERVER = " + os.Getenv("HTTP_SERVER"))
	logger.Info("HTTP_PORT = " + os.Getenv("HTTP_PORT"))
	logger.Info("HTTP_TLS = " + os.Getenv("HTTP_TLS"))
	logger.Info("HTTP_CERT_FILE = " + os.Getenv("HTTP_CERT_FILE"))
	logger.Info("HTTP_KEY_FILE = " + os.Getenv("HTTP_KEY_FILE"))
	logger.Info("HTTP_CORS = " + os.Getenv("HTTP_CORS"))
	logger.Info("HTTP_SENSOR_SERVER = " + os.Getenv("HTTP_SENSOR_SERVER"))
	logger.Info("HTTP_SENSOR_PORT = " + os.Getenv("HTTP_SENSOR_PORT"))
	logger.Info("HTTP_SENSOR_GETVALUE_API = " + os.Getenv("HTTP_SENSOR_GETVALUE_API"))
	logger.Info("HTTP_SENSOR_REGISTER_API = " + os.Getenv("HTTP_SENSOR_REGISTER_API"))
	logger.Info("HTTP_SENSOR_KEEPALIVE_API = " + os.Getenv("HTTP_SENSOR_KEEPALIVE_API"))
	logger.Info("GRPC_URL = " + os.Getenv("GRPC_URL"))
	logger.Info("GRPC_TLS = " + os.Getenv("GRPC_TLS"))
	logger.Info("CERT_FOLDER_PATH = " + os.Getenv("CERT_FOLDER_PATH"))
	logger.Info("SINGLE_USER_LOGIN_EMAIL = " + os.Getenv("SINGLE_USER_LOGIN_EMAIL"))
	logger.Info("JWT_PASSWORD = " + os.Getenv("JWT_PASSWORD"))
	logger.Info("COOKIE_SECRET = " + os.Getenv("COOKIE_SECRET"))
	logger.Info("OAUTH2_CLIENTID = " + os.Getenv("OAUTH2_CLIENTID"))
	logger.Info("OAUTH2_SECRETID = " + os.Getenv("OAUTH2_SECRETID"))
	logger.Info("INTERNAL_CLUSTER_PATH = " + os.Getenv("INTERNAL_CLUSTER_PATH"))
}
