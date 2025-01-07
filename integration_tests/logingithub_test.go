package integration_tests

import (
	"api-server/initialization"
	"encoding/base64"
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
		logger, router, _, _ = initialization.Start()
		defer logger.Sync()
	})

	Context("calling login api", func() {
		It("should return login oauth URL", func() {
			recorder := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/login", nil)
			router.ServeHTTP(recorder, req)
			Expect(http.StatusFound).To(Equal(recorder.Code))

			resLocation, _ := recorder.Result().Location()
			callbackURL := url.QueryEscape(os.Getenv("OAUTH2_CALLBACK"))
			suffix := `client_id=` +
				os.Getenv("OAUTH2_CLIENTID") +
				`&redirect_uri=` + callbackURL + `&response_type=code&scope=repo&state=`

			stateEscapedB64 := strings.ReplaceAll(resLocation.RawQuery, suffix, "")
			stateUnescapedB64, err := url.PathUnescape(stateEscapedB64)
			Expect(err).ShouldNot(HaveOccurred())
			decodeBytes, err := base64.StdEncoding.DecodeString(stateUnescapedB64)
			Expect(err).ShouldNot(HaveOccurred())
			decoded := string(decodeBytes)
			// state is generate via session.RandToken as a 32 byte string, so the format of that value must be the same
			Expect([]byte(decoded)).To(HaveLen(32))
		})

		It("should return login oauth mobile app URL", func() {
			recorder := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/login_app", nil)
			router.ServeHTTP(recorder, req)
			Expect(http.StatusFound).To(Equal(recorder.Code))

			resLocation, _ := recorder.Result().Location()
			callbackURL := url.QueryEscape(os.Getenv("OAUTH2_APP_CALLBACK"))
			suffix := `client_id=` +
				os.Getenv("OAUTH2_APP_CLIENTID") +
				`&redirect_uri=` + callbackURL + `&response_type=code&scope=repo&state=`

			stateEscapedB64 := strings.ReplaceAll(resLocation.RawQuery, suffix, "")
			stateUnescapedB64, err := url.PathUnescape(stateEscapedB64)
			Expect(err).ShouldNot(HaveOccurred())
			decodeBytes, err := base64.StdEncoding.DecodeString(stateUnescapedB64)
			Expect(err).ShouldNot(HaveOccurred())
			decoded := string(decodeBytes)
			// state is generate via session.RandToken as a 32 byte string, so the format of that value must be the same
			Expect([]byte(decoded)).To(HaveLen(32))
		})
	})
})
