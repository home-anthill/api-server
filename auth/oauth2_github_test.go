package auth

import (
	"api-server/models"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"
)

func TestIsGitHubEmailAllowed(t *testing.T) {
	tests := []struct {
		name          string
		email         string
		allowedEmails string
		want          bool
	}{
		{
			name:          "empty allowlist permits any email",
			email:         "user@example.com",
			allowedEmails: "",
			want:          true,
		},
		{
			name:          "email in comma separated allowlist is permitted",
			email:         "second@example.com",
			allowedEmails: "first@example.com,second@example.com",
			want:          true,
		},
		{
			name:          "email match ignores spaces and case",
			email:         "Second@Example.com",
			allowedEmails: " first@example.com, second@example.com ",
			want:          true,
		},
		{
			name:          "email outside allowlist is rejected",
			email:         "third@example.com",
			allowedEmails: "first@example.com,second@example.com",
			want:          false,
		},
		{
			name:          "empty github email is rejected when allowlist is set",
			email:         "",
			allowedEmails: "first@example.com",
			want:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isGitHubEmailAllowed(tt.email, tt.allowedEmails)
			if got != tt.want {
				t.Fatalf("isGitHubEmailAllowed(%q, %q) = %t, want %t", tt.email, tt.allowedEmails, got, tt.want)
			}
		})
	}
}

func TestBuildGitHubAuthorizationURL(t *testing.T) {
	t.Setenv("OAUTH2_CLIENTID", "web-client")
	t.Setenv("OAUTH2_SECRETID", "web-secret")
	t.Setenv("OAUTH2_CALLBACK", "https://example.com/web/callback")
	t.Setenv("OAUTH2_APP_CLIENTID", "app-client")
	t.Setenv("OAUTH2_APP_SECRETID", "app-secret")
	t.Setenv("OAUTH2_APP_CALLBACK", "https://example.com/app/callback")

	tests := []struct {
		name         string
		clientType   GitHubOAuthClient
		wantClientID string
		wantRedirect string
	}{
		{name: "web", clientType: GitHubOAuthClientWeb, wantClientID: "web-client", wantRedirect: "https://example.com/web/callback"},
		{name: "app", clientType: GitHubOAuthClientApp, wantClientID: "app-client", wantRedirect: "https://example.com/app/callback"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildGitHubAuthorizationURL(tt.clientType, "state", "challenge")
			if err != nil {
				t.Fatalf("BuildGitHubAuthorizationURL returned error: %v", err)
			}
			parsed, err := url.Parse(got)
			if err != nil {
				t.Fatalf("Parse returned error: %v", err)
			}
			if parsed.Scheme != "https" || parsed.Host != "github.com" || parsed.Path != "/login/oauth/authorize" {
				t.Fatalf("unexpected authorization URL: %s", got)
			}
			q := parsed.Query()
			if q.Get("client_id") != tt.wantClientID {
				t.Fatalf("client_id = %q, want %q", q.Get("client_id"), tt.wantClientID)
			}
			if q.Get("redirect_uri") != tt.wantRedirect {
				t.Fatalf("redirect_uri = %q, want %q", q.Get("redirect_uri"), tt.wantRedirect)
			}
			if q.Get("response_type") != "code" || q.Get("scope") != "read:user user:email" || q.Get("state") != "state" {
				t.Fatalf("unexpected query values: %s", parsed.RawQuery)
			}
			if q.Get("code_challenge") != "challenge" || q.Get("code_challenge_method") != "S256" {
				t.Fatalf("unexpected PKCE query values: %s", parsed.RawQuery)
			}
		})
	}

	if got, err := BuildGitHubAuthorizationURL("desktop", "state", "challenge"); err == nil || got != "" {
		t.Fatalf("BuildGitHubAuthorizationURL unsupported client = %q, %v; want empty error", got, err)
	}
}

func TestFetchGitHubUser(t *testing.T) {
	t.Run("maps successful response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Fatalf("method = %s, want GET", r.Method)
			}
			if r.Header.Get("Authorization") != "Bearer github-token" {
				t.Fatalf("Authorization = %q", r.Header.Get("Authorization"))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":         123,
				"login":      "octocat",
				"name":       "Octo Cat",
				"email":      "octo@example.com",
				"avatar_url": "https://example.com/avatar.png",
			})
		}))
		defer server.Close()
		t.Setenv("ENV", "testing")
		t.Setenv("GITHUB_CURRENT_USER_URL", server.URL)

		got, err := FetchGitHubUser(context.Background(), server.Client(), "github-token")
		if err != nil {
			t.Fatalf("FetchGitHubUser returned error: %v", err)
		}
		if got.ID != 123 || got.Login != "octocat" || got.AvatarURL == "" {
			t.Fatalf("FetchGitHubUser() = %#v", got)
		}
	})

	t.Run("rejects non ok status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer server.Close()
		t.Setenv("ENV", "testing")
		t.Setenv("GITHUB_CURRENT_USER_URL", server.URL)

		if _, err := FetchGitHubUser(context.Background(), server.Client(), "github-token"); err == nil {
			t.Fatal("expected non ok status error")
		}
	})

	t.Run("rejects invalid json", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("{"))
		}))
		defer server.Close()
		t.Setenv("ENV", "testing")
		t.Setenv("GITHUB_CURRENT_USER_URL", server.URL)

		if _, err := FetchGitHubUser(context.Background(), server.Client(), "github-token"); err == nil {
			t.Fatal("expected decode error")
		}
	})

	t.Run("rejects missing identity fields", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode(map[string]any{"id": 123})
		}))
		defer server.Close()
		t.Setenv("ENV", "testing")
		t.Setenv("GITHUB_CURRENT_USER_URL", server.URL)

		if _, err := FetchGitHubUser(context.Background(), server.Client(), "github-token"); err == nil {
			t.Fatal("expected missing identity error")
		}
	})
}

func TestExchangeGitHubCodeForAccessToken(t *testing.T) {
	setOAuthEnv(t)

	t.Run("returns access token", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Fatalf("method = %s, want POST", r.Method)
			}
			if err := r.ParseForm(); err != nil {
				t.Fatalf("ParseForm returned error: %v", err)
			}
			if r.Form.Get("client_id") != "web-client" || r.Form.Get("client_secret") != "web-secret" {
				t.Fatalf("unexpected client credentials: %v", r.Form)
			}
			if r.Form.Get("code") != "code" || r.Form.Get("code_verifier") != "verifier" {
				t.Fatalf("unexpected code form: %v", r.Form)
			}
			_ = json.NewEncoder(w).Encode(map[string]string{"access_token": "access", "token_type": "bearer"})
		}))
		defer server.Close()
		t.Setenv("ENV", "testing")
		t.Setenv("GITHUB_OAUTH_ACCESS_TOKEN_URL", server.URL)

		got, err := ExchangeGitHubCodeForAccessToken(context.Background(), server.Client(), GitHubOAuthClientWeb, "code", "verifier")
		if err != nil {
			t.Fatalf("ExchangeGitHubCodeForAccessToken returned error: %v", err)
		}
		if got != "access" {
			t.Fatalf("token = %q, want access", got)
		}
	})

	tests := []struct {
		name       string
		statusCode int
		body       string
	}{
		{name: "github error", statusCode: http.StatusOK, body: `{"error":"bad_verification_code"}`},
		{name: "non ok status", statusCode: http.StatusBadGateway, body: `{"access_token":"access","token_type":"bearer"}`},
		{name: "empty access token", statusCode: http.StatusOK, body: `{"token_type":"bearer"}`},
		{name: "unsupported token type", statusCode: http.StatusOK, body: `{"access_token":"access","token_type":"mac"}`},
		{name: "invalid json", statusCode: http.StatusOK, body: `{`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer server.Close()
			t.Setenv("ENV", "testing")
			t.Setenv("GITHUB_OAUTH_ACCESS_TOKEN_URL", server.URL)

			if got, err := ExchangeGitHubCodeForAccessToken(context.Background(), server.Client(), GitHubOAuthClientWeb, "code", "verifier"); err == nil || got != "" {
				t.Fatalf("ExchangeGitHubCodeForAccessToken = %q, %v; want empty error", got, err)
			}
		})
	}

	if got, err := ExchangeGitHubCodeForAccessToken(context.Background(), http.DefaultClient, "desktop", "code", "verifier"); err == nil || got != "" {
		t.Fatalf("unsupported client exchange = %q, %v; want empty error", got, err)
	}
}

func TestResolveGitHubOAuthConfig(t *testing.T) {
	setOAuthEnv(t)

	clientID, secret, redirectURL, scopes, err := resolveGitHubOAuthConfig(GitHubOAuthClientWeb)
	if err != nil {
		t.Fatalf("resolveGitHubOAuthConfig returned error: %v", err)
	}
	if clientID != "web-client" || secret != "web-secret" || redirectURL != "https://example.com/web/callback" {
		t.Fatalf("unexpected web config: %q %q %q", clientID, secret, redirectURL)
	}
	if strings.Join(scopes, " ") != "read:user user:email" {
		t.Fatalf("scopes = %v", scopes)
	}

	clientID, secret, redirectURL, _, err = resolveGitHubOAuthConfig(GitHubOAuthClientApp)
	if err != nil {
		t.Fatalf("resolveGitHubOAuthConfig returned error: %v", err)
	}
	if clientID != "app-client" || secret != "app-secret" || redirectURL != "https://example.com/app/callback" {
		t.Fatalf("unexpected app config: %q %q %q", clientID, secret, redirectURL)
	}

	if _, _, _, _, err = resolveGitHubOAuthConfig("desktop"); err == nil {
		t.Fatal("expected unsupported client error")
	}
}

func TestGitHubEndpointURL(t *testing.T) {
	t.Setenv("ENV", "testing")
	t.Setenv("GITHUB_TEST_URL", "http://127.0.0.1/github")
	if got := githubEndpointURL("https://github.com/default", "GITHUB_TEST_URL"); got != "http://127.0.0.1/github" {
		t.Fatalf("githubEndpointURL testing override = %q", got)
	}

	t.Setenv("ENV", "prod")
	if got := githubEndpointURL("https://github.com/default", "GITHUB_TEST_URL"); got != "https://github.com/default" {
		t.Fatalf("githubEndpointURL prod = %q", got)
	}
}

func TestFindOrCreateGitHubProfile(t *testing.T) {
	t.Run("rejects disallowed email before database lookup", func(t *testing.T) {
		t.Setenv("LIMIT_TO_USER_EMAILS", "allowed@example.com")
		_, err := FindOrCreateGitHubProfile(context.Background(), zap.NewNop().Sugar(), nil, models.GitHub{
			ID:    123,
			Login: "octocat",
			Email: "blocked@example.com",
		})
		if err == nil || !strings.Contains(err.Error(), "login not permitted") {
			t.Fatalf("FindOrCreateGitHubProfile error = %v, want login not permitted", err)
		}
	})

	t.Run("returns database lookup error", func(t *testing.T) {
		t.Setenv("LIMIT_TO_USER_EMAILS", "")
		client := disconnectedMongoClient(t)
		defer func() { _ = client.Disconnect(context.Background()) }()
		_, err := FindOrCreateGitHubProfile(context.Background(), zap.NewNop().Sugar(), client.Database("unit").Collection("profiles"), models.GitHub{
			ID:    123,
			Login: "octocat",
			Email: "octo@example.com",
		})
		if err == nil {
			t.Fatal("expected database lookup error")
		}
	})
}

func TestIssueGitHubLoginResult(t *testing.T) {
	client := disconnectedMongoClient(t)
	defer func() { _ = client.Disconnect(context.Background()) }()

	profile := models.Profile{
		ID: bson.NewObjectID(),
		Github: models.GitHub{
			ID:   123,
			Name: "Octo Cat",
		},
	}
	accessToken, refreshToken, expiresAt, err := IssueGitHubLoginResult(
		context.Background(),
		client.Database("unit").Collection("refresh_tokens"),
		profile,
		[]byte("0123456789abcdef0123456789abcdef"),
		time.Minute,
		time.Hour,
		RefreshTokenClientWeb,
	)
	if err == nil {
		t.Fatal("expected refresh token store error")
	}
	if accessToken != "" || refreshToken != "" || !expiresAt.IsZero() {
		t.Fatalf("IssueGitHubLoginResult = %q, %q, %v, %v; want zero values with error", accessToken, refreshToken, expiresAt, err)
	}
	if !strings.Contains(err.Error(), "store refresh token") {
		t.Fatalf("IssueGitHubLoginResult error = %v, want store refresh token", err)
	}
}

func setOAuthEnv(t *testing.T) {
	t.Helper()
	t.Setenv("OAUTH2_CLIENTID", "web-client")
	t.Setenv("OAUTH2_SECRETID", "web-secret")
	t.Setenv("OAUTH2_CALLBACK", "https://example.com/web/callback")
	t.Setenv("OAUTH2_APP_CLIENTID", "app-client")
	t.Setenv("OAUTH2_APP_SECRETID", "app-secret")
	t.Setenv("OAUTH2_APP_CALLBACK", "https://example.com/app/callback")
}

func disconnectedMongoClient(t *testing.T) *mongo.Client {
	t.Helper()
	client, err := mongo.Connect(options.Client().
		ApplyURI("mongodb://127.0.0.1:1").
		SetServerSelectionTimeout(10 * time.Millisecond))
	if err != nil {
		t.Fatalf("mongo.Connect returned error: %v", err)
	}
	return client
}
