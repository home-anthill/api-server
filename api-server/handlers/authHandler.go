package handlers

import (
	"air-conditioner/models"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"net/url"
	"time"
)

type AuthHandler struct {
	collection *mongo.Collection
	ctx        context.Context
}

// Create a struct that will be encoded to a JWT.
// We add jwt.StandardClaims as an embedded type, to provide fields like expiry time
type Claims struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	jwt.StandardClaims
}

// For HMAC signing method, the key can be any []byte. It is recommended to generate
// a key using crypto/rand or something equivalent. You need the same key for signing
// and validating.
var jwtKey = []byte("secretkey")

func NewAuthHandler(ctx context.Context, collection *mongo.Collection) *AuthHandler {
	return &AuthHandler{
		collection: collection,
		ctx:        ctx,
	}
}

func (handler *AuthHandler) LoginCallbackHandler(c *gin.Context) {
	var profile = c.Value("profile").(models.Profile)
	fmt.Println("LoginCallbackHandler with profile = ", profile)

	expirationTime := time.Now().Add(20 * time.Minute)

	claims := &Claims{
		ID: profile.Github.ID,
		Name: profile.Github.Name,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "Cannot generate JWT"})
		return
	}

	// second solution
	q := url.Values{}
	q.Set("token", tokenString)
	location := url.URL{Path: "/postlogin", RawQuery: q.Encode()}
	c.Redirect(http.StatusFound, location.RequestURI())

	c.JSON(http.StatusOK, gin.H{"token": tokenString})

	//// Finally, we set the client cookie for "token" as the JWT we just generated
	//// we also set an expiry time which is the same as the token itself
	//http.SetCookie(w, &http.Cookie{
	//	Name:    "token",
	//	Value:   tokenString,
	//	Expires: expirationTime,
	//})
}

func (handler *AuthHandler) JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		const BEARER_SCHEMA = "Bearer"
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "Authorization header not found",
			})
			c.Abort()
			return
		}

		tokenString := authHeader[(len(BEARER_SCHEMA) + 1):]
		fmt.Println("tokenString ", tokenString, len(tokenString))

		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "Bearer token not found",
			})
			c.Abort()
			return
		}

		claims := &Claims{}

		// Parse takes the token string and a function for looking up the key. The latter is especially
		// useful if you use multiple keys for your application.  The standard is to use 'kid' in the
		// head of the token to identify which key to use, but the parsed token (head and claims) is provided
		// to the callback, providing flexibility.
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			// Don't forget to validate the alg is what you expect:
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}
			return jwtKey, nil
		})

		if !token.Valid {
			if ve, ok := err.(*jwt.ValidationError); ok {
				if ve.Errors&jwt.ValidationErrorMalformed != 0 {
					fmt.Println("That's not even a token")
					c.JSON(http.StatusBadRequest, gin.H{
						"message": "That's not even a token",
					})
					c.Abort()
					return
				} else if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
					// Token is either expired or not active yet
					fmt.Println("Timing is everything")
					c.JSON(http.StatusUnauthorized, gin.H{
						"message": "Timing is everything",
					})
					c.Abort()
					return
				}

				c.JSON(http.StatusForbidden, gin.H{
					"message": "Not logged, token is not valud",
				})
				c.Abort()
				return
			}
		}

		if err != nil {
			fmt.Println("err != nil", err)
			if err == jwt.ErrSignatureInvalid {
				c.JSON(http.StatusUnauthorized, gin.H{
					"message": "Cannot login",
				})
				c.Abort()
				return
			}
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Bad request while login",
			})
			c.Abort()
			return
		}

		fmt.Println("Valid token: ", claims.ID, claims.Name, claims.ExpiresAt)

		c.Next()
	}
}
