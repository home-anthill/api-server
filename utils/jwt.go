package utils

import (
	"api-server/models"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenType distinguishes between access and refresh JWTs.
type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

// JWTIssuer and JWTAudience are the expected iss/aud claim values for all tokens.
const JWTIssuer = "home-anthill-api"
const JWTAudience = "home-anthill-api"

// JWTClaims creates a struct that will be encoded to a JWT.
// We add jwt.RegisteredClaims as an embedded type, to provide fields like expiry time
type JWTClaims struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	TokenType TokenType `json:"tokenType"`
	jwt.RegisteredClaims
}

func CreateJWT(profile models.Profile, expirationTime time.Time, tokenType TokenType, signingMethod jwt.SigningMethod, jwtKey []byte) (string, error) {
	claimsObj := &JWTClaims{
		ID:        profile.Github.ID,
		Name:      profile.Github.Name,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			Issuer:    JWTIssuer,
			Audience:  jwt.ClaimStrings{JWTAudience},
			Subject:   strconv.FormatInt(profile.Github.ID, 10),
		},
	}
	token := jwt.NewWithClaims(signingMethod, claimsObj)
	tokenString, err := token.SignedString(jwtKey)
	return tokenString, err
}
