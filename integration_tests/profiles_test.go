package integration_tests

import (
	"api-server/api"
	"api-server/initialization"
	"api-server/test_utils"
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
	ApiToken string `json:"apiToken"`
}

var _ = Describe("Profiles", func() {
	var ctx context.Context
	var logger *zap.SugaredLogger
	var router *gin.Engine
	var collProfiles *mongo.Collection
	var collHomes *mongo.Collection
	var collDevices *mongo.Collection

	BeforeEach(func() {
		logger, router, ctx, collProfiles, collHomes, collDevices = initialization.Start()
		defer logger.Sync()

		err := os.Setenv("SINGLE_USER_LOGIN_EMAIL", "test@test.com")
		Expect(err).ShouldNot(HaveOccurred())
	})

	AfterEach(func() {
		test_utils.DropAllCollections(ctx, collProfiles, collHomes, collDevices)
	})

	Context("calling profiles api", func() {
		It("should return logged profile", func() {
			jwtToken, cookieSession := test_utils.GetJwt(router)
			profileRes := test_utils.GetLoggedProfile(router, jwtToken, cookieSession)

			Expect(profileRes.Homes).To(HaveLen(0))
			Expect(profileRes.Devices).To(HaveLen(0))
			Expect(profileRes.Github).To(Equal(api.DbGithubUserTestmock))
		})

		It("should generate a new profile api token", func() {
			jwtToken, cookieSession := test_utils.GetJwt(router)
			profileRes := test_utils.GetLoggedProfile(router, jwtToken, cookieSession)

			profileId := profileRes.ID.Hex()

			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/profiles/"+profileId+"/tokens", nil)
			req.Header.Add("Cookie", cookieSession)
			req.Header.Add("Authorization", "Bearer "+jwtToken)
			req.Header.Add("Content-Type", `application/json`)
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusOK))
			var newTokenRes newTokenResponse
			err := json.Unmarshal(recorder.Body.Bytes(), &newTokenRes)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(newTokenRes.ApiToken).To(Not(BeNil()))
			// apiToken is an UUIDv4 token of 36 bytes
			Expect([]byte(newTokenRes.ApiToken)).To(HaveLen(36))
		})

		It("should return an error, if profileId is wrong", func() {
			jwtToken, cookieSession := test_utils.GetJwt(router)
			profileId := "bad_profile_id"
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/profiles/"+profileId+"/tokens", nil)
			req.Header.Add("Cookie", cookieSession)
			req.Header.Add("Authorization", "Bearer "+jwtToken)
			req.Header.Add("Content-Type", `application/json`)
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusBadRequest))
			Expect(recorder.Body.String()).To(Equal(`{"error":"wrong format of the path param 'id'"}`))
		})

		It("should return an error, if profileId is not the one in session", func() {
			jwtToken, cookieSession := test_utils.GetJwt(router)
			profileId := primitive.NewObjectID()
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/profiles/"+profileId.Hex()+"/tokens", nil)
			req.Header.Add("Cookie", cookieSession)
			req.Header.Add("Authorization", "Bearer "+jwtToken)
			req.Header.Add("Content-Type", `application/json`)
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusBadRequest))
			Expect(recorder.Body.String()).To(Equal(`{"error":"cannot re-generate ApiToken for a different profile then yours"}`))
		})
	})
})
