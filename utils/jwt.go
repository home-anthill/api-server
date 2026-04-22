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

// CreateJWT builds and signs a local access JWT for an authenticated profile.
// Callers must pass an explicit expiration, expected token type, approved HMAC
// signing method, and the correct signing key for that token class. The token
// includes issuer, audience, subject, issued-at, not-before, and expiry claims
// so validation can reject tokens from the wrong context or outside their
// validity window.
func CreateJWT(profile models.Profile, expirationTime time.Time, tokenType TokenType, signingMethod jwt.SigningMethod, jwtKey []byte) (string, error) {
	claims := &JWTClaims{
		ID:        profile.Github.ID,
		Name:      profile.Github.Name,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    JWTIssuer,
			Audience:  jwt.ClaimStrings{JWTAudience},
			Subject:   strconv.FormatInt(profile.Github.ID, 10),
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			NotBefore: jwt.NewNumericDate(time.Now().UTC()),
		},
	}
	token := jwt.NewWithClaims(signingMethod, claims)
	signed, err := token.SignedString(jwtKey)
	return signed, err
}
