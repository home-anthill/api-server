package api

import (
	authpkg "api-server/auth"
	"api-server/db"
	"api-server/models"
	"api-server/utils"
	"context"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"
)

type OAuthCommon struct {
	logger            *zap.SugaredLogger
	jwtKey            []byte
	collProfiles      *mongo.Collection
	collRefreshTokens *mongo.Collection
}

// GithubAccessTokenResponse is the response body for the GitHub access token endpoint.
type GithubAccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
	Error       string `json:"error"`
	Description string `json:"error_description"`
	ErrorURI    string `json:"error_uri"`
}

// GithubAppCurrentUserResponse is the response body for the GitHub user endpoint.
type GithubAppCurrentUserResponse struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

type appRefreshTokenReq struct {
	RefreshToken string `json:"refreshToken"`
}

type appRefreshTokenResp struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken"`
}

var (
	errRefreshTokenNotFound        = errors.New("refresh token not found")
	errRefreshTokenReuse           = errors.New("refresh token reuse detected")
	errRefreshTokenExpired         = errors.New("refresh token expired")
	errRefreshTokenContextMismatch = errors.New("refresh token context mismatch")
	errRefreshTokenProfileNotFound = errors.New("refresh token profile not found")
)

func NewOAuthCommon(logger *zap.SugaredLogger, client *mongo.Client) *OAuthCommon {
	colls := db.GetCollections(client)
	return &OAuthCommon{
		logger:            logger,
		jwtKey:            []byte(os.Getenv("JWT_PASSWORD")),
		collProfiles:      colls.Profiles,
		collRefreshTokens: colls.RefreshTokens,
	}
}

// RefreshToken reads the refresh token from the cookie, validates it, and issues a new access token.
func (oc *OAuthCommon) RefreshToken(c *gin.Context) {
	oc.logger.Info("REST - POST - RefreshToken called")

	session := sessions.Default(c)

	rawRefreshToken, err := c.Cookie(utils.RefreshTokenCookieName)
	if err != nil || rawRefreshToken == "" {
		oc.logger.Error("REST - POST - RefreshToken - refresh token cookie not found")
		utils.ClearRefreshTokenCookie(c, os.Getenv("ENV") == "prod")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token not found"})
		return
	}

	// This is not really required because the timeout is already 10s, but in this way it's more explicit.
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	tokenRecord, profile, err := oc.validateRefreshToken(ctx, rawRefreshToken, authpkg.RefreshTokenClientWeb, c.GetHeader("User-Agent"))
	if err != nil {
		utils.ClearRefreshTokenCookie(c, os.Getenv("ENV") == "prod")
		switch {
		case errors.Is(err, errRefreshTokenNotFound):
			oc.logger.Error("REST - POST - RefreshToken - invalid refresh token")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
			return
		case errors.Is(err, errRefreshTokenReuse):
			oc.logger.Error("REST - POST - RefreshToken - refresh token reuse detected")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token reuse detected"})
			return
		case errors.Is(err, errRefreshTokenExpired):
			oc.logger.Error("REST - POST - RefreshToken - refresh token expired")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token expired"})
			return
		case errors.Is(err, errRefreshTokenContextMismatch):
			oc.logger.Error("REST - POST - RefreshToken - refresh token context mismatch")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token context mismatch"})
			return
		case errors.Is(err, errRefreshTokenProfileNotFound):
			oc.logger.Errorw("REST - POST - RefreshToken - profile not found", "profileID", tokenRecord.ProfileID.Hex(), "error", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "profile not found"})
			return
		default:
			oc.logger.Errorw("REST - POST - RefreshToken - cannot validate refresh token", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot validate refresh token"})
			return
		}
	}

	session.Set("profileID", profile.ID.Hex())
	session.Set("githubID", profile.Github.ID)
	if err = session.Save(); err != nil {
		oc.logger.Errorw("REST - POST - RefreshToken - cannot save session", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot refresh session"})
		return
	}

	expirationTime := time.Now().Add(authpkg.WebTokenTTL)
	accessTokenString, err := utils.CreateJWT(profile, expirationTime, utils.AccessToken, jwt.SigningMethodHS512, oc.jwtKey)
	if err != nil {
		oc.logger.Error("REST - POST - RefreshToken - cannot generate access JWT")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot generate access token"})
		return
	}

	// regenerate refreshToken
	newRefreshToken, err := oc.rotateStoredRefreshToken(ctx, tokenRecord, c.GetHeader("User-Agent"), authpkg.WebRefreshTokenTTL)
	if err != nil {
		oc.logger.Errorw("REST - POST - RefreshToken - cannot rotate refresh token", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot rotate refresh token"})
		return
	}
	utils.SetRefreshTokenCookie(c, newRefreshToken, authpkg.WebRefreshTokenTTL, os.Getenv("ENV") == "prod")

	oc.logger.Infow("AUDIT - access token refreshed",
		"profileID", profile.ID.Hex(),
		"clientIP", c.ClientIP(),
		"expiry", expirationTime,
	)

	c.JSON(http.StatusOK, gin.H{"token": accessTokenString})
}

// RefreshAppToken reads the mobile refresh token from JSON, validates it, and issues rotated mobile tokens.
func (oc *OAuthCommon) RefreshAppToken(c *gin.Context) {
	oc.logger.Info("REST - POST - RefreshAppToken called")

	var req appRefreshTokenReq
	if err := c.ShouldBindJSON(&req); err != nil {
		oc.logger.Error("REST - POST - RefreshAppToken - invalid request payload")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	rawRefreshToken := req.RefreshToken
	if rawRefreshToken == "" {
		oc.logger.Error("REST - POST - RefreshAppToken - refresh token not found")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token not found"})
		return
	}

	// This is not really required because the timeout is already 10s, but in this way it's more explicit.
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	tokenRecord, profile, err := oc.validateRefreshToken(ctx, rawRefreshToken, authpkg.RefreshTokenClientMobile, c.GetHeader("User-Agent"))
	if err != nil {
		switch {
		case errors.Is(err, errRefreshTokenNotFound):
			oc.logger.Error("REST - POST - RefreshAppToken - invalid refresh token")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
			return
		case errors.Is(err, errRefreshTokenReuse):
			oc.logger.Error("REST - POST - RefreshAppToken - refresh token reuse detected")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token reuse detected"})
			return
		case errors.Is(err, errRefreshTokenExpired):
			oc.logger.Error("REST - POST - RefreshAppToken - refresh token expired")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token expired"})
			return
		case errors.Is(err, errRefreshTokenContextMismatch):
			oc.logger.Error("REST - POST - RefreshAppToken - refresh token context mismatch")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token context mismatch"})
			return
		case errors.Is(err, errRefreshTokenProfileNotFound):
			oc.logger.Errorw("REST - POST - RefreshAppToken - profile not found", "profileID", tokenRecord.ProfileID.Hex(), "error", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "profile not found"})
			return
		default:
			oc.logger.Errorw("REST - POST - RefreshAppToken - cannot validate refresh token", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot validate refresh token"})
			return
		}
	}

	expirationTime := time.Now().Add(authpkg.MobileTokenTTL)
	accessTokenString, err := utils.CreateJWT(profile, expirationTime, utils.AccessToken, jwt.SigningMethodHS512, oc.jwtKey)
	if err != nil {
		oc.logger.Error("REST - POST - RefreshAppToken - cannot generate access JWT")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot generate access token"})
		return
	}

	newRefreshToken, err := oc.rotateStoredRefreshToken(ctx, tokenRecord, c.GetHeader("User-Agent"), authpkg.MobileRefreshTokenTTL)
	if err != nil {
		oc.logger.Errorw("REST - POST - RefreshAppToken - cannot rotate refresh token", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot rotate refresh token"})
		return
	}

	oc.logger.Infow("AUDIT - mobile access token refreshed",
		"profileID", profile.ID.Hex(),
		"clientIP", c.ClientIP(),
		"expiry", expirationTime,
	)

	c.JSON(http.StatusOK, appRefreshTokenResp{
		Token:        accessTokenString,
		RefreshToken: newRefreshToken,
	})
}

func (oc *OAuthCommon) Logout(c *gin.Context) {
	if rawRefreshToken, err := c.Cookie(utils.RefreshTokenCookieName); err == nil && rawRefreshToken != "" {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()
		if err = oc.revokeRefreshTokenByHash(ctx, utils.HashToken(rawRefreshToken), time.Now().UTC()); err != nil {
			oc.logger.Warnw("REST - POST - Logout - cannot revoke refresh token", "error", err)
		}
	}

	session := sessions.Default(c)
	session.Clear()
	session.Options(sessions.Options{
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   os.Getenv("ENV") == "prod",
		SameSite: http.SameSiteLaxMode,
	})
	if err := session.Save(); err != nil {
		oc.logger.Errorw("REST - POST - Logout - cannot clear session", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot logout"})
		return
	}

	utils.ClearRefreshTokenCookie(c, os.Getenv("ENV") == "prod")
	c.Status(http.StatusNoContent)
}

func (oc *OAuthCommon) validateRefreshToken(ctx context.Context, rawRefreshToken, clientType, userAgent string) (models.RefreshToken, models.Profile, error) {
	tokenHash := utils.HashToken(rawRefreshToken)
	var tokenRecord models.RefreshToken
	err := oc.collRefreshTokens.FindOne(ctx, bson.M{"tokenHash": tokenHash, "clientType": clientType}).Decode(&tokenRecord)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return models.RefreshToken{}, models.Profile{}, errRefreshTokenNotFound
		}
		return models.RefreshToken{}, models.Profile{}, err
	}

	now := time.Now().UTC()
	if tokenRecord.RevokedAt != nil {
		_ = oc.revokeRefreshTokenFamily(ctx, tokenRecord.FamilyID, now)
		return tokenRecord, models.Profile{}, errRefreshTokenReuse
	}

	if now.After(tokenRecord.ExpiresAt) {
		_ = oc.revokeRefreshTokenByHash(ctx, tokenRecord.TokenHash, now)
		return tokenRecord, models.Profile{}, errRefreshTokenExpired
	}

	if tokenRecord.UserAgent != "" && tokenRecord.UserAgent != utils.TruncateString(userAgent, 512) {
		_ = oc.revokeRefreshTokenFamily(ctx, tokenRecord.FamilyID, now)
		return tokenRecord, models.Profile{}, errRefreshTokenContextMismatch
	}

	var profile models.Profile
	err = oc.collProfiles.FindOne(ctx, bson.M{"_id": tokenRecord.ProfileID}).Decode(&profile)
	if err != nil {
		return tokenRecord, models.Profile{}, errRefreshTokenProfileNotFound
	}

	return tokenRecord, profile, nil
}

func (oc *OAuthCommon) rotateStoredRefreshToken(ctx context.Context, tokenRecord models.RefreshToken, userAgent string, ttl time.Duration) (string, error) {
	newRefreshToken, err := utils.RandomString(64)
	if err != nil {
		return "", err
	}

	now := time.Now().UTC()
	newHash := utils.HashToken(newRefreshToken)
	_, err = oc.collRefreshTokens.UpdateByID(ctx, tokenRecord.ID, bson.M{
		"$set": bson.M{
			"revokedAt":      now,
			"replacedByHash": newHash,
			"lastUsedAt":     now,
		},
	})
	if err != nil {
		return "", err
	}

	newTokenRecord := models.RefreshToken{
		ID:         bson.NewObjectID(),
		ProfileID:  tokenRecord.ProfileID,
		TokenHash:  newHash,
		FamilyID:   tokenRecord.FamilyID,
		ClientType: tokenRecord.ClientType,
		UserAgent:  utils.TruncateString(userAgent, 512),
		CreatedAt:  now,
		ExpiresAt:  now.Add(ttl),
	}
	if _, err = oc.collRefreshTokens.InsertOne(ctx, newTokenRecord); err != nil {
		return "", err
	}
	return newRefreshToken, nil
}

func (oc *OAuthCommon) revokeRefreshTokenFamily(ctx context.Context, familyID string, revokedAt time.Time) error {
	_, err := oc.collRefreshTokens.UpdateMany(ctx,
		bson.M{"familyId": familyID, "revokedAt": bson.M{"$exists": false}},
		bson.M{"$set": bson.M{"revokedAt": revokedAt}},
	)
	return err
}

func (oc *OAuthCommon) revokeRefreshTokenByHash(ctx context.Context, tokenHash string, revokedAt time.Time) error {
	_, err := oc.collRefreshTokens.UpdateOne(ctx,
		bson.M{"tokenHash": tokenHash, "revokedAt": bson.M{"$exists": false}},
		bson.M{"$set": bson.M{"revokedAt": revokedAt}},
	)
	return err
}
