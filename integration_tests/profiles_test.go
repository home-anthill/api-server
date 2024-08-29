package integration_tests

import (
	"api-server/api"
	"api-server/db"
	"api-server/initialization"
	"api-server/testuutils"
	"encoding/json"
	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"net/http"
	"net/http/httptest"
	"os"
)

type newTokenResponse struct {
	APIToken string `json:"apiToken"`
}

var _ = Describe("Profiles", func() {
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

	Context("calling profiles api", func() {
		It("should return logged profile", func() {
			jwtToken, cookieSession := testuutils.GetJwt(router)
			profileRes := testuutils.GetLoggedProfile(router, jwtToken, cookieSession)

			Expect(profileRes.Homes).To(HaveLen(0))
			Expect(profileRes.Devices).To(HaveLen(0))
			Expect(profileRes.Github).To(Equal(api.DbGithubUserTestmock))
		})

		It("should generate a new profile api token", func() {
			jwtToken, cookieSession := testuutils.GetJwt(router)
			profileRes := testuutils.GetLoggedProfile(router, jwtToken, cookieSession)

			profileID := profileRes.ID.Hex()

			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/profiles/"+profileID+"/tokens", nil)
			req.Header.Add("Cookie", cookieSession)
			req.Header.Add("Authorization", "Bearer "+jwtToken)
			req.Header.Add("Content-Type", `application/json`)
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusOK))
			var newTokenRes newTokenResponse
			err := json.Unmarshal(recorder.Body.Bytes(), &newTokenRes)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(newTokenRes.APIToken).To(Not(BeNil()))
			// apiToken is an UUIDv4 token of 36 bytes
			Expect([]byte(newTokenRes.APIToken)).To(HaveLen(36))
		})

		It("should return an error, if profileId is wrong", func() {
			jwtToken, cookieSession := testuutils.GetJwt(router)
			profileID := "bad_profile_id"
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/profiles/"+profileID+"/tokens", nil)
			req.Header.Add("Cookie", cookieSession)
			req.Header.Add("Authorization", "Bearer "+jwtToken)
			req.Header.Add("Content-Type", `application/json`)
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusBadRequest))
			Expect(recorder.Body.String()).To(Equal(`{"error":"wrong format of the path param 'id'"}`))
		})

		It("should return an error, if profileId is not the one in session", func() {
			jwtToken, cookieSession := testuutils.GetJwt(router)
			profileID := primitive.NewObjectID()
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/profiles/"+profileID.Hex()+"/tokens", nil)
			req.Header.Add("Cookie", cookieSession)
			req.Header.Add("Authorization", "Bearer "+jwtToken)
			req.Header.Add("Content-Type", `application/json`)
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusBadRequest))
			Expect(recorder.Body.String()).To(Equal(`{"error":"cannot re-generate APIToken for a different profile then yours"}`))
		})
	})
})
