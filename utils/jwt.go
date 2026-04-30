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
	ID         int64     `json:"id"`
	ProfileID  string    `json:"profileId"`
	Name       string    `json:"name"`
	TokenType  TokenType `json:"tokenType"`
	ClientType string    `json:"clientType"`
	jwt.RegisteredClaims
}

// CreateJWT builds and signs a local JWT for an authenticated profile.
// The algorithm is intentionally fixed to HS512 so callers cannot accidentally
// issue tokens with a weaker or inconsistent signing method. The token includes
// issuer, audience, subject, issued-at, not-before, and expiry claims so
// validation can reject tokens from the wrong context or outside their validity
// window.
func CreateJWT(profile models.Profile, expirationTime time.Time, tokenType TokenType, clientType string, jwtKey []byte) (string, error) {
	now := time.Now().UTC()
	claims := &JWTClaims{
		ID:         profile.Github.ID,
		ProfileID:  profile.ID.Hex(),
		Name:       profile.Github.Name,
		TokenType:  tokenType,
		ClientType: clientType,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    JWTIssuer,
			Audience:  jwt.ClaimStrings{JWTAudience},
			Subject:   strconv.FormatInt(profile.Github.ID, 10),
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	signed, err := token.SignedString(jwtKey)
	return signed, err
}
