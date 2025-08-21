package api

import (
	"api-server/models"
	"api-server/utils"
	"context"
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

// Auth struct
type Auth struct {
	ctx    context.Context
	logger *zap.SugaredLogger
}

// NewAuth function
func NewAuth(ctx context.Context, logger *zap.SugaredLogger) *Auth {
	return &Auth{
		ctx:    ctx,
		logger: logger,
	}
}

// LoginCallback function
func (handler *Auth) LoginCallback(c *gin.Context) {
	handler.logger.Info("REST - GET - LoginCallback called")
	// jwtKey is a []byte containing your secret, e.g. []byte("my_secret_key")
	var jwtKey = []byte(os.Getenv("JWT_PASSWORD"))

	profile := c.Value("profile").(models.Profile)
	expirationTime := time.Now().Add(60 * time.Minute)

	tokenString, err := utils.CreateJWT(profile, expirationTime, jwt.SigningMethodHS256, jwtKey)
	if err != nil {
		handler.logger.Error("REST - GET - LoginCallback - cannot generate JWT")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot generate JWT"})
		return
	}

	queryParams := url.Values{}
	queryParams.Set("token", tokenString)
	location := url.URL{Path: "/postlogin", RawQuery: queryParams.Encode()}

	c.Redirect(http.StatusFound, location.RequestURI())
}

// LoginMobileAppCallback function
func (handler *Auth) LoginMobileAppCallback(c *gin.Context) {
	handler.logger.Info("REST - GET - LoginMobileAppCallback called")
	// jwtKey is a []byte containing your secret, e.g. []byte("my_secret_key")
	var jwtKey = []byte(os.Getenv("JWT_PASSWORD"))

	profile := c.Value("profile").(models.Profile)
	expirationTime := time.Now().Add((60 * time.Minute) * 24 * 30 * 6) // 6 months

	tokenString, err := utils.CreateJWT(profile, expirationTime, jwt.SigningMethodHS256, jwtKey)
	if err != nil {
		handler.logger.Error("REST - GET - LoginMobileAppCallback - cannot generate JWT")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot generate JWT"})
		return
	}

	cookie, err := c.Request.Cookie("mysession")
	if err != nil {
		handler.logger.Error("REST - GET - LoginMobileAppCallback - cannot get session cookie from request")
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
func (handler *Auth) JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		const BearerSchema = "Bearer"
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			handler.logger.Error("JWTMiddleware - authorization header not found")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "authorization header not found",
			})
			c.Abort()
			return
		}

		tokenString := authHeader[(len(BearerSchema) + 1):]

		if tokenString == "" {
			handler.logger.Error("JWTMiddleware - bearer token not found")
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
			// jwtKey is a []byte containing your secret, e.g. []byte("my_secret_key")
			var jwtKey = []byte(os.Getenv("JWT_PASSWORD"))
			return jwtKey, nil
		})

		if token == nil || !token.Valid || err != nil {
			if errors.Is(err, jwt.ErrTokenMalformed) {
				handler.logger.Error("JWTMiddleware - " + err.Error())
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "that's not even a token",
				})
				c.Abort()
				return
			} else if errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet) {
				// Token is either expired or not active yet
				handler.logger.Error("JWTMiddleware - " + err.Error())
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "token is expired",
				})
				c.Abort()
				return
			} else {
				handler.logger.Error("JWTMiddleware - not logged, token is not valid")
				c.JSON(http.StatusUnauthorized, gin.H{"error": "not logged, token is not valid"})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
