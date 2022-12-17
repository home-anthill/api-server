package api

import (
  "api-server/models"
  "context"
  "fmt"
  "github.com/gin-gonic/gin"
  "github.com/golang-jwt/jwt/v4"
  "go.mongodb.org/mongo-driver/mongo"
  "go.uber.org/zap"
  "net/http"
  "net/url"
  "os"
  "time"
)

type Auth struct {
  collection *mongo.Collection
  ctx        context.Context
  logger     *zap.SugaredLogger
}

// claims Creates a struct that will be encoded to a JWT.
// We add jwt.RegisteredClaims as an embedded type, to provide fields like expiry time
type claims struct {
  ID   int64  `json:"id"`
  Name string `json:"name"`
  jwt.RegisteredClaims
}

// For HMAC signing method, the key can be any []byte. It is recommended to generate
// a key using crypto/rand or something equivalent. You need the same key for signing
// and validating.
var jwtKey = []byte(os.Getenv("JWT_PASSWORD"))

func NewAuth(ctx context.Context, logger *zap.SugaredLogger, collection *mongo.Collection) *Auth {
  return &Auth{
    collection: collection,
    ctx:        ctx,
    logger:     logger,
  }
}

func (handler *Auth) LoginCallback(c *gin.Context) {
  handler.logger.Info("REST - GET - LoginCallback called")

  var profile = c.Value("profile").(models.Profile)

  expirationTime := time.Now().Add(60 * time.Minute)

  claims := &claims{
    ID:   profile.Github.ID,
    Name: profile.Github.Name,
    RegisteredClaims: jwt.RegisteredClaims{
      ExpiresAt: jwt.NewNumericDate(expirationTime),
    },
  }
  token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
  tokenString, err := token.SignedString(jwtKey)

  if err != nil {
    handler.logger.Error("REST - GET - LoginCallback - cannot generate JWT")
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

func (handler *Auth) JWTMiddleware() gin.HandlerFunc {
  return func(c *gin.Context) {
    const BearerSchema = "Bearer"
    authHeader := c.GetHeader("Authorization")

    if authHeader == "" {
      handler.logger.Error("JWTMiddleware - authorization header not found")
      c.JSON(http.StatusUnauthorized, gin.H{
        "message": "authorization header not found",
      })
      c.Abort()
      return
    }

    tokenString := authHeader[(len(BearerSchema) + 1):]

    if tokenString == "" {
      handler.logger.Error("JWTMiddleware - bearer token not found")
      c.JSON(http.StatusUnauthorized, gin.H{
        "message": "bearer token not found",
      })
      c.Abort()
      return
    }

    claims := &claims{}

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

    if token == nil || !token.Valid {
      if ve, ok := err.(*jwt.ValidationError); ok {
        if ve.Errors&jwt.ValidationErrorMalformed != 0 {
          handler.logger.Error("JWTMiddleware - that's not even a token")
          c.JSON(http.StatusBadRequest, gin.H{
            "message": "that's not even a token",
          })
          c.Abort()
          return
        } else if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
          // Token is either expired or not active yet
          handler.logger.Error("JWTMiddleware - JWT token expired")
          c.JSON(http.StatusUnauthorized, gin.H{
            "message": "JWT token is expired",
          })
          c.Abort()
          return
        }

        handler.logger.Error("JWTMiddleware - not logged, token is not valid")
        c.JSON(http.StatusForbidden, gin.H{"message": "not logged, token is not valid"})
        c.Abort()
        return
      }
    }

    if err != nil {
      if err == jwt.ErrSignatureInvalid {
        handler.logger.Error("JWTMiddleware - cannot login", err)
        c.JSON(http.StatusUnauthorized, gin.H{"message": "cannot login"})
        c.Abort()
        return
      }
      handler.logger.Error("JWTMiddleware - bad request while login", err)
      c.JSON(http.StatusBadRequest, gin.H{"message": "bad request while login"})
      c.Abort()
      return
    }

    //fmt.Println("Valid token: ", claims.ID, claims.Name, claims.ExpiresAt)
    c.Next()
  }
}
