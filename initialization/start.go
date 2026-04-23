package initialization

import (
	"api-server/db"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"
)

// Start initializes logger, environment, database, and router.
func Start() (*zap.SugaredLogger, *gin.Engine, *mongo.Client, error) {
	// 1. Init logger
	logger := InitLogger()

	// 2. Init env
	if err := InitEnv(logger); err != nil {
		return logger, nil, nil, fmt.Errorf("init env: %w", err)
	}

	// 3. Init db
	// Connect to DB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	mongoDbClient, err := db.InitDb(ctx, logger)
	if err != nil {
		return logger, nil, nil, fmt.Errorf("init db: %w", err)
	}

	// 4. Init server
	router := BuildServer(logger, mongoDbClient)

	return logger, router, mongoDbClient, nil
}

// MustStart initializes the application and panics on error. It is intended for tests.
func MustStart() (*zap.SugaredLogger, *gin.Engine, *mongo.Client) {
	logger, router, mongoDbClient, err := Start()
	if err != nil {
		panic(err)
	}
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
