package initialization

import (
	"api-server/db"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"os"
)

func Start() (*zap.SugaredLogger, *gin.Engine, context.Context, *mongo.Collection, *mongo.Collection, *mongo.Collection) {
	// 1. Init logger
	logger := InitLogger()
	defer logger.Sync()

	// 2. Init env
	InitEnv(logger)

	// 3. Init server
	port := os.Getenv("HTTP_PORT")
	httpOrigin := os.Getenv("HTTP_SERVER") + ":" + port
	router, ctx, collectionProfiles, collectionHomes, collectionDevices := BuildServer(httpOrigin, logger)

	return logger, router, ctx, collectionProfiles, collectionHomes, collectionDevices
}

// BuildServer - Exposed only for testing purposes
func BuildServer(httpOrigin string, logger *zap.SugaredLogger) (*gin.Engine, context.Context, *mongo.Collection, *mongo.Collection, *mongo.Collection) {
	// Initialization
	ctx := context.Background()
	// Create a singleton validator instance. Validate is designed to be used as a singleton instance.
	// It caches information about struct and validations.
	validate := validator.New()

	// Config Gin framework mode based on env
	setGinMode()

	// Connect to DB
	collectionProfiles, collectionHomes, collectionDevices := db.InitDb(ctx, logger)

	// Instantiate GIN and apply some middlewares
	logger.Info("BuildServer - GIN - Initializing...")
	router := SetupRouter(httpOrigin, logger)
	RegisterRoutes(router, ctx, logger, validate, collectionProfiles, collectionHomes, collectionDevices)
	return router, ctx, collectionProfiles, collectionHomes, collectionDevices
}

func setGinMode() {
	switch os.Getenv("ENV") {
	case "prod":
		gin.SetMode(gin.ReleaseMode)
		break
	case "testing":
		gin.SetMode(gin.TestMode)
		break
	default:
		gin.SetMode(gin.DebugMode)
	}
}
