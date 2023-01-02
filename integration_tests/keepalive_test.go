package integration_tests

import (
	"api-server/init_config"
	"api-server/models"
	"encoding/json"
	"fmt"
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
		fmt.Println("BeforeEach called")
		// 1. Init config
		logger = init_config.BuildConfig()
		defer logger.Sync()

		// 2. Init server
		port := os.Getenv("HTTP_PORT")
		httpOrigin := os.Getenv("HTTP_SERVER") + ":" + port
		router = init_config.BuildServer(httpOrigin, logger)
	})

	Context("with more than 300 pages", func() {
		It("should be a novel", func() {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/keepalive", nil)
			router.ServeHTTP(w, req)
			Expect(200).To(Equal(w.Code))

			response := models.KeepAlive{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(response.Message).To(Equal("ok"))
		})
	})

	AfterEach(func() {
		fmt.Println("AfterEach called")
	})
})
