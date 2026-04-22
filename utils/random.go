package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// RandomString returns n cryptographically secure random bytes encoded as
// unpadded base64url text. Use it for OAuth state, PKCE verifiers, one-time
// codes, and opaque tokens that must be unguessable.
func RandomString(n int) (string, error) {
	if n <= 0 {
		return "", fmt.Errorf("token length must be positive")
	}

	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate random bytes: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
