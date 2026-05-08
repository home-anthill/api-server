package initialization

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

const projectDirName = "api-server"

// InitEnv loads the .env file and validates the minimum required environment.
func InitEnv(logger *zap.SugaredLogger) error {
	// Load .env file and print variables
	envFile, err := readEnv()
	logger.Debugf("BuildConfig - envFile = %s", envFile)
	if err != nil {
		logger.Error("BuildConfig - failed to load the env file")
		return fmt.Errorf("load env file at %s: %w", envFile, err)
	}
	return printEnv(logger)
}

func readEnv() (string, error) {
	// solution taken from https://stackoverflow.com/a/68347834/3590376
	projectName := regexp.MustCompile(`^(.*` + projectDirName + `)`)
	currentWorkDirectory, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("cannot get current working directory: %w", err)
	}
	rootPath := projectName.Find([]byte(currentWorkDirectory))
	envFilePath := filepath.Join(string(rootPath), ".env")
	err = godotenv.Load(envFilePath)
	return envFilePath, err
}

func printEnv(logger *zap.SugaredLogger) error {
	required := []string{
		"JWT_PASSWORD",
		"JWT_REFRESH_PASSWORD",
		"REFRESH_TOKEN_HASH_SECRET",
		"COOKIE_SECRET",
		"OAUTH2_CLIENTID",
		"OAUTH2_SECRETID",
		"OAUTH2_APP_CLIENTID",
		"OAUTH2_APP_SECRETID",
		"OAUTH2_CALLBACK",
		"OAUTH2_APP_CALLBACK",
	}
	for _, name := range required {
		if strings.TrimSpace(os.Getenv(name)) == "" {
			return fmt.Errorf("'%s' environment variable is mandatory", name)
		}
	}
	if len(os.Getenv("JWT_PASSWORD")) < 32 {
		return errors.New("'JWT_PASSWORD' environment variable must be at least 32 characters")
	}
	if len(os.Getenv("JWT_REFRESH_PASSWORD")) < 32 {
		return errors.New("'JWT_REFRESH_PASSWORD' environment variable must be at least 32 characters")
	}
	if len(os.Getenv("REFRESH_TOKEN_HASH_SECRET")) < 32 {
		return errors.New("'REFRESH_TOKEN_HASH_SECRET' environment variable must be at least 32 characters")
	}
	if len(os.Getenv("COOKIE_SECRET")) < 32 {
		return errors.New("'COOKIE_SECRET' environment variable is mandatory and must be at least 32 characters")
	}
	validateOAuthURL := func(name string) error {
		raw := strings.TrimSpace(os.Getenv(name))
		parsed, err := url.Parse(raw)
		if err != nil {
			return fmt.Errorf("'%s' environment variable must be a valid URL: %w", name, err)
		}
		if parsed.Scheme == "" || parsed.Host == "" {
			return fmt.Errorf("'%s' environment variable must include scheme and host", name)
		}
		return nil
	}
	if err := validateOAuthURL("OAUTH2_CALLBACK"); err != nil {
		return err
	}
	if err := validateOAuthURL("OAUTH2_APP_CALLBACK"); err != nil {
		return err
	}
	if os.Getenv("ENV") == "prod" && os.Getenv("HTTP_CORS") == "true" {
		return errors.New("'HTTP_CORS' must be false in production")
	}

	logger.Infof("ENVIRONMENT = %s", os.Getenv("ENV"))
	logger.Infof("LOG_FOLDER = %s", os.Getenv("LOG_FOLDER"))
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
	logger.Infof("HTTP_ONLINE_ROTATE_APITOKEN_API = %s", os.Getenv("HTTP_ONLINE_ROTATE_APITOKEN_API"))
	logger.Infof("HTTP_ONLINE_KEEPALIVE_API = %s", os.Getenv("HTTP_ONLINE_KEEPALIVE_API"))
	logger.Infof("GRPC_URL = %s", os.Getenv("GRPC_URL"))
	logger.Infof("GRPC_TLS = %s", os.Getenv("GRPC_TLS"))
	logger.Infof("CERT_FOLDER_PATH = %s", os.Getenv("CERT_FOLDER_PATH"))
	logger.Infof("LIMIT_TO_USER_EMAILS = %s", os.Getenv("LIMIT_TO_USER_EMAILS"))
	logger.Infof("INTERNAL_CLUSTER_PATH = %s", os.Getenv("INTERNAL_CLUSTER_PATH"))
	logger.Infof("MONGODB_URL = [redacted]")
	logger.Infof("COOKIE_SECRET = [redacted]")
	logger.Infof("JWT_PASSWORD = [redacted]")
	logger.Infof("JWT_REFRESH_PASSWORD = [redacted]")
	logger.Infof("OAUTH2_SECRETID = [redacted]")
	logger.Infof("OAUTH2_APP_SECRETID = [redacted]")
	return nil
}
