package integration_tests

import (
	"api-server/init_config"
	"api-server/models"
	"encoding/json"
	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"os"
)

var _ = Describe("KeepAlive", func() {
	var logger *zap.SugaredLogger
	var router *gin.Engine

	BeforeEach(func() {
		// 1. Init config
		logger = init_config.BuildConfig()
		defer logger.Sync()

		// 2. Init server
		port := os.Getenv("HTTP_PORT")
		httpOrigin := os.Getenv("HTTP_SERVER") + ":" + port
		router, _, _, _, _ = init_config.BuildServer(httpOrigin, logger)
	})

	Context("calling keepalive api", func() {
		It("should return 'ok'", func() {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/keepalive", nil)
			router.ServeHTTP(w, req)
			Expect(http.StatusOK).To(Equal(w.Code))

			response := models.KeepAlive{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(response.Message).To(Equal("ok"))
		})
	})
})
