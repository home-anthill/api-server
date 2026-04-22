package initialization

import (
	"api-server/db"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"
)

// Start function
func Start() (*zap.SugaredLogger, *gin.Engine, *mongo.Client) {
	// 1. Init logger
	logger := InitLogger()

	// 2. Init env
	InitEnv(logger)

	// 3. Init db
	// Connect to DB
	mongoDbClient := db.InitDb(logger)

	// 4. Init server
	router := BuildServer(logger, mongoDbClient)

	return logger, router, mongoDbClient
}

// BuildServer - Exposed only for testing purposes
func BuildServer(logger *zap.SugaredLogger, client *mongo.Client) *gin.Engine {
	// Create a singleton validator instance. Validate is designed to be used as a singleton instance.
	// It caches information about struct and validations.
	validate := validator.New()

	// Config Gin framework mode based on env
	setGinMode()

	// Instantiate GIN and apply some middlewares
	logger.Info("BuildServer - GIN - Initializing...")
	router := SetupRouter(logger)
	RegisterRoutes(router, logger, validate, client)
	return router
}

func setGinMode() {
	if os.Getenv("ENV") == "prod" {
		gin.SetMode(gin.ReleaseMode)
	} else if os.Getenv("ENV") == "testing" {
		gin.SetMode(gin.TestMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}
}
