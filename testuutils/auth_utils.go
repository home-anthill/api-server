package testuutils

import (
	"api-server/models"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
)

func GetJwt(router *gin.Engine) (string, string) {
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/login", nil)
	router.ServeHTTP(recorder, req)
	gomega.Expect(http.StatusFound).To(gomega.Equal(recorder.Code))

	resLocation, _ := recorder.Result().Location()
	callbackURL := url.QueryEscape(os.Getenv("OAUTH2_CALLBACK"))
	suffix := `client_id=` +
		os.Getenv("OAUTH2_CLIENTID") +
		`&redirect_uri=` + callbackURL + `&response_type=code&scope=repo&state=`

	stateEscapedB64 := strings.ReplaceAll(resLocation.RawQuery, suffix, "")
	cookieSession := recorder.Header().Get("Set-Cookie")

	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/callback?code=ecf9b407c22a56b61ecf&state="+stateEscapedB64, nil)
	req.Header.Add("Cookie", cookieSession)
	router.ServeHTTP(recorder, req)
	gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusFound))
	redirectUrl := recorder.Header().Get("location")
	gomega.Expect(redirectUrl).To(gomega.HavePrefix("/postlogin?token="))
	jwtToken := strings.ReplaceAll(redirectUrl, "/postlogin?token=", "")
	cookieSession = recorder.Header().Get("Set-Cookie")
	return jwtToken, cookieSession
}

func GetJwtMobileApp(router *gin.Engine) (string, string) {
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/login_app", nil)
	router.ServeHTTP(recorder, req)
	gomega.Expect(http.StatusFound).To(gomega.Equal(recorder.Code))

	resLocation, _ := recorder.Result().Location()
	callbackURL := url.QueryEscape(os.Getenv("OAUTH2_APP_CALLBACK"))
	suffix := `client_id=` +
		os.Getenv("OAUTH2_CLIENTID") +
		`&redirect_uri=` + callbackURL + `&response_type=code&scope=repo&state=`

	stateEscapedB64 := strings.ReplaceAll(resLocation.RawQuery, suffix, "")
	cookieSession := recorder.Header().Get("Set-Cookie")

	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/callback?code=ecf9b407c22a56b61ecf&state="+stateEscapedB64, nil)
	req.Header.Add("Cookie", cookieSession)
	router.ServeHTTP(recorder, req)
	gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusFound))
	redirectUrl := recorder.Header().Get("location")
	gomega.Expect(redirectUrl).To(gomega.HavePrefix("/postlogin?token="))
	jwtToken := strings.ReplaceAll(redirectUrl, "/postlogin?token=", "")
	cookieSession = recorder.Header().Get("Set-Cookie")
	return jwtToken, cookieSession
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
