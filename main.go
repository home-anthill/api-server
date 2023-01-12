package main

import (
	"api-server/init_config"
	"os"
)

func main() {
	// 1. Init config
	logger := init_config.BuildConfig()
	defer logger.Sync()

	// 2. Init server
	port := os.Getenv("HTTP_PORT")
	httpOrigin := os.Getenv("HTTP_SERVER") + ":" + port
	router, _, _, _, _ := init_config.BuildServer(httpOrigin, logger)

	// 3. Start server
	var err error
	logger.Info("GIN - up and running with port: " + port)
	if os.Getenv("HTTP_TLS") == "true" {
		logger.Info("TLS enabled, running HTTPS server...")
		err = router.RunTLS(
			":"+port,
			os.Getenv("HTTP_CERT_FILE"),
			os.Getenv("HTTP_KEY_FILE"),
		)
	} else {
		err = router.Run(":" + port)
	}
	if err != nil {
		logger.Error("Cannot start HTTP server", err)
		panic(err)
	}
}
