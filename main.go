package main

import (
	"api-server/initialization"
	"os"
)

func main() {
	logger, router, _, _, _, _ := initialization.Start()

	// Start server
	var err error
	port := os.Getenv("HTTP_PORT")
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
