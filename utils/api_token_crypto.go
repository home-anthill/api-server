package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"
)

const apiTokenNonceSize = 12

// HashAPIToken returns a stable peppered hash for apiToken lookups.
func HashAPIToken(apiToken string) (string, error) {
	secret, err := getAPITokenHashSecret()
	if err != nil {
		return "", err
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(apiToken))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}

// EncryptAPIToken encrypts the raw apiToken for services that still need it as
// an HMAC key. The encoded value is base64url(nonce || ciphertext).
func EncryptAPIToken(apiToken string) (string, error) {
	key, err := getAPITokenEncryptionKey()
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, apiTokenNonceSize)
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nil, nonce, []byte(apiToken), nil)
	encoded := append(nonce, ciphertext...)
	return base64.RawURLEncoding.EncodeToString(encoded), nil
}

// DecryptAPIToken decrypts a value created by EncryptAPIToken.
func DecryptAPIToken(encrypted string) (string, error) {
	raw, err := base64.RawURLEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}
	if len(raw) <= apiTokenNonceSize {
		return "", fmt.Errorf("encrypted api token is too short")
	}
	key, err := getAPITokenEncryptionKey()
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	plaintext, err := gcm.Open(nil, raw[:apiTokenNonceSize], raw[apiTokenNonceSize:], nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

func getAPITokenHashSecret() (string, error) {
	if secret := os.Getenv("API_TOKEN_HASH_SECRET"); secret != "" {
		return secret, nil
	}
	return "", fmt.Errorf("API_TOKEN_HASH_SECRET is required")
}

func getAPITokenEncryptionKey() ([]byte, error) {
	key := os.Getenv("API_TOKEN_ENCRYPTION_KEY")
	if key == "" {
		return nil, fmt.Errorf("API_TOKEN_ENCRYPTION_KEY is required")
	}
	if decoded, err := base64.RawURLEncoding.DecodeString(key); err == nil && len(decoded) == 32 {
		return decoded, nil
	}
	if decoded, err := base64.StdEncoding.DecodeString(key); err == nil && len(decoded) == 32 {
		return decoded, nil
	}
	if len(key) == 32 {
		return []byte(key), nil
	}
	return nil, fmt.Errorf("API_TOKEN_ENCRYPTION_KEY must be 32 raw bytes or base64-encoded 32 bytes")
}
