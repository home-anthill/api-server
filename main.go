package main

import (
	"api-server/initialization"
	"context"
	"os"
)

func main() {
	logger, router, _, client := initialization.Start()
	defer client.Disconnect(context.TODO())

	// Start server
	port := os.Getenv("HTTP_PORT")
	logger.Info("GIN - up and running with port: " + port)
	err := router.Run(":" + port)
	if err != nil {
		logger.Error("Cannot start HTTP server", err)
		panic(err)
	}
}
