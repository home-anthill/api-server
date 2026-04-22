package integration_tests_test

import (
	"api-server/models"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTests(t *testing.T) {
	if err := os.Setenv("ENV", "testing"); err != nil {
		t.Fatalf("cannot force ENV=testing: %v", err)
	}

	githubMock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/login/oauth/access_token":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{
				"access_token": "test-github-access-token",
				"token_type":   "bearer",
				"scope":        "read:user,user:email",
			})
		case "/user":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id":         models.DbGithubUserTestmock.ID,
				"login":      models.DbGithubUserTestmock.Login,
				"name":       models.DbGithubUserTestmock.Name,
				"email":      models.DbGithubUserTestmock.Email,
				"avatar_url": models.DbGithubUserTestmock.AvatarURL,
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer githubMock.Close()

	if err := os.Setenv("GITHUB_OAUTH_ACCESS_TOKEN_URL", githubMock.URL+"/login/oauth/access_token"); err != nil {
		t.Fatalf("cannot set GITHUB_OAUTH_ACCESS_TOKEN_URL: %v", err)
	}
	if err := os.Setenv("GITHUB_CURRENT_USER_URL", githubMock.URL+"/user"); err != nil {
		t.Fatalf("cannot set GITHUB_CURRENT_USER_URL: %v", err)
	}

	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration tests")
}
