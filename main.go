package main

import (
	"api-server/initialization"
	"context"
	"os"
)

func main() {
	logger, router, mongoDbClient := initialization.Start()
	defer logger.Sync()
	defer mongoDbClient.Disconnect(context.TODO())

	// Start server
	port := os.Getenv("HTTP_PORT")
	logger.Infof("GIN - up and running with port: %s", port)
	err := router.Run(":" + port)
	if err != nil {
		logger.Error("Cannot start HTTP server", err)
		os.Exit(1)
	}
}
