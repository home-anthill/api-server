package utils

import (
	"api-server/models"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

// JWTClaims creates a struct that will be encoded to a JWT.
// We add jwt.RegisteredClaims as an embedded type, to provide fields like expiry time
type JWTClaims struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	jwt.RegisteredClaims
}

func CreateJWT(profile models.Profile, expirationTime time.Time, singingMethod jwt.SigningMethod, jwtKey []byte) (string, error) {
	claimsObj := &JWTClaims{
		ID:   profile.Github.ID,
		Name: profile.Github.Name,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(singingMethod, claimsObj)
	tokenString, err := token.SignedString(jwtKey)
	return tokenString, err
}
