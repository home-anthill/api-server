package api

import (
	authpkg "api-server/auth"
	"api-server/db"
	"api-server/utils"
	"context"
	"crypto/subtle"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"
)

type GitHubWebHandler struct {
	collProfiles              *mongo.Collection
	auth                      *authpkg.Auth
	logger                    *zap.SugaredLogger
	sessionStateName          string
	sessionGitHubVerifierName string
	httpClient                *http.Client
}

func NewGitHubWebHandler(auth *authpkg.Auth, logger *zap.SugaredLogger, client *mongo.Client, sessionStateName, sessionPKCEName string) *GitHubWebHandler {
	return &GitHubWebHandler{
		collProfiles:              db.GetCollections(client).Profiles,
		auth:                      auth,
		logger:                    logger,
		sessionStateName:          sessionStateName,
		sessionGitHubVerifierName: sessionPKCEName,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (gh *GitHubWebHandler) GitHubLogin(c *gin.Context) {
	gh.logger.Info("REST - GET - GitHubLogin called")

	session := sessions.Default(c)

	// build state for CSRF protection
	state, err := utils.RandomString(36)
	if err != nil {
		gh.logger.Error("REST - GET - GitHubLogin - cannot create random state token")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not initialize oauth flow"})
		return
	}

	// build PKCE plain secret verifier
	// (it will be used only on our server-side as a verification step)
	githubVerifier, err := utils.NewPKCEVerifier()
	if err != nil {
		gh.logger.Errorw("REST - GET - GitHubLogin - cannot create GitHub PKCE verifier", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not initialize oauth flow"})
		return
	}
	// build PKCE codeChallenge from verifier with sha256 and base64 function
	// This will be used later to send it to GitHub (so we send the hashed version and not the plain verifier code)
	githubCodeChallenge, err := utils.BuildPKCECodeChallenge(githubVerifier)
	if err != nil {
		gh.logger.Errorw("REST - GET - GitHubLogin - cannot create GitHub PKCE challenge", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not initialize oauth flow"})
		return
	}

	// store both state and GitHub PKCE verifier in session
	session.Set(gh.sessionStateName, state)
	session.Set(gh.sessionGitHubVerifierName, githubVerifier)

	if err = session.Save(); err != nil {
		gh.logger.Error("REST - GET - GitHubLogin - cannot save session")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not initialize oauth flow"})
		return
	}

	// authURL must expose only the S256 challenge. Keep the raw verifier server-side.
	authURL, err := authpkg.BuildGitHubAuthorizationURL(authpkg.GitHubOAuthClientWeb, state, githubCodeChallenge)
	if err != nil {
		gh.logger.Errorw("REST - GET - GitHubLogin - cannot build authorization URL", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not build authURL during oauth flow initialization"})
		return
	}

	gh.logger.Debug("REST - GET - GitHubLogin - authURL: ", authURL)
	c.Redirect(http.StatusTemporaryRedirect, authURL)
}

func (gh *GitHubWebHandler) GitHubCallback(c *gin.Context) {
	gh.auth.Logger.Info("REST - GET - GitHubCallback called")

	session := sessions.Default(c)
	defer func() {
		session.Delete(gh.sessionStateName)
		session.Delete(gh.sessionGitHubVerifierName)
		if err := session.Save(); err != nil {
			gh.logger.Warnw("GitHubCallback - cannot clear oauth session", "error", err)
		}
	}()

	// extract state (for CSRF protection)
	queryState := strings.TrimSpace(c.Query("state"))
	// extract code: a one-time authorization code from GitHub, used to get a GitHub access token.
	queryCode := strings.TrimSpace(c.Query("code"))
	if queryState == "" || queryCode == "" {
		gh.logger.Error("REST - GET - GitHubCallback - missing either state or code callback parameters")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid oauth callback"})
		return
	}

	// check if state and PKCE verifier are in session
	sessionState, _ := session.Get(gh.sessionStateName).(string)
	githubVerifier, _ := session.Get(gh.sessionGitHubVerifierName).(string)
	if sessionState == "" || !utils.IsValidPKCEVerifier(githubVerifier) {
		gh.logger.Error("REST - GET - GitHubCallback - oauth session is missing or expired")
		c.JSON(http.StatusBadRequest, gin.H{"error": "oauth session is missing or expired"})
		return
	}

	// state must be = to the one in session
	if subtle.ConstantTimeCompare([]byte(queryState), []byte(sessionState)) != 1 {
		gh.logger.Error("REST - GET - GitHubCallback - oauth state verification failed")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid oauth callback"})
		return
	}

	// 10s timeout, so the GitHub token exchange and profile request cannot pin the request indefinitely.
	// This is not really required because the timeout is already 10s, but in this way it's more explicit.
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// get GitHub access token passing both:
	// - one-time code
	// - PKCE plain verifier (the plain string, because GitHub will do the sha256 and base64 to compare it)
	githubAccessToken, err := authpkg.ExchangeGitHubCodeForAccessToken(
		ctx,
		gh.httpClient,
		authpkg.GitHubOAuthClientWeb,
		queryCode,
		githubVerifier,
	)
	if err != nil {
		gh.logger.Errorw("REST - GET - GitHubCallback - github token exchange failed", "error", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "github token exchange failed"})
		return
	}

	// get GitHub profile using the githubAccessToken
	githubProfile, err := authpkg.FetchGitHubUser(ctx, gh.httpClient, githubAccessToken)
	if err != nil {
		gh.logger.Errorw("REST - GET - GitHubCallback - could not load github profile", "error", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "could not load github profile"})
		return
	}

	// find existing local profile or create a new one
	profile, err := authpkg.FindOrCreateGitHubProfile(ctx, gh.logger, gh.collProfiles, githubProfile)
	if err != nil {
		gh.logger.Errorw("REST - GET - GitHubCallback - could not persist user", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not persist user"})
		return
	}

	session.Set("profileID", profile.ID.Hex())
	session.Set("githubID", profile.Github.ID)
	if err = session.Save(); err != nil {
		gh.auth.Logger.Errorw("REST - GET - GitHubCallback - failed to save profile in session", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not persist user"})
		return
	}

	// issue the local access JWT and store a hashed refresh token server-side.
	accessToken, refreshToken, expirationTime, err := authpkg.IssueGitHubLoginResult(
		ctx,
		gh.auth.CollRefreshTokens,
		profile,
		gh.auth.JwtKey,
		authpkg.WebTokenTTL,
		authpkg.WebRefreshTokenTTL,
		authpkg.RefreshTokenClientWeb,
	)
	if err != nil {
		gh.auth.Logger.Errorw("REST - GET - GitHubCallback - could not issue web login result", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not complete login"})
		return
	}

	// Send the raw refresh token back only as an HttpOnly cookie.
	utils.SetRefreshTokenCookie(c, refreshToken, authpkg.WebRefreshTokenTTL, os.Getenv("ENV") == "prod")

	gh.auth.Logger.Infow("AUDIT - JWT issued (web)",
		"profileID", profile.ID.Hex(),
		"expiry", expirationTime,
	)

	// The access token is returned in the fragment because the SPA consumes it after /postlogin.
	location := url.URL{Path: "/postlogin", Fragment: "token=" + accessToken}
	c.Redirect(http.StatusFound, location.String())
}
