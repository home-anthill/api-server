package integration_tests

import (
	"api-server/db"
	"api-server/initialization"
	"api-server/models"
	"api-server/testuutils"
	"api-server/utils"
	"context"
	"crypto/sha256"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"
)

var _ = Describe("LoginGithub", func() {
	var ctx context.Context
	var logger *zap.SugaredLogger
	var router *gin.Engine
	var client *mongo.Client
	var collProfiles *mongo.Collection
	var collHomes *mongo.Collection
	var collDevices *mongo.Collection
	var collAppLoginCodes *mongo.Collection
	var collRefreshTokens *mongo.Collection

	BeforeEach(func() {
		logger, router, client = initialization.MustStart()
		ctx = context.Background()
		defer logger.Sync()

		colls := db.GetCollections(client)
		collProfiles = colls.Profiles
		collHomes = colls.Homes
		collDevices = colls.Devices
		collAppLoginCodes = colls.AppLoginCodes
		collRefreshTokens = colls.RefreshTokens

		err := os.Setenv("SINGLE_USER_LOGIN_EMAIL", "test@test.com")
		Expect(err).ShouldNot(HaveOccurred())
	})

	AfterEach(func() {
		testuutils.DropAllCollections(ctx, collProfiles, collHomes, collDevices, collAppLoginCodes, collRefreshTokens)
	})

	Context("calling protected api", func() {
		It("should return an error if Authorization header is an empty string", func() {
			jwtToken, cookieSession := testuutils.GetJwt(router)
			profileRes := testuutils.GetLoggedProfile(router, jwtToken, cookieSession)
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/profiles/"+profileRes.ID.Hex()+"/tokens", nil)
			req.Header.Add("Cookie", cookieSession)
			req.Header.Add("Authorization", "")
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusUnauthorized))
			Expect(recorder.Body.String()).To(Equal(`{"error":"authorization header not found"}`))
		})

		It("should return an error if Bearer header is an empty string", func() {
			jwtToken, cookieSession := testuutils.GetJwt(router)
			profileRes := testuutils.GetLoggedProfile(router, jwtToken, cookieSession)
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/profiles/"+profileRes.ID.Hex()+"/tokens", nil)
			req.Header.Add("Cookie", cookieSession)
			req.Header.Add("Authorization", "Bearer ")
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusUnauthorized))
			Expect(recorder.Body.String()).To(Equal(`{"error":"bearer token not found"}`))
		})

		It("should return an error if token is not valid", func() {
			jwtToken, cookieSession := testuutils.GetJwt(router)
			profileRes := testuutils.GetLoggedProfile(router, jwtToken, cookieSession)
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/profiles/"+profileRes.ID.Hex()+"/tokens", nil)
			req.Header.Add("Cookie", cookieSession)
			req.Header.Add("Authorization", "Bearer bad.jwt.token")
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusBadRequest))
			Expect(recorder.Body.String()).To(Equal(`{"error":"that's not even a token"}`))
		})

		It("should return an error if token is expired", func() {
			jwtToken, cookieSession := testuutils.GetJwt(router)
			profileRes := testuutils.GetLoggedProfile(router, jwtToken, cookieSession)

			// create an expired JWT
			expirationTime := time.Now().Add(-60 * time.Minute)
			tokenString, err := utils.CreateJWT(profileRes, expirationTime, utils.AccessToken, []byte(os.Getenv("JWT_PASSWORD")))
			Expect(err).ShouldNot(HaveOccurred())
			logger.Infof("tokenString = %s", tokenString)

			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/profiles/"+profileRes.ID.Hex()+"/tokens", nil)
			req.Header.Add("Cookie", cookieSession)
			req.Header.Add("Authorization", "Bearer "+tokenString)
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusUnauthorized))
			Expect(recorder.Body.String()).To(Equal(`{"error":"token is expired"}`))
		})

		It("should reject a refresh token used as an access token", func() {
			jwtToken, cookieSession := testuutils.GetJwt(router)
			profileRes := testuutils.GetLoggedProfile(router, jwtToken, cookieSession)

			expirationTime := time.Now().Add(60 * time.Minute)
			refreshTokenString, err := utils.CreateJWT(profileRes, expirationTime, utils.RefreshToken, []byte(os.Getenv("JWT_PASSWORD")))
			Expect(err).ShouldNot(HaveOccurred())

			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/api/profile", nil)
			req.Header.Add("Cookie", cookieSession)
			req.Header.Add("Authorization", "Bearer "+refreshTokenString)
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusUnauthorized))
			Expect(recorder.Body.String()).To(Equal(`{"error":"token is not an access token"}`))
		})

		It("should reject requests when JWT and session belong to different users", func() {
			jwtToken, cookieSession := testuutils.GetJwt(router)
			profileRes := testuutils.GetLoggedProfile(router, jwtToken, cookieSession)

			otherProfile := models.Profile{
				ID: bson.NewObjectID(),
				Github: models.GitHub{
					ID:        profileRes.Github.ID + 1,
					Login:     "other-user",
					Name:      "Other User",
					Email:     "other@test.com",
					AvatarURL: "https://example.com/avatar.png",
				},
				CreatedAt:  time.Now(),
				ModifiedAt: time.Now(),
			}

			blockKey := sha256.Sum256([]byte(os.Getenv("COOKIE_SECRET")))
			store := cookie.NewStore([]byte(os.Getenv("COOKIE_SECRET")), blockKey[:])
			store.Options(sessions.Options{
				Path:     "/",
				HttpOnly: true,
				Secure:   os.Getenv("ENV") == "prod",
				SameSite: http.SameSiteLaxMode,
			})
			mismatchRouter := gin.New()
			mismatchRouter.Use(sessions.Sessions(utils.SessionName, store))
			mismatchRouter.GET("/set", func(c *gin.Context) {
				session := sessions.Default(c)
				session.Set("profileID", otherProfile.ID.Hex())
				session.Set("githubID", otherProfile.Github.ID)
				err := session.Save()
				Expect(err).ShouldNot(HaveOccurred())
				c.Status(http.StatusNoContent)
			})

			mismatchRecorder := httptest.NewRecorder()
			mismatchReq := httptest.NewRequest("GET", "/set", nil)
			mismatchRouter.ServeHTTP(mismatchRecorder, mismatchReq)
			Expect(mismatchRecorder.Code).To(Equal(http.StatusNoContent))
			mismatchCookie := mismatchRecorder.Header().Get("Set-Cookie")
			Expect(mismatchCookie).ToNot(BeEmpty())
			Expect(mismatchCookie).ToNot(Equal(cookieSession))

			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/profiles/"+profileRes.ID.Hex()+"/tokens", nil)
			req.Header.Add("Cookie", mismatchCookie)
			req.Header.Add("Authorization", "Bearer "+jwtToken)
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusUnauthorized))
			Expect(recorder.Body.String()).To(Equal(`{"error":"session does not match token identity"}`))
		})
	})

	Context("calling refresh token api", func() {
		It("should return a new access token with a valid refresh token cookie", func() {
			_, cookieSession := testuutils.GetJwt(router)

			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/oauth/refresh", nil)
			req.Header.Add("Cookie", cookieSession)
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusOK))

			var body map[string]string
			err := json.Unmarshal(recorder.Body.Bytes(), &body)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(body["token"]).ShouldNot(BeEmpty())

			// refresh token cookie must also be rotated
			Expect(strings.Join(recorder.Header().Values("Set-Cookie"), "; ")).To(ContainSubstring(utils.RefreshTokenCookieName + "="))
		})

		It("should return an error if refresh token cookie is missing", func() {
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/oauth/refresh", nil)
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusUnauthorized))
			Expect(recorder.Body.String()).To(Equal(`{"error":"refresh token not found"}`))
		})

		It("should return an error if refresh token is expired", func() {
			jwtToken, cookieSession := testuutils.GetJwt(router)
			profileRes := testuutils.GetLoggedProfile(router, jwtToken, cookieSession)

			expiredRefreshToken, err := utils.RandomString(64)
			Expect(err).ShouldNot(HaveOccurred())
			now := time.Now().UTC()
			_, err = collRefreshTokens.InsertOne(ctx, models.RefreshToken{
				ID:         bson.NewObjectID(),
				ProfileID:  profileRes.ID,
				TokenHash:  utils.HashToken(expiredRefreshToken),
				FamilyID:   bson.NewObjectID().Hex(),
				ClientType: "web",
				CreatedAt:  now.Add(-2 * time.Hour),
				ExpiresAt:  now.Add(-1 * time.Hour),
			})
			Expect(err).ShouldNot(HaveOccurred())

			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/oauth/refresh", nil)
			req.AddCookie(&http.Cookie{Name: utils.RefreshTokenCookieName, Value: expiredRefreshToken})
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusUnauthorized))
			Expect(recorder.Body.String()).To(Equal(`{"error":"refresh token expired"}`))
		})

		It("should return an error if an access token is used as refresh token", func() {
			jwtToken, cookieSession := testuutils.GetJwt(router)
			profileRes := testuutils.GetLoggedProfile(router, jwtToken, cookieSession)

			expirationTime := time.Now().Add(60 * time.Minute)
			accessToken, err := utils.CreateJWT(profileRes, expirationTime, utils.AccessToken, []byte(os.Getenv("JWT_REFRESH_PASSWORD")))
			Expect(err).ShouldNot(HaveOccurred())

			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/oauth/refresh", nil)
			req.AddCookie(&http.Cookie{Name: utils.RefreshTokenCookieName, Value: accessToken})
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusUnauthorized))
			Expect(recorder.Body.String()).To(Equal(`{"error":"invalid refresh token"}`))
		})

		It("should return rotated mobile tokens and renew the session cookie", func() {
			_, refreshToken := testuutils.GetJwtMobileApp(router)

			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/oauth/app/refresh", strings.NewReader(`{"refreshToken":"`+refreshToken+`"}`))
			req.Header.Add("Content-Type", "application/json")
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusOK))

			var body map[string]string
			err := json.Unmarshal(recorder.Body.Bytes(), &body)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(body["token"]).ShouldNot(BeEmpty())
			Expect(body["refreshToken"]).ShouldNot(BeEmpty())

			sessionCookie := recorder.Header().Get("Set-Cookie")
			Expect(sessionCookie).To(ContainSubstring(utils.SessionName + "="))
			testuutils.GetLoggedProfile(router, body["token"], sessionCookie)
		})
	})

	Context("calling app code exchange api", func() {
		It("should exchange a valid app code exactly once", func() {
			codeVerifier, err := utils.RandomString(32)
			Expect(err).ShouldNot(HaveOccurred())
			codeChallenge, err := utils.BuildPKCECodeChallenge(codeVerifier)
			Expect(err).ShouldNot(HaveOccurred())

			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/api/oauth/app/login?code_challenge="+url.QueryEscape(codeChallenge)+"&code_challenge_method="+utils.PKCEChallengeMethodS256, nil)
			router.ServeHTTP(recorder, req)
			Expect(http.StatusTemporaryRedirect).To(Equal(recorder.Code))

			resLocation, err := recorder.Result().Location()
			Expect(err).ShouldNot(HaveOccurred())
			stateEscapedB64 := resLocation.Query().Get("state")
			cookieSession := recorder.Header().Get("Set-Cookie")

			recorder = httptest.NewRecorder()
			req = httptest.NewRequest("GET", "/api/oauth/app/callback?code=ecf9b407c22a56b61ecf&state="+stateEscapedB64, nil)
			req.Header.Add("Cookie", cookieSession)
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusFound))

			locationHeader := recorder.Header().Get("location")
			callbackURL, err := url.Parse(os.Getenv("OAUTH2_APP_CALLBACK"))
			Expect(err).ShouldNot(HaveOccurred())
			Expect(locationHeader).To(HavePrefix(callbackURL.Scheme + "://" + callbackURL.Host + "/postlogin?code="))
			redirectLocation, err := url.Parse(locationHeader)
			Expect(err).ShouldNot(HaveOccurred())
			code := redirectLocation.Query().Get("code")
			Expect(code).ToNot(BeEmpty())

			recorder = httptest.NewRecorder()
			req = httptest.NewRequest("POST", "/api/oauth/app/exchange-code", strings.NewReader(`{"code":"`+code+`","codeVerifier":"`+codeVerifier+`"}`))
			req.Header.Add("Content-Type", "application/json")
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusOK))

			var body map[string]string
			err = json.Unmarshal(recorder.Body.Bytes(), &body)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(body["token"]).ShouldNot(BeEmpty())
			Expect(body["refreshToken"]).ShouldNot(BeEmpty())

			recorder = httptest.NewRecorder()
			req = httptest.NewRequest("POST", "/api/oauth/app/exchange-code", strings.NewReader(`{"code":"`+code+`","codeVerifier":"`+codeVerifier+`"}`))
			req.Header.Add("Content-Type", "application/json")
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusUnauthorized))
			Expect(recorder.Body.String()).To(Equal(`{"error":"invalid or expired code"}`))
		})

		It("should reject app code exchange without a valid PKCE verifier", func() {
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/oauth/app/exchange-code", strings.NewReader(`{"code":"abc"}`))
			req.Header.Add("Content-Type", "application/json")
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusBadRequest))
			Expect(recorder.Body.String()).To(Equal(`{"error":"invalid request payload"}`))
		})
	})
})
