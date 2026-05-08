package integration_tests

import (
	"api-server/api"
	"api-server/db"
	"api-server/initialization"
	"api-server/models"
	"api-server/testuutils"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"
)

type newTokenResponse struct {
	APIToken string `json:"apiToken"`
}

type updateFCMTokenResponse struct {
	Message string `json:"message"`
}

type rotateOnlineAPITokenPayload struct {
	OldAPIToken string `json:"oldApiToken"`
	NewAPIToken string `json:"newApiToken"`
}

var _ = Describe("Profiles", func() {
	var ctx context.Context
	var logger *zap.SugaredLogger
	var router *gin.Engine
	var client *mongo.Client
	var collProfiles *mongo.Collection
	var collHomes *mongo.Collection
	var collDevices *mongo.Collection
	var onlineMockServer *httptest.Server
	var oldOnlineServer string
	var oldOnlinePort string
	var oldOnlineServerSet bool
	var oldOnlinePortSet bool

	BeforeEach(func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/keepalive/", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			w.WriteHeader(http.StatusOK)
		})
		mux.HandleFunc("/api-token/rotate", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			var payload rotateOnlineAPITokenPayload
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid json", http.StatusBadRequest)
				return
			}
			if payload.OldAPIToken == "" || payload.NewAPIToken == "" {
				http.Error(w, "missing api token", http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusOK)
		})
		onlineMockServer = httptest.NewServer(mux)
		onlineMockURL, err := url.Parse(onlineMockServer.URL)
		Expect(err).ShouldNot(HaveOccurred())
		oldOnlineServer, oldOnlineServerSet = os.LookupEnv("HTTP_ONLINE_SERVER")
		oldOnlinePort, oldOnlinePortSet = os.LookupEnv("HTTP_ONLINE_PORT")
		err = os.Setenv("HTTP_ONLINE_SERVER", onlineMockURL.Scheme+"://"+onlineMockURL.Hostname())
		Expect(err).ShouldNot(HaveOccurred())
		err = os.Setenv("HTTP_ONLINE_PORT", onlineMockURL.Port())
		Expect(err).ShouldNot(HaveOccurred())

		logger, router, client = initialization.MustStart()
		ctx = context.Background()
		defer logger.Sync()

		collProfiles = db.GetCollections(client).Profiles
		collHomes = db.GetCollections(client).Homes
		collDevices = db.GetCollections(client).Devices

		err = os.Setenv("LIMIT_TO_USER_EMAILS", "test@test.com")
		Expect(err).ShouldNot(HaveOccurred())
	})

	AfterEach(func() {
		testuutils.DropAllCollections(ctx, collProfiles, collHomes, collDevices)
		onlineMockServer.Close()
		if oldOnlineServerSet {
			err := os.Setenv("HTTP_ONLINE_SERVER", oldOnlineServer)
			Expect(err).ShouldNot(HaveOccurred())
		} else {
			err := os.Unsetenv("HTTP_ONLINE_SERVER")
			Expect(err).ShouldNot(HaveOccurred())
		}
		if oldOnlinePortSet {
			err := os.Setenv("HTTP_ONLINE_PORT", oldOnlinePort)
			Expect(err).ShouldNot(HaveOccurred())
		} else {
			err := os.Unsetenv("HTTP_ONLINE_PORT")
			Expect(err).ShouldNot(HaveOccurred())
		}
	})

	Context("calling profiles api GET", func() {
		It("should return logged profile", func() {
			jwtToken, cookieSession := testuutils.GetJwt(router)
			profileRes := testuutils.GetLoggedProfile(router, jwtToken, cookieSession)

			Expect(profileRes.Homes).To(HaveLen(0))
			Expect(profileRes.Devices).To(HaveLen(0))
			Expect(profileRes.Github).To(Equal(models.DbGithubUserTestmock))
		})
	})

	Context("calling profiles apiToken api POST", func() {
		It("should generate a new profile apiToken", func() {
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
			profileID := bson.NewObjectID()
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

	Context("calling profiles fcmToken api POST", func() {
		It("should update existing profile with FCM Token", func() {
			jwtToken, cookieSession := testuutils.GetJwt(router)
			profileRes := testuutils.GetLoggedProfile(router, jwtToken, cookieSession)

			profileID := profileRes.ID.Hex()
			profileUpdateFCMTokenReq := api.ProfileUpdateFCMTokenReq{
				FCMToken: "MOCKED_FCM_TOKEN",
			}
			var profileUpdateBuf bytes.Buffer
			err := json.NewEncoder(&profileUpdateBuf).Encode(profileUpdateFCMTokenReq)
			Expect(err).ShouldNot(HaveOccurred())

			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/profiles/"+profileID+"/fcmTokens", &profileUpdateBuf)
			req.Header.Add("Cookie", cookieSession)
			req.Header.Add("Authorization", "Bearer "+jwtToken)
			req.Header.Add("Content-Type", `application/json`)
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusOK))
			var response updateFCMTokenResponse
			err = json.Unmarshal(recorder.Body.Bytes(), &response)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(response.Message).To(Equal("Profile update with FCM Token"))
		})

		It("should return an error, if profileId is wrong", func() {
			jwtToken, cookieSession := testuutils.GetJwt(router)
			profileID := "bad_profile_id"

			profileUpdateFCMTokenReq := api.ProfileUpdateFCMTokenReq{
				FCMToken: "MOCKED_FCM_TOKEN",
			}
			var profileUpdateBuf bytes.Buffer
			err := json.NewEncoder(&profileUpdateBuf).Encode(profileUpdateFCMTokenReq)
			Expect(err).ShouldNot(HaveOccurred())

			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/profiles/"+profileID+"/fcmTokens", &profileUpdateBuf)
			req.Header.Add("Cookie", cookieSession)
			req.Header.Add("Authorization", "Bearer "+jwtToken)
			req.Header.Add("Content-Type", `application/json`)
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusBadRequest))
			Expect(recorder.Body.String()).To(Equal(`{"error":"wrong format of the path param 'id'"}`))
		})

		It("should return an error, if profileId is not the one in session", func() {
			jwtToken, cookieSession := testuutils.GetJwt(router)
			profileID := bson.NewObjectID()

			profileUpdateFCMTokenReq := api.ProfileUpdateFCMTokenReq{
				FCMToken: "MOCKED_FCM_TOKEN",
			}
			var profileUpdateBuf bytes.Buffer
			err := json.NewEncoder(&profileUpdateBuf).Encode(profileUpdateFCMTokenReq)
			Expect(err).ShouldNot(HaveOccurred())

			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/profiles/"+profileID.Hex()+"/fcmTokens", &profileUpdateBuf)
			req.Header.Add("Cookie", cookieSession)
			req.Header.Add("Authorization", "Bearer "+jwtToken)
			req.Header.Add("Content-Type", `application/json`)
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusBadRequest))
			Expect(recorder.Body.String()).To(Equal(`{"error":"cannot set FCMToken for a different profile then yours"}`))
		})

		It("should return an error, if request body is not a JSON", func() {
			jwtToken, cookieSession := testuutils.GetJwt(router)
			profileRes := testuutils.GetLoggedProfile(router, jwtToken, cookieSession)
			profileID := profileRes.ID.Hex()

			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/profiles/"+profileID+"/fcmTokens", nil)
			req.Header.Add("Cookie", cookieSession)
			req.Header.Add("Authorization", "Bearer "+jwtToken)
			req.Header.Add("Content-Type", `application/json`)
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusBadRequest))
			Expect(recorder.Body.String()).To(Equal(`{"error":"invalid request payload"}`))
		})

		It("should return an error, if request body is not valid", func() {
			jwtToken, cookieSession := testuutils.GetJwt(router)
			profileRes := testuutils.GetLoggedProfile(router, jwtToken, cookieSession)
			profileID := profileRes.ID.Hex()

			profileUpdateFCMTokenReq := api.ProfileUpdateFCMTokenReq{
				// FCMToken is a required field, but here it's omitted
			}
			var profileUpdateBuf bytes.Buffer
			err := json.NewEncoder(&profileUpdateBuf).Encode(profileUpdateFCMTokenReq)
			Expect(err).ShouldNot(HaveOccurred())

			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/profiles/"+profileID+"/fcmTokens", &profileUpdateBuf)
			req.Header.Add("Cookie", cookieSession)
			req.Header.Add("Authorization", "Bearer "+jwtToken)
			req.Header.Add("Content-Type", `application/json`)
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusBadRequest))
			Expect(recorder.Body.String()).To(Equal(`{"error":"invalid request body, these fields are not valid: fcmtoken"}`))
		})
	})
})
