package integration_tests

import (
	"api-server/init_config"
	"api-server/models"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
)

var _ = Describe("LoginGithub", func() {
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

	Describe("Categorizing books", func() {
		Context("with more than 300 pages", func() {
			It("should be a novel", func() {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/api/login", nil)
				router.ServeHTTP(w, req)
				Expect(200).To(Equal(w.Code))

				response := models.LoginUrl{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				Expect(err).ShouldNot(HaveOccurred())
				fmt.Println("response: " + response.LoginURL)

				// oNe7MGolm0toqd7qkP9K6wnENdwXuNeDxqlLgbeIFwY%3D

				prefix := `https://github.com/login/oauth/authorize?client_id=181a9b0e2ed4bb00c42f&redirect_uri=http%3A%2F%2Flocalhost%3A8082%2Fapi%2Fcallback%2F&response_type=code&scope=repo&state=`

				stateEscapedB64 := strings.ReplaceAll(response.LoginURL, prefix, "")
				stateUnescapedB64, err := url.QueryUnescape(stateEscapedB64)
				Expect(err).ShouldNot(HaveOccurred())
				fmt.Println("test stateUnescapedB64 " + stateUnescapedB64)
				decodeBytes, err := base64.StdEncoding.DecodeString(stateUnescapedB64)
				Expect(err).ShouldNot(HaveOccurred())
				decoded := string(decodeBytes)
				fmt.Println("------------------decodeString " + decoded)
				Expect(response.LoginURL).To(HavePrefix(prefix))
				// state is generate via session.RandToken as a 32 byte string, so the format of that value must be the same
				Expect([]byte(decoded)).To(HaveLen(32))
			})
		})
	})

	AfterEach(func() {
		fmt.Println("AfterEach called")
	})
})
