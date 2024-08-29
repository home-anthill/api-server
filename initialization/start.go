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

// Start function
func Start() (*zap.SugaredLogger, *gin.Engine, context.Context, *mongo.Client) {
	// 1. Init logger
	logger := InitLogger()
	defer logger.Sync()

	// 2. Init env
	InitEnv(logger)

	// 3. Init db
	ctx := context.Background()
	// Connect to DB
	client := db.InitDb(ctx, logger)

	// 4. Init server
	port := os.Getenv("HTTP_PORT")
	httpServer := os.Getenv("HTTP_SERVER")
	oauthCallback := os.Getenv("OAUTH_CALLBACK")
	router, ctx := BuildServer(ctx, httpServer, port, oauthCallback, logger, client)

	return logger, router, ctx, client
}

// BuildServer - Exposed only for testing purposes
func BuildServer(ctx context.Context, httpServer string, port string, oauthCallback string, logger *zap.SugaredLogger, client *mongo.Client) (*gin.Engine, context.Context) {
	// Create a singleton validator instance. Validate is designed to be used as a singleton instance.
	// It caches information about struct and validations.
	validate := validator.New()

	// Config Gin framework mode based on env
	setGinMode()

	// Instantiate GIN and apply some middlewares
	logger.Info("BuildServer - GIN - Initializing...")
	router, cookieStore := SetupRouter(httpServer, port, oauthCallback, logger)
	RegisterRoutes(ctx, router, &cookieStore, logger, validate, client)
	return router, ctx
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
