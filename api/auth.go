package api

import (
	"api-server/db"
	"api-server/models"
	"api-server/utils"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"
)

const webTokenTTL = 30 * time.Minute              // 30 minutes
const mobileTokenTTL = 30 * time.Minute           // 30 minutes
const webRefreshTokenTTL = 7 * 24 * time.Hour     // 7 days
const mobileRefreshTokenTTL = 90 * 24 * time.Hour // 3 months

const refreshTokenCookieName = "refresh_token"
const refreshTokenCookiePath = "/api/token/refresh"

// Auth handles JWT token issuance and validation.
type Auth struct {
	logger        *zap.SugaredLogger
	jwtKey        []byte
	jwtRefreshKey []byte
	collProfiles  *mongo.Collection
}

// NewAuth constructs an Auth using the JWT keys from the environment.
func NewAuth(logger *zap.SugaredLogger, client *mongo.Client) *Auth {
	return &Auth{
		logger:        logger,
		jwtKey:        []byte(os.Getenv("JWT_PASSWORD")),
		jwtRefreshKey: []byte(os.Getenv("JWT_REFRESH_PASSWORD")),
		collProfiles:  db.GetCollections(client).Profiles,
	}
}

// LoginCallback function
func (a *Auth) LoginCallback(c *gin.Context) {
	a.logger.Info("REST - GET - LoginCallback called")

	profile, ok := c.Value("profile").(models.Profile)
	if !ok {
		a.logger.Error("REST - GET - LoginCallback - profile not found in context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "profile not found in context"})
		return
	}
	expirationTime := time.Now().Add(webTokenTTL)

	tokenString, err := utils.CreateJWT(profile, expirationTime, utils.AccessToken, jwt.SigningMethodHS256, a.jwtKey)
	if err != nil {
		a.logger.Error("REST - GET - LoginCallback - cannot generate access JWT")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot generate JWT"})
		return
	}

	// Issue refresh token as HttpOnly cookie
	_, err = a.setRefreshTokenCookie(c, profile, webRefreshTokenTTL)
	if err != nil {
		a.logger.Error("REST - GET - LoginCallback - cannot generate refresh JWT")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot generate refresh token"})
		return
	}

	a.logger.Infow("AUDIT - JWT issued (web)",
		"profileID", profile.ID.Hex(),
		"clientIP", c.ClientIP(),
		"expiry", expirationTime,
	)

	// Redirect with the token as a URL fragment so it is never sent to the server
	// in subsequent requests and does not appear in access logs.
	location := url.URL{Path: "/postlogin", Fragment: "token=" + tokenString}

	c.Redirect(http.StatusFound, location.String())
}

// LoginMobileAppCallback function
func (a *Auth) LoginMobileAppCallback(c *gin.Context) {
	a.logger.Info("REST - GET - LoginMobileAppCallback called")

	profile, ok := c.Value("profile").(models.Profile)
	if !ok {
		a.logger.Error("REST - GET - LoginMobileAppCallback - profile not found in context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "profile not found in context"})
		return
	}
	expirationTime := time.Now().Add(mobileTokenTTL)

	tokenString, err := utils.CreateJWT(profile, expirationTime, utils.AccessToken, jwt.SigningMethodHS256, a.jwtKey)
	if err != nil {
		a.logger.Error("REST - GET - LoginMobileAppCallback - cannot generate access JWT")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot generate JWT"})
		return
	}

	// Issue refresh token as HttpOnly cookie; also capture the value to include in the
	// deep-link so the Android app can store it (the browser Intent system does not expose
	// Set-Cookie headers to the app).
	refreshTokenString, err := a.setRefreshTokenCookie(c, profile, mobileRefreshTokenTTL)
	if err != nil {
		a.logger.Error("REST - GET - LoginMobileAppCallback - cannot generate refresh JWT")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot generate refresh token"})
		return
	}

	a.logger.Infow("AUDIT - JWT issued (mobile)",
		"profileID", profile.ID.Hex(),
		"clientIP", c.ClientIP(),
		"expiry", expirationTime,
	)

	// Prefer the session cookie written to the response by OauthAuth (when the profile was
	// freshly loaded and session.Save() was called — e.g. on first install). Fall back to
	// the request cookie when the profile was already in the session and Save() was not called.
	cookieValue := ""
	for _, h := range c.Writer.Header()["Set-Cookie"] {
		if strings.HasPrefix(h, "mysession=") {
			cookieValue = strings.SplitN(strings.TrimPrefix(h, "mysession="), ";", 2)[0]
			break
		}
	}
	if cookieValue == "" {
		reqCookie, err := c.Request.Cookie("mysession")
		if err != nil {
			a.logger.Error("REST - GET - LoginMobileAppCallback - cannot get session cookie from request")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot get session cookie"})
			return
		}
		cookieValue = reqCookie.Value
	}

	queryParams := url.Values{}
	queryParams.Set("session_cookie", cookieValue)
	queryParams.Set("token", tokenString)
	queryParams.Set("refresh_token", refreshTokenString)
	location := url.URL{Path: "homeanthill://homeanthill.eu/postlogin", RawQuery: queryParams.Encode()}

	c.Redirect(http.StatusFound, location.RequestURI())
}

// JWTMiddleware function
func (a *Auth) JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		const bearerPrefix = "Bearer "
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			a.logger.Error("JWTMiddleware - authorization header not found")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "authorization header not found",
			})
			c.Abort()
			return
		}

		if !strings.HasPrefix(authHeader, bearerPrefix) {
			a.logger.Error("JWTMiddleware - bearer token not found")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "bearer token not found",
			})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, bearerPrefix)

		if tokenString == "" {
			a.logger.Error("JWTMiddleware - bearer token not found")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "bearer token not found",
			})
			return
		}

		claimsObj := &utils.JWTClaims{}

		// Parse takes the token string and a function for looking up the key. The latter is especially
		// useful if you use multiple keys for your application.  The standard is to use 'kid' in the
		// head of the token to identify which key to use, but the parsed token (head and claims) is provided
		// to the callback, providing flexibility.
		token, err := jwt.ParseWithClaims(tokenString, claimsObj, func(token *jwt.Token) (interface{}, error) {
			// Don't forget to validate the alg is what you expect:
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			// jwtKey is injected in Auth struct
			return a.jwtKey, nil
		}, jwt.WithIssuer(utils.JWTIssuer), jwt.WithAudience(utils.JWTAudience))

		if token == nil || !token.Valid || err != nil {
			if errors.Is(err, jwt.ErrTokenMalformed) {
				a.logger.Errorw("JWTMiddleware - token validation failed", "error", err)
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "that's not even a token",
				})
				c.Abort()
				return
			} else if errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet) {
				// Token is either expired or not active yet
				a.logger.Errorw("JWTMiddleware - token validation failed", "error", err)
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "token is expired",
				})
				c.Abort()
				return
			}

			a.logger.Error("JWTMiddleware - not logged, token is not valid")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "not logged, token is not valid"})
			c.Abort()
			return
		}

		// Reject refresh tokens used as access tokens
		if claimsObj.TokenType == utils.RefreshToken {
			a.logger.Error("JWTMiddleware - refresh token cannot be used as access token")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token cannot be used as access token"})
			c.Abort()
			return
		}

		c.Set("jwt_claims", claimsObj)
		c.Next()
	}
}

// setRefreshTokenCookie creates a refresh JWT, sets it as an HttpOnly cookie, and returns the
// raw token string. The string is needed by LoginMobileAppCallback to embed the token in the
// deep-link URL, because the Android Intent system does not expose Set-Cookie headers.
func (a *Auth) setRefreshTokenCookie(c *gin.Context, profile models.Profile, ttl time.Duration) (string, error) {
	expirationTime := time.Now().Add(ttl)
	refreshTokenString, err := utils.CreateJWT(profile, expirationTime, utils.RefreshToken, jwt.SigningMethodHS256, a.jwtRefreshKey)
	if err != nil {
		return "", err
	}
	secure := os.Getenv("ENV") == "prod"
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		refreshTokenCookieName,
		refreshTokenString,
		int(ttl.Seconds()),
		refreshTokenCookiePath,
		"", // domain: let the browser default to the request host
		secure,
		true, // httpOnly
	)
	return refreshTokenString, nil
}

// RefreshToken reads the refresh token from the cookie, validates it, and issues a new access token.
func (a *Auth) RefreshToken(c *gin.Context) {
	a.logger.Info("REST - POST - RefreshToken called")

	refreshCookie, err := c.Cookie(refreshTokenCookieName)
	if err != nil {
		a.logger.Error("REST - POST - RefreshToken - refresh token cookie not found")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token not found"})
		return
	}

	claimsObj := &utils.JWTClaims{}
	token, err := jwt.ParseWithClaims(refreshCookie, claimsObj, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return a.jwtRefreshKey, nil
	}, jwt.WithIssuer(utils.JWTIssuer), jwt.WithAudience(utils.JWTAudience))

	if token == nil || !token.Valid || err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			a.logger.Error("REST - POST - RefreshToken - refresh token expired")
			// Clear the expired cookie
			a.clearRefreshTokenCookie(c)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token expired"})
			return
		}
		a.logger.Errorw("REST - POST - RefreshToken - invalid refresh token", "error", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}

	// Ensure this is actually a refresh token
	if claimsObj.TokenType != utils.RefreshToken {
		a.logger.Error("REST - POST - RefreshToken - token is not a refresh token")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token type"})
		return
	}

	// Look up profile in DB by GitHub ID from claims
	var profile models.Profile
	err = a.collProfiles.FindOne(c.Request.Context(), bson.M{
		"github.id": claimsObj.ID,
	}).Decode(&profile)
	if err != nil {
		a.logger.Errorw("REST - POST - RefreshToken - profile not found", "githubID", claimsObj.ID, "error", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "profile not found"})
		return
	}

	// Issue a new access token
	expirationTime := time.Now().Add(webTokenTTL)
	accessTokenString, err := utils.CreateJWT(profile, expirationTime, utils.AccessToken, jwt.SigningMethodHS256, a.jwtKey)
	if err != nil {
		a.logger.Error("REST - POST - RefreshToken - cannot generate access JWT")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot generate access token"})
		return
	}

	// Rotate the refresh token so each use yields a fresh cookie
	_, err = a.setRefreshTokenCookie(c, profile, webRefreshTokenTTL)
	if err != nil {
		a.logger.Error("REST - POST - RefreshToken - cannot rotate refresh JWT")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot rotate refresh token"})
		return
	}

	a.logger.Infow("AUDIT - access token refreshed",
		"profileID", profile.ID.Hex(),
		"clientIP", c.ClientIP(),
		"expiry", expirationTime,
	)

	c.JSON(http.StatusOK, gin.H{"token": accessTokenString})
}

// clearRefreshTokenCookie removes the refresh token cookie.
func (a *Auth) clearRefreshTokenCookie(c *gin.Context) {
	secure := os.Getenv("ENV") == "prod"
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		refreshTokenCookieName,
		"",
		-1,
		refreshTokenCookiePath,
		"",
		secure,
		true,
	)
}
