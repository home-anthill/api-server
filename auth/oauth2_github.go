package auth

import (
	"api-server/models"
	"api-server/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"
)

const (
	GitHubAuthorizeURL   = "https://github.com/login/oauth/authorize"
	GitHubAccessTokenURL = "https://github.com/login/oauth/access_token"
	GitHubCurrentUserURL = "https://api.github.com/user"
)

type GitHubOAuthClient string

const (
	GitHubOAuthClientWeb GitHubOAuthClient = "web"
	GitHubOAuthClientApp GitHubOAuthClient = "app"
)

const (
	RefreshTokenClientWeb    = "web"
	RefreshTokenClientMobile = "mobile"
)

// BuildGitHubAuthorizationURL builds the GitHub authorization URL with state
// and S256 PKCE challenge parameters.
func BuildGitHubAuthorizationURL(clientType GitHubOAuthClient, state, codeChallenge string) (string, error) {
	clientID, _, redirectURL, scopes, err := resolveGitHubOAuthConfig(clientType)
	if err != nil {
		return "", err
	}

	u, err := url.Parse(GitHubAuthorizeURL)
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Set("client_id", clientID)
	q.Set("redirect_uri", redirectURL)
	q.Set("response_type", "code")
	q.Set("scope", strings.Join(scopes, " "))
	q.Set("state", state)
	q.Set("code_challenge", codeChallenge)
	q.Set("code_challenge_method", utils.PKCEChallengeMethodS256)
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// FetchGitHubUser loads the current GitHub user profile with a GitHub OAuth
// access token and maps the response into the local GitHub profile model.
func FetchGitHubUser(ctx context.Context, httpClient *http.Client, accessToken string) (models.GitHub, error) {
	currentUserURL := githubEndpointURL(GitHubCurrentUserURL, "GITHUB_CURRENT_USER_URL")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, currentUserURL, nil)
	if err != nil {
		return models.GitHub{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("User-Agent", "home-anthill-api")

	resp, err := httpClient.Do(req)
	if err != nil {
		return models.GitHub{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return models.GitHub{}, fmt.Errorf("github user request failed with status %d", resp.StatusCode)
	}

	var currentUser models.GithubAppCurrentUserResponse
	if err = json.NewDecoder(resp.Body).Decode(&currentUser); err != nil {
		return models.GitHub{}, err
	}
	if currentUser.ID == 0 || currentUser.Login == "" {
		return models.GitHub{}, fmt.Errorf("github user response missing required identity fields")
	}

	return models.GitHub(currentUser), nil
}

// ExchangeGitHubCodeForAccessToken exchanges a GitHub OAuth authorization code
// and PKCE verifier for a GitHub access token.
func ExchangeGitHubCodeForAccessToken(ctx context.Context, httpClient *http.Client, clientType GitHubOAuthClient, code, pkceVerifier string) (string, error) {
	clientID, clientSecret, redirectURL, _, err := resolveGitHubOAuthConfig(clientType)
	if err != nil {
		return "", err
	}

	accessTokenURL := githubEndpointURL(GitHubAccessTokenURL, "GITHUB_OAUTH_ACCESS_TOKEN_URL")

	form := url.Values{}
	form.Set("client_id", clientID)
	form.Set("client_secret", clientSecret)
	form.Set("code", code)
	form.Set("redirect_uri", redirectURL)
	form.Set("code_verifier", pkceVerifier)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, accessTokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var tokenResp models.GithubAccessTokenResponse
	if err = json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", err
	}
	if tokenResp.Error != "" {
		return "", fmt.Errorf("github token exchange failed: %s", tokenResp.Error)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github token exchange failed with status %d", resp.StatusCode)
	}
	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("github token exchange returned empty access token")
	}
	if tokenResp.TokenType != "" && !strings.EqualFold(tokenResp.TokenType, "bearer") {
		return "", fmt.Errorf("github token exchange returned unsupported token type %q", tokenResp.TokenType)
	}
	return tokenResp.AccessToken, nil
}

func resolveGitHubOAuthConfig(clientType GitHubOAuthClient) (string, string, string, []string, error) {
	scopes := []string{"read:user", "user:email"}

	switch clientType {
	case GitHubOAuthClientWeb:
		return os.Getenv("OAUTH2_CLIENTID"), os.Getenv("OAUTH2_SECRETID"), os.Getenv("OAUTH2_CALLBACK"), scopes, nil
	case GitHubOAuthClientApp:
		return os.Getenv("OAUTH2_APP_CLIENTID"), os.Getenv("OAUTH2_APP_SECRETID"), os.Getenv("OAUTH2_APP_CALLBACK"), scopes, nil
	default:
		return "", "", "", nil, fmt.Errorf("unsupported GitHub OAuth client type %q", clientType)
	}
}

func githubEndpointURL(defaultURL, envName string) string {
	if os.Getenv("ENV") == "testing" {
		if override := strings.TrimSpace(os.Getenv(envName)); override != "" {
			return override
		}
	}
	return defaultURL
}

// FindOrCreateGitHubProfile returns the local profile for a GitHub identity,
// creating one when this is the first successful login for that GitHub user.
func FindOrCreateGitHubProfile(ctx context.Context, logger *zap.SugaredLogger, collProfiles *mongo.Collection, githubProfile models.GitHub) (models.Profile, error) {
	singleUserLoginEmail := os.Getenv("SINGLE_USER_LOGIN_EMAIL")
	if singleUserLoginEmail != "" {
		if githubProfile.Email == "" || githubProfile.Email != singleUserLoginEmail {
			return models.Profile{}, fmt.Errorf("login not permitted")
		}
	}

	var profile models.Profile
	err := collProfiles.FindOne(ctx, bson.M{"github.id": githubProfile.ID}).Decode(&profile)
	if err == nil {
		logger.Infow("AUDIT - user login",
			"profileID", profile.ID.Hex(),
			"githubLogin", githubProfile.Login,
		)
		return profile, nil
	}
	if !errors.Is(err, mongo.ErrNoDocuments) {
		return models.Profile{}, err
	}

	now := time.Now().UTC()
	profile = models.Profile{
		ID:         bson.NewObjectID(),
		Github:     githubProfile,
		APIToken:   uuid.NewString(),
		Homes:      []bson.ObjectID{},
		Devices:    []bson.ObjectID{},
		CreatedAt:  now,
		ModifiedAt: now,
	}

	if _, err = collProfiles.InsertOne(ctx, profile); err != nil {
		return models.Profile{}, err
	}

	logger.Infow("AUDIT - user created",
		"profileID", profile.ID.Hex(),
		"githubLogin", githubProfile.Login,
	)
	return profile, nil
}

// IssueGitHubLoginResult creates a local access JWT and an opaque refresh token
// for a successful GitHub login. Only the refresh-token hash is stored.
func IssueGitHubLoginResult(ctx context.Context, collRefreshTokens *mongo.Collection, profile models.Profile, jwtKey []byte, accessTokenTTL, refreshTokenTTL time.Duration, refreshTokenClientType string) (string, string, time.Time, error) {
	now := time.Now().UTC()
	accessTokenExpTime := now.Add(accessTokenTTL)
	accessToken, err := utils.CreateJWT(profile, accessTokenExpTime, utils.AccessToken, jwtKey)
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("create access token: %w", err)
	}

	refreshToken, err := utils.RandomString(64)
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("create refresh token: %w", err)
	}

	refreshTokenRecord := models.RefreshToken{
		ID:         bson.NewObjectID(),
		ProfileID:  profile.ID,
		TokenHash:  utils.HashToken(refreshToken),
		FamilyID:   bson.NewObjectID().Hex(),
		ClientType: refreshTokenClientType,
		CreatedAt:  now,
		ExpiresAt:  now.Add(refreshTokenTTL),
	}
	if _, err = collRefreshTokens.InsertOne(ctx, refreshTokenRecord); err != nil {
		return "", "", time.Time{}, fmt.Errorf("store refresh token: %w", err)
	}

	return accessToken, refreshToken, accessTokenExpTime, nil
}
