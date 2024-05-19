package api

import (
	"api-server/models"
	"api-server/utils"
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"net/http"
	"net/url"
	"os"
	"time"
)

type Auth struct {
	ctx    context.Context
	logger *zap.SugaredLogger
}

// For HMAC signing method, the key can be any []byte. It is recommended to generate
// a key using crypto/rand or something equivalent. You need the same key for signing
// and validating.
var jwtKey = []byte(os.Getenv("JWT_PASSWORD"))

func NewAuth(ctx context.Context, logger *zap.SugaredLogger) *Auth {
	return &Auth{
		ctx:    ctx,
		logger: logger,
	}
}

func (handler *Auth) LoginCallback(c *gin.Context) {
	handler.logger.Info("REST - GET - LoginCallback called")

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
	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

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
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "bearer token not found",
			})
			c.Abort()
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
