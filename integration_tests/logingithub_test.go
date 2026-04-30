package integration_tests

import (
	"api-server/initialization"
	"api-server/utils"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
)

var _ = Describe("LoginGithub", func() {
	var logger *zap.SugaredLogger
	var router *gin.Engine

	BeforeEach(func() {
		logger, router, _ = initialization.MustStart()
		defer logger.Sync()
	})

	Context("calling login api", func() {
		It("should return login oauth URL", func() {
			recorder := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/oauth/login", nil)
			router.ServeHTTP(recorder, req)
			Expect(http.StatusTemporaryRedirect).To(Equal(recorder.Code))

			resLocation, err := recorder.Result().Location()
			Expect(err).ShouldNot(HaveOccurred())
			query := resLocation.Query()
			Expect(query.Get("client_id")).To(Equal(os.Getenv("OAUTH2_CLIENTID")))
			Expect(query.Get("redirect_uri")).To(Equal(os.Getenv("OAUTH2_CALLBACK")))
			Expect(query.Get("response_type")).To(Equal("code"))
			Expect(query.Get("scope")).To(Equal("read:user user:email"))
			Expect(query.Get("code_challenge_method")).To(Equal(utils.PKCEChallengeMethodS256))
			Expect(utils.IsValidPKCECodeChallenge(query.Get("code_challenge"))).To(BeTrue())

			state := query.Get("state")
			Expect(state).To(HaveLen(128))
			decodeBytes, err := base64.RawURLEncoding.DecodeString(state)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(decodeBytes).To(HaveLen(96))
		})

		It("should return login oauth mobile app URL", func() {
			codeVerifier, err := utils.NewPKCEVerifier()
			Expect(err).ShouldNot(HaveOccurred())
			codeChallenge, err := utils.BuildPKCECodeChallenge(codeVerifier)
			Expect(err).ShouldNot(HaveOccurred())
			appState, err := utils.NewPKCEVerifier()
			Expect(err).ShouldNot(HaveOccurred())

			recorder := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/oauth/app/login?code_challenge="+url.QueryEscape(codeChallenge)+"&code_challenge_method="+utils.PKCEChallengeMethodS256+"&app_state="+url.QueryEscape(appState), nil)
			router.ServeHTTP(recorder, req)
			Expect(http.StatusTemporaryRedirect).To(Equal(recorder.Code))

			resLocation, err := recorder.Result().Location()
			Expect(err).ShouldNot(HaveOccurred())
			query := resLocation.Query()
			Expect(query.Get("client_id")).To(Equal(os.Getenv("OAUTH2_APP_CLIENTID")))
			Expect(query.Get("redirect_uri")).To(Equal(os.Getenv("OAUTH2_APP_CALLBACK")))
			Expect(query.Get("response_type")).To(Equal("code"))
			Expect(query.Get("scope")).To(Equal("read:user user:email"))
			Expect(query.Get("code_challenge_method")).To(Equal(utils.PKCEChallengeMethodS256))
			Expect(utils.IsValidPKCECodeChallenge(query.Get("code_challenge"))).To(BeTrue())

			state := query.Get("state")
			Expect(state).To(HaveLen(128))
			decodeBytes, err := base64.RawURLEncoding.DecodeString(state)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(decodeBytes).To(HaveLen(96))
		})

		It("should reject mobile app login without PKCE parameters", func() {
			recorder := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/oauth/app/login", nil)
			router.ServeHTTP(recorder, req)
			Expect(http.StatusBadRequest).To(Equal(recorder.Code))
			Expect(recorder.Body.String()).To(Equal(`{"error":"missing or invalid PKCE parameters"}`))
		})

		It("should reject mobile app login without app state", func() {
			codeVerifier, err := utils.NewPKCEVerifier()
			Expect(err).ShouldNot(HaveOccurred())
			codeChallenge, err := utils.BuildPKCECodeChallenge(codeVerifier)
			Expect(err).ShouldNot(HaveOccurred())

			recorder := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/oauth/app/login?code_challenge="+url.QueryEscape(codeChallenge)+"&code_challenge_method="+utils.PKCEChallengeMethodS256, nil)
			router.ServeHTTP(recorder, req)
			Expect(http.StatusBadRequest).To(Equal(recorder.Code))
			Expect(recorder.Body.String()).To(Equal(`{"error":"missing or invalid app state"}`))
		})

		It("should complete web oauth callback and redirect to postlogin with a token", func() {
			recorder := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/oauth/login", nil)
			router.ServeHTTP(recorder, req)
			Expect(http.StatusTemporaryRedirect).To(Equal(recorder.Code))

			resLocation, err := recorder.Result().Location()
			Expect(err).ShouldNot(HaveOccurred())
			state := resLocation.Query().Get("state")
			Expect(state).ToNot(BeEmpty())
			cookieSession := recorder.Header().Get("Set-Cookie")
			Expect(cookieSession).ToNot(BeEmpty())

			recorder = httptest.NewRecorder()
			req = httptest.NewRequest("GET", "/api/oauth/callback?code=ecf9b407c22a56b61ecf&state="+url.QueryEscape(state), nil)
			req.Header.Add("Cookie", cookieSession)
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusFound))
			Expect(recorder.Header().Get("Location")).To(HavePrefix("/postlogin#token="))
			Expect(strings.TrimPrefix(recorder.Header().Get("Location"), "/postlogin#token=")).ToNot(BeEmpty())
		})
	})
})
