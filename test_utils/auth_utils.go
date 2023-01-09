package test_utils

import (
  "api-server/models"
  "encoding/json"
  "github.com/gin-gonic/gin"
  "github.com/onsi/gomega"
  "net/http/httptest"
  "os"
  "strings"
)

func GetJwt(router *gin.Engine) (string, string) {
  recorder := httptest.NewRecorder()
  req := httptest.NewRequest("GET", "/api/login", nil)
  router.ServeHTTP(recorder, req)
  gomega.Expect(200).To(gomega.Equal(recorder.Code))
  response := models.LoginUrl{}
  err := json.Unmarshal(recorder.Body.Bytes(), &response)
  gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
  prefix := `https://github.com/login/oauth/authorize?client_id=` +
    os.Getenv("OAUTH2_CLIENTID") +
    `&redirect_uri=http%3A%2F%2Flocalhost%3A8082%2Fapi%2Fcallback%2F&response_type=code&scope=repo&state=`
  stateEscapedB64 := strings.ReplaceAll(response.LoginURL, prefix, "")
  cookieSession := recorder.Header().Get("Set-Cookie")

  recorder = httptest.NewRecorder()
  req = httptest.NewRequest("GET", "/api/callback?code=ecf9b407c22a56b61ecf&state="+stateEscapedB64, nil)
  req.Header.Add("Cookie", cookieSession)
  router.ServeHTTP(recorder, req)
  gomega.Expect(recorder.Code).To(gomega.Equal(302))
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
