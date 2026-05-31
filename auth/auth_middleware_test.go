package auth

import (
	"api-server/models"
	"api-server/utils"
	"crypto/sha256"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.uber.org/zap"
)

func TestJWTMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtKey := []byte("0123456789abcdef0123456789abcdef")
	profile := models.Profile{
		ID: bson.NewObjectID(),
		Github: models.GitHub{
			ID:   123,
			Name: "Octo Cat",
		},
	}

	validWebToken := mustCreateJWT(t, profile, time.Now().Add(time.Hour), utils.AccessToken, RefreshTokenClientWeb, jwtKey)
	validMobileToken := mustCreateJWT(t, profile, time.Now().Add(time.Hour), utils.AccessToken, RefreshTokenClientMobile, jwtKey)
	expiredToken := mustCreateJWT(t, profile, time.Now().Add(-time.Hour), utils.AccessToken, RefreshTokenClientWeb, jwtKey)
	refreshToken := mustCreateJWT(t, profile, time.Now().Add(time.Hour), utils.RefreshToken, RefreshTokenClientWeb, jwtKey)
	wrongKeyToken := mustCreateJWT(t, profile, time.Now().Add(time.Hour), utils.AccessToken, RefreshTokenClientWeb, []byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"))
	wrongAlgToken := mustCreateJWTWithMethod(t, profile, time.Now().Add(time.Hour), jwt.SigningMethodHS256, jwtKey)
	notYetValidToken := mustCreateNotBeforeJWT(t, profile, time.Now().Add(time.Hour), time.Now().Add(time.Hour), jwtKey)
	wrongIssuerToken := mustCreateJWTWithIssuer(t, profile, time.Now().Add(time.Hour), jwtKey, "other-issuer")

	tests := []struct {
		name          string
		authHeader    string
		session       *models.Profile
		wantStatus    int
		wantError     string
		wantNext      bool
		wantClaimsSet bool
	}{
		{name: "missing authorization header", wantStatus: http.StatusUnauthorized, wantError: "authorization header not found"},
		{name: "missing bearer prefix", authHeader: validWebToken, wantStatus: http.StatusUnauthorized, wantError: "bearer token not found"},
		{name: "empty bearer token", authHeader: "Bearer ", wantStatus: http.StatusUnauthorized, wantError: "bearer token not found"},
		{name: "malformed token", authHeader: "Bearer bad.jwt.token", wantStatus: http.StatusBadRequest, wantError: "that's not even a token"},
		{name: "expired token", authHeader: "Bearer " + expiredToken, session: &profile, wantStatus: http.StatusUnauthorized, wantError: "token is expired"},
		{name: "not yet valid token", authHeader: "Bearer " + notYetValidToken, session: &profile, wantStatus: http.StatusUnauthorized, wantError: "token is expired"},
		{name: "wrong signing method", authHeader: "Bearer " + wrongAlgToken, session: &profile, wantStatus: http.StatusUnauthorized, wantError: "not logged, token is not valid"},
		{name: "wrong signing key", authHeader: "Bearer " + wrongKeyToken, session: &profile, wantStatus: http.StatusUnauthorized, wantError: "not logged, token is not valid"},
		{name: "wrong issuer", authHeader: "Bearer " + wrongIssuerToken, session: &profile, wantStatus: http.StatusUnauthorized, wantError: "not logged, token is not valid"},
		{name: "refresh token rejected", authHeader: "Bearer " + refreshToken, session: &profile, wantStatus: http.StatusUnauthorized, wantError: "token is not an access token"},
		{name: "web token missing session", authHeader: "Bearer " + validWebToken, wantStatus: http.StatusUnauthorized, wantError: "cannot find profile in session"},
		{name: "web token mismatched session", authHeader: "Bearer " + validWebToken, session: &models.Profile{ID: bson.NewObjectID(), Github: models.GitHub{ID: 456}}, wantStatus: http.StatusUnauthorized, wantError: "session does not match token identity"},
		{name: "web token accepted", authHeader: "Bearer " + validWebToken, session: &profile, wantStatus: http.StatusOK, wantNext: true, wantClaimsSet: true},
		{name: "mobile token accepted without session", authHeader: "Bearer " + validMobileToken, wantStatus: http.StatusOK, wantNext: true, wantClaimsSet: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newJWTMiddlewareTestRouter(jwtKey, tt.session)
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			router.ServeHTTP(recorder, req)

			if recorder.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", recorder.Code, tt.wantStatus, recorder.Body.String())
			}
			if tt.wantError != "" {
				var body map[string]string
				if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
					t.Fatalf("Unmarshal body returned error: %v", err)
				}
				if body["error"] != tt.wantError {
					t.Fatalf("error = %q, want %q", body["error"], tt.wantError)
				}
			}
			if tt.wantNext && strings.TrimSpace(recorder.Body.String()) != `{"ok":true}` {
				t.Fatalf("body = %s, want ok body", recorder.Body.String())
			}
			if tt.wantClaimsSet && recorder.Header().Get("X-Claims-Set") != "true" {
				t.Fatalf("X-Claims-Set = %q, want true", recorder.Header().Get("X-Claims-Set"))
			}
		})
	}
}

func newJWTMiddlewareTestRouter(jwtKey []byte, sessionProfile *models.Profile) *gin.Engine {
	storeSecret := []byte("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
	blockKey := sha256.Sum256(storeSecret)
	store := cookie.NewStore(storeSecret, blockKey[:])
	store.Options(sessions.Options{Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode})

	auth := &Auth{
		Logger: zap.NewNop().Sugar(),
		JwtKey: jwtKey,
	}
	router := gin.New()
	router.Use(sessions.Sessions(utils.SessionName, store))
	if sessionProfile != nil {
		router.Use(func(c *gin.Context) {
			session := sessions.Default(c)
			session.Set("profileID", sessionProfile.ID.Hex())
			session.Set("githubID", sessionProfile.Github.ID)
			if err := session.Save(); err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.Next()
		})
	}
	router.Use(auth.JWTMiddleware())
	router.GET("/protected", func(c *gin.Context) {
		if _, exists := c.Get("jwt_claims"); exists {
			c.Header("X-Claims-Set", "true")
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	return router
}

func mustCreateJWT(t *testing.T, profile models.Profile, exp time.Time, tokenType utils.TokenType, clientType string, key []byte) string {
	t.Helper()
	token, err := utils.CreateJWT(profile, exp, tokenType, clientType, key)
	if err != nil {
		t.Fatalf("CreateJWT returned error: %v", err)
	}
	return token
}

func mustCreateJWTWithMethod(t *testing.T, profile models.Profile, exp time.Time, method jwt.SigningMethod, key []byte) string {
	t.Helper()
	now := time.Now().UTC()
	claims := &utils.JWTClaims{
		ID:         profile.Github.ID,
		ProfileID:  profile.ID.Hex(),
		Name:       profile.Github.Name,
		TokenType:  utils.AccessToken,
		ClientType: RefreshTokenClientWeb,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    utils.JWTIssuer,
			Audience:  jwt.ClaimStrings{utils.JWTAudience},
			Subject:   "123",
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}
	token, err := jwt.NewWithClaims(method, claims).SignedString(key)
	if err != nil {
		t.Fatalf("SignedString returned error: %v", err)
	}
	return token
}

func mustCreateNotBeforeJWT(t *testing.T, profile models.Profile, exp, notBefore time.Time, key []byte) string {
	t.Helper()
	claims := &utils.JWTClaims{
		ID:         profile.Github.ID,
		ProfileID:  profile.ID.Hex(),
		Name:       profile.Github.Name,
		TokenType:  utils.AccessToken,
		ClientType: RefreshTokenClientWeb,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    utils.JWTIssuer,
			Audience:  jwt.ClaimStrings{utils.JWTAudience},
			Subject:   "123",
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			NotBefore: jwt.NewNumericDate(notBefore),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS512, claims).SignedString(key)
	if err != nil {
		t.Fatalf("SignedString returned error: %v", err)
	}
	return token
}

func mustCreateJWTWithIssuer(t *testing.T, profile models.Profile, exp time.Time, key []byte, issuer string) string {
	t.Helper()
	now := time.Now().UTC()
	claims := &utils.JWTClaims{
		ID:         profile.Github.ID,
		ProfileID:  profile.ID.Hex(),
		Name:       profile.Github.Name,
		TokenType:  utils.AccessToken,
		ClientType: RefreshTokenClientWeb,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Audience:  jwt.ClaimStrings{utils.JWTAudience},
			Subject:   "123",
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS512, claims).SignedString(key)
	if err != nil {
		t.Fatalf("SignedString returned error: %v", err)
	}
	return token
}
