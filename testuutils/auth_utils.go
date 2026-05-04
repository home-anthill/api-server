package testuutils

import (
	"api-server/models"
	"api-server/utils"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/onsi/gomega"
)

func joinCookies(recorder *httptest.ResponseRecorder) string {
	return strings.Join(recorder.Header().Values("Set-Cookie"), "; ")
}

func GetJwt(router *gin.Engine) (string, string) {
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/oauth/login", nil)
	router.ServeHTTP(recorder, req)
	gomega.Expect(http.StatusTemporaryRedirect).To(gomega.Equal(recorder.Code))

	resLocation, err := recorder.Result().Location()
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	stateEscapedB64 := resLocation.Query().Get("state")
	gomega.Expect(stateEscapedB64).ToNot(gomega.BeEmpty())
	cookieSession := recorder.Header().Get("Set-Cookie")

	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/oauth/callback?code=ecf9b407c22a56b61ecf&state="+stateEscapedB64, nil)
	req.Header.Add("Cookie", cookieSession)
	router.ServeHTTP(recorder, req)
	gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusFound))
	redirectUrl := recorder.Header().Get("location")
	gomega.Expect(redirectUrl).To(gomega.HavePrefix("/postlogin#token="))
	jwtToken := strings.ReplaceAll(redirectUrl, "/postlogin#token=", "")
	allCookies := joinCookies(recorder)
	return jwtToken, allCookies
}

func GetJwtMobileApp(router *gin.Engine) (string, string) {
	codeVerifier, err := utils.NewPKCEVerifier()
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	codeChallenge, err := utils.BuildPKCECodeChallenge(codeVerifier)
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	appState, err := utils.NewPKCEVerifier()
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/oauth/app/login?code_challenge="+url.QueryEscape(codeChallenge)+"&code_challenge_method="+utils.PKCEChallengeMethodS256+"&app_state="+url.QueryEscape(appState), nil)
	router.ServeHTTP(recorder, req)
	gomega.Expect(http.StatusTemporaryRedirect).To(gomega.Equal(recorder.Code))
	resLocation, err := recorder.Result().Location()
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	stateEscapedB64 := resLocation.Query().Get("state")
	cookieSession := recorder.Header().Get("Set-Cookie")

	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/oauth/app/callback?code=ecf9b407c22a56b61ecf&state="+stateEscapedB64, nil)
	req.Header.Add("Cookie", cookieSession)
	router.ServeHTTP(recorder, req)
	gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusFound))
	redirectUrl := recorder.Header().Get("location")
	callbackURL, err := url.Parse(os.Getenv("OAUTH2_APP_CALLBACK"))
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	gomega.Expect(redirectUrl).To(gomega.HavePrefix(callbackURL.Scheme + "://" + callbackURL.Host + "/app/postlogin?code="))

	redirectLocation, err := url.Parse(redirectUrl)
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	code := redirectLocation.Query().Get("code")
	gomega.Expect(code).ToNot(gomega.BeEmpty())
	gomega.Expect(redirectLocation.Query().Get("state")).To(gomega.Equal(appState))

	exchangeBody := `{"code":"` + code + `","codeVerifier":"` + codeVerifier + `"}`
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/oauth/app/exchange-code", strings.NewReader(exchangeBody))
	req.Header.Add("Content-Type", "application/json")
	router.ServeHTTP(recorder, req)
	gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusOK))

	var exchangeResp struct {
		Token        string `json:"token"`
		RefreshToken string `json:"refreshToken"`
	}
	err = json.Unmarshal(recorder.Body.Bytes(), &exchangeResp)
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	gomega.Expect(exchangeResp.Token).ToNot(gomega.BeEmpty())
	gomega.Expect(exchangeResp.RefreshToken).ToNot(gomega.BeEmpty())
	return exchangeResp.Token, exchangeResp.RefreshToken
}

func GetLoggedProfile(router *gin.Engine, jwtToken, cookieSession string) models.Profile {
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/profile", nil)
	req.Header.Add("Cookie", cookieSession)
	req.Header.Add("Authorization", "Bearer "+jwtToken)
	req.Header.Add("Content-Type", `application/json`)
	router.ServeHTTP(recorder, req)
	gomega.Expect(recorder.Code).To(gomega.Equal(200))
	var profileRes models.Profile
	err := json.Unmarshal(recorder.Body.Bytes(), &profileRes)
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	return profileRes
}
