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
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/writeconcern"
	"go.uber.org/zap"
)

type OAuthHandler struct {
	logger            *zap.SugaredLogger
	client            *mongo.Client
	jwtKey            []byte
	collProfiles      *mongo.Collection
	collRefreshTokens *mongo.Collection
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
	errRefreshTokenProfileNotFound = errors.New("refresh token profile not found")
)

func NewOAuthHandler(logger *zap.SugaredLogger, client *mongo.Client) *OAuthHandler {
	colls := db.GetCollections(client)
	return &OAuthHandler{
		logger:            logger,
		client:            client,
		jwtKey:            []byte(os.Getenv("JWT_PASSWORD")),
		collProfiles:      colls.Profiles,
		collRefreshTokens: colls.RefreshTokens,
	}
}

// RefreshToken reads the refresh token from the cookie, validates it, and issues a new access token.
func (oc *OAuthHandler) RefreshToken(c *gin.Context) {
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

	tokenRecord, profile, err := oc.validateRefreshToken(ctx, rawRefreshToken, authpkg.RefreshTokenClientWeb)
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

	now := time.Now().UTC()
	expirationTime := now.Add(authpkg.WebTokenTTL)
	accessTokenString, err := utils.CreateJWT(profile, expirationTime, utils.AccessToken, oc.jwtKey)
	if err != nil {
		oc.logger.Error("REST - POST - RefreshToken - cannot generate access JWT")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot generate access token"})
		return
	}

	// regenerate refreshToken
	newRefreshToken, err := oc.rotateStoredRefreshToken(ctx, tokenRecord, authpkg.WebRefreshTokenTTL)
	if err != nil {
		if errors.Is(err, errRefreshTokenReuse) {
			utils.ClearRefreshTokenCookie(c, os.Getenv("ENV") == "prod")
			oc.logger.Error("REST - POST - RefreshToken - refresh token reuse detected during rotation")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token reuse detected"})
			return
		}
		oc.logger.Errorw("REST - POST - RefreshToken - cannot rotate refresh token", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot rotate refresh token"})
		return
	}
	utils.SetRefreshTokenCookie(c, newRefreshToken, authpkg.WebRefreshTokenTTL, os.Getenv("ENV") == "prod")

	oc.logger.Infow("AUDIT - access token refreshed",
		"profileID", profile.ID.Hex(),
		"expiry", expirationTime,
	)

	c.JSON(http.StatusOK, gin.H{"token": accessTokenString})
}

// RefreshMobileToken reads the mobile refresh token from JSON, validates it, and issues rotated mobile tokens.
func (oc *OAuthHandler) RefreshMobileToken(c *gin.Context) {
	oc.logger.Info("REST - POST - RefreshMobileToken called")

	var req appRefreshTokenReq
	if err := c.ShouldBindJSON(&req); err != nil {
		oc.logger.Error("REST - POST - RefreshMobileToken - invalid request payload")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	rawRefreshToken := req.RefreshToken
	if rawRefreshToken == "" {
		oc.logger.Error("REST - POST - RefreshMobileToken - refresh token not found")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token not found"})
		return
	}

	// This is not really required because the timeout is already 10s, but in this way it's more explicit.
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	tokenRecord, profile, err := oc.validateRefreshToken(ctx, rawRefreshToken, authpkg.RefreshTokenClientMobile)
	if err != nil {
		switch {
		case errors.Is(err, errRefreshTokenNotFound):
			oc.logger.Error("REST - POST - RefreshMobileToken - invalid refresh token")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
			return
		case errors.Is(err, errRefreshTokenReuse):
			oc.logger.Error("REST - POST - RefreshMobileToken - refresh token reuse detected")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token reuse detected"})
			return
		case errors.Is(err, errRefreshTokenExpired):
			oc.logger.Error("REST - POST - RefreshMobileToken - refresh token expired")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token expired"})
			return
		case errors.Is(err, errRefreshTokenProfileNotFound):
			oc.logger.Errorw("REST - POST - RefreshMobileToken - profile not found", "profileID", tokenRecord.ProfileID.Hex(), "error", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "profile not found"})
			return
		default:
			oc.logger.Errorw("REST - POST - RefreshMobileToken - cannot validate refresh token", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot validate refresh token"})
			return
		}
	}

	now := time.Now().UTC()
	expirationTime := now.Add(authpkg.MobileTokenTTL)
	accessTokenString, err := utils.CreateJWT(profile, expirationTime, utils.AccessToken, oc.jwtKey)
	if err != nil {
		oc.logger.Error("REST - POST - RefreshMobileToken - cannot generate access JWT")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot generate access token"})
		return
	}

	newRefreshToken, err := oc.rotateStoredRefreshToken(ctx, tokenRecord, authpkg.MobileRefreshTokenTTL)
	if err != nil {
		if errors.Is(err, errRefreshTokenReuse) {
			oc.logger.Error("REST - POST - RefreshMobileToken - refresh token reuse detected during rotation")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token reuse detected"})
			return
		}
		oc.logger.Errorw("REST - POST - RefreshMobileToken - cannot rotate refresh token", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot rotate refresh token"})
		return
	}

	oc.logger.Infow("AUDIT - mobile access token refreshed",
		"profileID", profile.ID.Hex(),
		"expiry", expirationTime,
	)

	c.JSON(http.StatusOK, appRefreshTokenResp{
		Token:        accessTokenString,
		RefreshToken: newRefreshToken,
	})
}

func (oc *OAuthHandler) Logout(c *gin.Context) {
	if rawRefreshToken, err := c.Cookie(utils.RefreshTokenCookieName); err == nil && rawRefreshToken != "" {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()
		now := time.Now().UTC()
		if err = oc.revokeRefreshTokenFamilyByHash(ctx, utils.HashToken(rawRefreshToken), authpkg.RefreshTokenClientWeb, now); err != nil {
			oc.logger.Warnw("REST - POST - Logout - cannot revoke refresh token family", "error", err)
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

func (oc *OAuthHandler) LogoutApp(c *gin.Context) {
	var req appRefreshTokenReq
	if err := c.ShouldBindJSON(&req); err != nil {
		oc.logger.Error("REST - POST - LogoutApp - invalid request payload")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	rawRefreshToken := req.RefreshToken
	if rawRefreshToken == "" {
		oc.logger.Error("REST - POST - LogoutApp - refresh token not found")
		c.JSON(http.StatusBadRequest, gin.H{"error": "refresh token not found"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	now := time.Now().UTC()
	if err := oc.revokeRefreshTokenFamilyByHash(ctx, utils.HashToken(rawRefreshToken), authpkg.RefreshTokenClientMobile, now); err != nil {
		oc.logger.Warnw("REST - POST - LogoutApp - cannot revoke refresh token family", "error", err)
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
		oc.logger.Errorw("REST - POST - LogoutApp - cannot clear session", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot logout"})
		return
	}

	c.Status(http.StatusNoContent)
}

func (oc *OAuthHandler) validateRefreshToken(ctx context.Context, rawRefreshToken, clientType string) (models.RefreshToken, models.Profile, error) {
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
		if revokeErr := oc.revokeRefreshTokenFamily(ctx, tokenRecord.FamilyID, now); revokeErr != nil {
			return tokenRecord, models.Profile{}, revokeErr
		}
		return tokenRecord, models.Profile{}, errRefreshTokenReuse
	}

	if now.After(tokenRecord.ExpiresAt) {
		if revokeErr := oc.revokeRefreshTokenByHash(ctx, tokenRecord.TokenHash, now); revokeErr != nil {
			return tokenRecord, models.Profile{}, revokeErr
		}
		return tokenRecord, models.Profile{}, errRefreshTokenExpired
	}

	var profile models.Profile
	err = oc.collProfiles.FindOne(ctx, bson.M{"_id": tokenRecord.ProfileID}).Decode(&profile)
	if err != nil {
		return tokenRecord, models.Profile{}, errRefreshTokenProfileNotFound
	}

	return tokenRecord, profile, nil
}

func (oc *OAuthHandler) rotateStoredRefreshToken(ctx context.Context, tokenRecord models.RefreshToken, ttl time.Duration) (string, error) {
	newRefreshToken, err := utils.RandomString(64)
	if err != nil {
		return "", err
	}

	now := time.Now().UTC()
	newHash := utils.HashToken(newRefreshToken)

	dbSession, err := oc.client.StartSession()
	if err != nil {
		return "", err
	}
	defer dbSession.EndSession(context.Background())

	newTokenRecord := models.RefreshToken{
		ID:         bson.NewObjectID(),
		ProfileID:  tokenRecord.ProfileID,
		TokenHash:  newHash,
		FamilyID:   tokenRecord.FamilyID,
		ClientType: tokenRecord.ClientType,
		CreatedAt:  now,
		ExpiresAt:  now.Add(ttl),
	}

	_, err = dbSession.WithTransaction(ctx, func(sessionCtx context.Context) (interface{}, error) {
		result, updateErr := oc.collRefreshTokens.UpdateOne(
			sessionCtx,
			bson.M{
				"_id":        tokenRecord.ID,
				"tokenHash":  tokenRecord.TokenHash,
				"clientType": tokenRecord.ClientType,
				"revokedAt":  bson.M{"$exists": false},
				"expiresAt":  bson.M{"$gt": now},
			},
			bson.M{
				"$set": bson.M{
					"revokedAt":      now,
					"replacedByHash": newHash,
					"lastUsedAt":     now,
				},
			},
		)
		if updateErr != nil {
			return nil, updateErr
		}
		if result.ModifiedCount != 1 {
			return nil, errRefreshTokenReuse
		}

		if _, err = oc.collRefreshTokens.InsertOne(sessionCtx, newTokenRecord); err != nil {
			return nil, err
		}
		return nil, nil
	}, options.Transaction().SetWriteConcern(writeconcern.Majority()))
	if err != nil {
		if errors.Is(err, errRefreshTokenReuse) {
			if revokeErr := oc.revokeRefreshTokenFamily(ctx, tokenRecord.FamilyID, now); revokeErr != nil {
				return "", revokeErr
			}
			return "", errRefreshTokenReuse
		}
		return "", err
	}

	return newRefreshToken, nil
}

func (oc *OAuthHandler) revokeRefreshTokenFamily(ctx context.Context, familyID string, revokedAt time.Time) error {
	_, err := oc.collRefreshTokens.UpdateMany(ctx,
		bson.M{"familyId": familyID, "revokedAt": bson.M{"$exists": false}},
		bson.M{"$set": bson.M{"revokedAt": revokedAt}},
	)
	return err
}

func (oc *OAuthHandler) revokeRefreshTokenByHash(ctx context.Context, tokenHash string, revokedAt time.Time) error {
	_, err := oc.collRefreshTokens.UpdateOne(ctx,
		bson.M{"tokenHash": tokenHash, "revokedAt": bson.M{"$exists": false}},
		bson.M{"$set": bson.M{"revokedAt": revokedAt}},
	)
	return err
}

func (oc *OAuthHandler) revokeRefreshTokenFamilyByHash(ctx context.Context, tokenHash, clientType string, revokedAt time.Time) error {
	var tokenRecord models.RefreshToken
	err := oc.collRefreshTokens.FindOne(ctx, bson.M{"tokenHash": tokenHash, "clientType": clientType}).Decode(&tokenRecord)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil
		}
		return err
	}
	return oc.revokeRefreshTokenFamily(ctx, tokenRecord.FamilyID, revokedAt)
}
