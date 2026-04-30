package auth

import (
	"api-server/db"
	"api-server/utils"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"
)

// web app
const WebTokenTTL = 15 * time.Minute
const WebRefreshTokenTTL = 7 * 24 * time.Hour

// mobile app
const MobileTokenTTL = 15 * time.Minute
const MobileRefreshTokenTTL = 7 * 24 * time.Hour
const MobileAppLoginCodeTTL = 1 * time.Minute

// Auth handles JWT token issuance and validation.
type Auth struct {
	Logger            *zap.SugaredLogger
	JwtKey            []byte
	JwtRefreshKey     []byte
	CollProfiles      *mongo.Collection
	CollAppLoginCodes *mongo.Collection
	CollRefreshTokens *mongo.Collection
}

// NewAuth constructs an Auth using the JWT keys from the environment.
func NewAuth(logger *zap.SugaredLogger, client *mongo.Client) *Auth {
	colls := db.GetCollections(client)
	return &Auth{
		Logger:            logger,
		JwtKey:            []byte(os.Getenv("JWT_PASSWORD")),
		JwtRefreshKey:     []byte(os.Getenv("JWT_REFRESH_PASSWORD")),
		CollProfiles:      colls.Profiles,
		CollAppLoginCodes: colls.AppLoginCodes,
		CollRefreshTokens: colls.RefreshTokens,
	}
}

// JWTMiddleware function
func (a *Auth) JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		const bearerPrefix = "Bearer "
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			a.Logger.Error("JWTMiddleware - authorization header not found")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "authorization header not found",
			})
			c.Abort()
			return
		}

		if !strings.HasPrefix(authHeader, bearerPrefix) {
			a.Logger.Error("JWTMiddleware - bearer token not found")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "bearer token not found",
			})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, bearerPrefix)

		if tokenString == "" {
			a.Logger.Error("JWTMiddleware - bearer token not found")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "bearer token not found",
			})
			return
		}

		claimsObj := &utils.JWTClaims{}

		// Parse takes the token string and a function for looking up the key. The latter is especially
		// useful if you use multiple keys for your application. The standard is to use 'kid' in the
		// head of the token to identify which key to use, but the parsed token (head and claims) is provided
		// to the callback, providing flexibility.
		token, err := jwt.ParseWithClaims(tokenString, claimsObj, func(token *jwt.Token) (interface{}, error) {
			if token.Method != jwt.SigningMethodHS512 {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			// jwtKey is injected in Auth struct
			return a.JwtKey, nil
		}, jwt.WithIssuer(utils.JWTIssuer), jwt.WithAudience(utils.JWTAudience))

		if token == nil || !token.Valid || err != nil {
			if errors.Is(err, jwt.ErrTokenMalformed) {
				a.Logger.Errorw("JWTMiddleware - token validation failed", "error", err)
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "that's not even a token",
				})
				c.Abort()
				return
			} else if errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet) {
				// Token is either expired or not active yet
				a.Logger.Errorw("JWTMiddleware - token validation failed", "error", err)
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "token is expired",
				})
				c.Abort()
				return
			}

			a.Logger.Error("JWTMiddleware - not logged, token is not valid")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "not logged, token is not valid"})
			c.Abort()
			return
		}

		// Reject non access tokens used as access tokens
		if claimsObj.TokenType != utils.AccessToken {
			a.Logger.Errorw("JWTMiddleware - token is not an access token", "tokenType", claimsObj.TokenType)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token is not an access token"})
			c.Abort()
			return
		}

		// TODO The cleaner final design is to remove this bridge by refactoring handlers to read identity
		//  from c.Get("jwt_claims") or a helper like utils.GetLoggedProfileFromContext(c, collProfiles).
		//  Then mobile auth would be truly session-free end-to-end, while web
		//  could still keep the session-check-in middleware.
		if claimsObj.ClientType == RefreshTokenClientMobile {
			session := sessions.Default(c)
			session.Set("profileID", claimsObj.ProfileID)
			session.Set("githubID", claimsObj.ID)
			c.Set("jwt_claims", claimsObj)
			c.Next()
			return
		}

		// Private handlers still rely on the session profile. Enforce that the
		// session identity matches the already-validated JWT, so callers cannot
		// mix one user's bearer token with another user's session cookie.
		session := sessions.Default(c)
		profileSession, err := utils.GetProfileFromSession(session)
		if err != nil {
			a.Logger.Error("JWTMiddleware - profile not found in session")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "cannot find profile in session"})
			c.Abort()
			return
		}
		if profileSession.ID.Hex() != claimsObj.ProfileID || profileSession.GithubID != claimsObj.ID {
			a.Logger.Errorw("JWTMiddleware - session/JWT identity mismatch",
				"sessionProfileID", profileSession.ID.Hex(),
				"jwtProfileID", claimsObj.ProfileID,
				"sessionGithubID", profileSession.GithubID,
				"jwtGithubID", claimsObj.ID,
			)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "session does not match token identity"})
			c.Abort()
			return
		}

		c.Set("jwt_claims", claimsObj)
		c.Next()
	}
}
