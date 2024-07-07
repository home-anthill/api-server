package main

import (
	"api-server/initialization"
	"os"
)

func main() {
	logger, router, _, _, _, _ := initialization.Start()

	// Start server
	port := os.Getenv("HTTP_PORT")
	logger.Info("GIN - up and running with port: " + port)
	err := router.Run(":" + port)
	if err != nil {
		logger.Error("Cannot start HTTP server", err)
		panic(err)
	}
}
