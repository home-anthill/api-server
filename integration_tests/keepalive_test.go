package integration_tests

import (
	"api-server/initialization"
	"api-server/models"
	"encoding/json"
	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
)

var _ = Describe("KeepAlive", func() {
	var logger *zap.SugaredLogger
	var router *gin.Engine

	BeforeEach(func() {
		logger, router, _, _ = initialization.Start()
		defer logger.Sync()
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
