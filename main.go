package main

import (
	"api-server/initialization"
	"context"
	"os"
	"time"
)

func main() {
	logger, router, mongoDbClient, err := initialization.Start()
	if err != nil {
		if logger != nil {
			logger.Errorw("Cannot start application", "error", err)
		}
		os.Exit(1)
	}
	defer logger.Sync()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err = mongoDbClient.Disconnect(ctx); err != nil {
			logger.Warnw("Cannot disconnect MongoDB cleanly", "error", err)
		}
	}()

	// Start server
	port := os.Getenv("HTTP_PORT")
	logger.Infof("GIN - up and running with port: %s", port)
	if err = router.Run(":" + port); err != nil {
		logger.Error("Cannot start HTTP server", err)
		os.Exit(1)
	}
}
