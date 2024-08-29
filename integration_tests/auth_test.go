package integration_tests

import (
	"api-server/db"
	"api-server/initialization"
	"api-server/testuutils"
	"api-server/utils"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"net/http"
	"net/http/httptest"
	"os"
	"time"
)

var jwtKey = []byte(os.Getenv("JWT_PASSWORD"))

var _ = Describe("LoginGithub", func() {
	var ctx context.Context
	var logger *zap.SugaredLogger
	var router *gin.Engine
	var client *mongo.Client
	var collProfiles *mongo.Collection
	var collHomes *mongo.Collection
	var collDevices *mongo.Collection

	BeforeEach(func() {
		logger, router, ctx, client = initialization.Start()
		defer logger.Sync()

		collProfiles = db.GetCollections(client).Profiles
		collHomes = db.GetCollections(client).Homes
		collDevices = db.GetCollections(client).Devices

		err := os.Setenv("SINGLE_USER_LOGIN_EMAIL", "test@test.com")
		Expect(err).ShouldNot(HaveOccurred())
	})

	AfterEach(func() {
		testuutils.DropAllCollections(ctx, collProfiles, collHomes, collDevices)
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

			// create an expired JWY
			expirationTime := time.Now().Add(-60 * time.Minute)
			tokenString, err := utils.CreateJWT(profileRes, expirationTime, jwt.SigningMethodHS256, jwtKey)
			Expect(err).ShouldNot(HaveOccurred())
			logger.Infof("tokenString = %s", tokenString)

			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/profiles/"+profileRes.ID.Hex()+"/tokens", nil)
			req.Header.Add("Cookie", cookieSession)
			req.Header.Add("Authorization", "Bearer "+tokenString)
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusUnauthorized))
			Expect(recorder.Body.String()).To(Equal(`{"error":"not logged, token is not valid"}`))
		})
	})
})
