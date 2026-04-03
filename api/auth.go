package api

import (
	"api-server/models"
	"api-server/utils"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

const webTokenTTL = 60 * time.Minute
const mobileTokenTTL = 6 * 30 * 24 * time.Hour

// Auth handles JWT token issuance and validation.
type Auth struct {
	logger *zap.SugaredLogger
	jwtKey []byte
}

// NewAuth constructs an Auth using the JWT key from the environment.
func NewAuth(logger *zap.SugaredLogger) *Auth {
	return &Auth{
		logger: logger,
		jwtKey: []byte(os.Getenv("JWT_PASSWORD")),
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

	tokenString, err := utils.CreateJWT(profile, expirationTime, jwt.SigningMethodHS256, a.jwtKey)
	if err != nil {
		a.logger.Error("REST - GET - LoginCallback - cannot generate JWT")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot generate JWT"})
		return
	}

	a.logger.Infow("AUDIT - JWT issued (web)",
		"profileID", profile.ID.Hex(),
		"clientIP", c.ClientIP(),
		"expiry", expirationTime,
	)

	queryParams := url.Values{}
	queryParams.Set("token", tokenString)
	location := url.URL{Path: "/postlogin", RawQuery: queryParams.Encode()}

	c.Redirect(http.StatusFound, location.RequestURI())
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

	tokenString, err := utils.CreateJWT(profile, expirationTime, jwt.SigningMethodHS256, a.jwtKey)
	if err != nil {
		a.logger.Error("REST - GET - LoginMobileAppCallback - cannot generate JWT")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot generate JWT"})
		return
	}

	a.logger.Infow("AUDIT - JWT issued (mobile)",
		"profileID", profile.ID.Hex(),
		"clientIP", c.ClientIP(),
		"expiry", expirationTime,
	)

	cookie, err := c.Request.Cookie("mysession")
	if err != nil {
		a.logger.Error("REST - GET - LoginMobileAppCallback - cannot get session cookie from request")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot get session cookie"})
		return
	}

	queryParams := url.Values{}
	queryParams.Set("session_cookie", cookie.Value)
	queryParams.Set("token", tokenString)
	location := url.URL{Path: "homeanthill://homeanthill.eu/postlogin", RawQuery: queryParams.Encode()}

	c.Redirect(http.StatusFound, location.RequestURI())
}

// JWTMiddleware function
func (a *Auth) JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		const BearerSchema = "Bearer"
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			a.logger.Error("JWTMiddleware - authorization header not found")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "authorization header not found",
			})
			c.Abort()
			return
		}

		tokenString := authHeader[(len(BearerSchema) + 1):]

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
		})

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
			} else {
				a.logger.Error("JWTMiddleware - not logged, token is not valid")
				c.JSON(http.StatusUnauthorized, gin.H{"error": "not logged, token is not valid"})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
