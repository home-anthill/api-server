package utils

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestPKCEValidation(t *testing.T) {
	validVerifier := strings.Repeat("A", 43)
	if !IsValidPKCEVerifier(validVerifier) {
		t.Fatal("expected 43 character verifier to be valid")
	}
	if !IsValidPKCEVerifier(strings.Repeat("~", 128)) {
		t.Fatal("expected 128 character verifier with RFC character to be valid")
	}
	for _, verifier := range []string{strings.Repeat("A", 42), strings.Repeat("A", 129), "bad verifier!"} {
		if IsValidPKCEVerifier(verifier) {
			t.Fatalf("expected verifier %q to be invalid", verifier)
		}
	}

	validChallenge := strings.Repeat("_", 43)
	if !IsValidPKCECodeChallenge(validChallenge) {
		t.Fatal("expected 43 character challenge to be valid")
	}
	for _, challenge := range []string{strings.Repeat("_", 42), strings.Repeat("_", 129), "bad.challenge"} {
		if IsValidPKCECodeChallenge(challenge) {
			t.Fatalf("expected challenge %q to be invalid", challenge)
		}
	}

	validAppCode := strings.Repeat("-", 128)
	if !IsValidAppLoginCode(validAppCode) {
		t.Fatal("expected 128 character app login code to be valid")
	}
	for _, code := range []string{strings.Repeat("-", 127), strings.Repeat("-", 129), strings.Repeat(".", 128)} {
		if IsValidAppLoginCode(code) {
			t.Fatalf("expected app login code %q to be invalid", code)
		}
	}
}

func TestNewPKCEVerifier(t *testing.T) {
	verifier, err := NewPKCEVerifier()
	if err != nil {
		t.Fatalf("NewPKCEVerifier returned error: %v", err)
	}
	if len(verifier) != 128 {
		t.Fatalf("NewPKCEVerifier length = %d, want 128", len(verifier))
	}
	if !IsValidPKCEVerifier(verifier) {
		t.Fatalf("NewPKCEVerifier returned invalid verifier: %q", verifier)
	}
}

func TestBuildPKCECodeChallenge(t *testing.T) {
	const verifier = "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	const want = "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"

	got, err := BuildPKCECodeChallenge(verifier)
	if err != nil {
		t.Fatalf("BuildPKCECodeChallenge returned error: %v", err)
	}
	if got != want {
		t.Fatalf("BuildPKCECodeChallenge() = %q, want %q", got, want)
	}

	if got, err = BuildPKCECodeChallenge("short"); err == nil || got != "" {
		t.Fatalf("BuildPKCECodeChallenge invalid verifier = %q, %v; want empty error", got, err)
	}
}

func TestHashToken(t *testing.T) {
	t.Setenv("JWT_PASSWORD", "jwt-secret")
	t.Setenv("JWT_REFRESH_PASSWORD", "refresh-secret")
	t.Setenv("REFRESH_TOKEN_HASH_SECRET", "preferred-secret")

	got := HashToken("token")
	again := HashToken("token")
	if got == "" {
		t.Fatal("HashToken returned empty hash")
	}
	if got != again {
		t.Fatalf("HashToken is not stable: %q != %q", got, again)
	}

	t.Setenv("REFRESH_TOKEN_HASH_SECRET", "other-preferred-secret")
	if other := HashToken("token"); other == got {
		t.Fatal("HashToken did not change after changing HMAC secret")
	}
}

func TestRefreshTokenHashSecretPrecedence(t *testing.T) {
	t.Setenv("JWT_PASSWORD", "jwt")
	t.Setenv("JWT_REFRESH_PASSWORD", "refresh")
	t.Setenv("REFRESH_TOKEN_HASH_SECRET", "preferred")
	if got := refreshTokenHashSecret(); got != "preferred" {
		t.Fatalf("refreshTokenHashSecret() = %q, want preferred", got)
	}

	t.Setenv("REFRESH_TOKEN_HASH_SECRET", "")
	if got := refreshTokenHashSecret(); got != "refresh" {
		t.Fatalf("refreshTokenHashSecret() = %q, want refresh", got)
	}

	t.Setenv("JWT_REFRESH_PASSWORD", "")
	if got := refreshTokenHashSecret(); got != "jwt" {
		t.Fatalf("refreshTokenHashSecret() = %q, want jwt", got)
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{name: "shorter than limit", input: "hello", maxLen: 10, want: "hello"},
		{name: "equal to limit", input: "hello", maxLen: 5, want: "hello"},
		{name: "longer than limit", input: "hello", maxLen: 3, want: "hel"},
		{name: "negative disables truncation", input: "hello", maxLen: -1, want: "hello"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TruncateString(tt.input, tt.maxLen); got != tt.want {
				t.Fatalf("TruncateString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestHashAPIToken(t *testing.T) {
	t.Run("returns stable hash with configured secret", func(t *testing.T) {
		t.Setenv("API_TOKEN_HASH_SECRET", "hash-secret")
		got, err := HashAPIToken("api-token")
		if err != nil {
			t.Fatalf("HashAPIToken returned error: %v", err)
		}
		again, err := HashAPIToken("api-token")
		if err != nil {
			t.Fatalf("HashAPIToken returned error: %v", err)
		}
		if got == "" || got != again {
			t.Fatalf("HashAPIToken unstable or empty: %q %q", got, again)
		}
	})

	t.Run("requires secret", func(t *testing.T) {
		t.Setenv("API_TOKEN_HASH_SECRET", "")
		if got, err := HashAPIToken("api-token"); err == nil || got != "" {
			t.Fatalf("HashAPIToken without secret = %q, %v; want empty error", got, err)
		}
	})
}

func TestEncryptDecryptAPIToken(t *testing.T) {
	t.Run("round trips with raw key", func(t *testing.T) {
		t.Setenv("API_TOKEN_ENCRYPTION_KEY", "12345678901234567890123456789012")
		encrypted, err := EncryptAPIToken("api-token")
		if err != nil {
			t.Fatalf("EncryptAPIToken returned error: %v", err)
		}
		if encrypted == "" || encrypted == "api-token" {
			t.Fatalf("EncryptAPIToken returned suspicious value: %q", encrypted)
		}
		decrypted, err := DecryptAPIToken(encrypted)
		if err != nil {
			t.Fatalf("DecryptAPIToken returned error: %v", err)
		}
		if decrypted != "api-token" {
			t.Fatalf("DecryptAPIToken() = %q, want api-token", decrypted)
		}
	})

	t.Run("accepts raw url base64 key", func(t *testing.T) {
		key := base64.RawURLEncoding.EncodeToString([]byte("12345678901234567890123456789012"))
		t.Setenv("API_TOKEN_ENCRYPTION_KEY", key)
		encrypted, err := EncryptAPIToken("api-token")
		if err != nil {
			t.Fatalf("EncryptAPIToken returned error: %v", err)
		}
		if decrypted, err := DecryptAPIToken(encrypted); err != nil || decrypted != "api-token" {
			t.Fatalf("DecryptAPIToken() = %q, %v; want api-token", decrypted, err)
		}
	})

	t.Run("accepts standard base64 key", func(t *testing.T) {
		key := base64.StdEncoding.EncodeToString([]byte("12345678901234567890123456789012"))
		t.Setenv("API_TOKEN_ENCRYPTION_KEY", key)
		encrypted, err := EncryptAPIToken("api-token")
		if err != nil {
			t.Fatalf("EncryptAPIToken returned error: %v", err)
		}
		if decrypted, err := DecryptAPIToken(encrypted); err != nil || decrypted != "api-token" {
			t.Fatalf("DecryptAPIToken() = %q, %v; want api-token", decrypted, err)
		}
	})
}

func TestAPITokenCryptoErrors(t *testing.T) {
	t.Run("encrypt requires key", func(t *testing.T) {
		t.Setenv("API_TOKEN_ENCRYPTION_KEY", "")
		if got, err := EncryptAPIToken("api-token"); err == nil || got != "" {
			t.Fatalf("EncryptAPIToken without key = %q, %v; want empty error", got, err)
		}
	})

	t.Run("rejects invalid key length", func(t *testing.T) {
		t.Setenv("API_TOKEN_ENCRYPTION_KEY", "short")
		if key, err := getAPITokenEncryptionKey(); err == nil || key != nil {
			t.Fatalf("getAPITokenEncryptionKey invalid key = %v, %v; want nil error", key, err)
		}
	})

	t.Run("decrypt rejects invalid base64", func(t *testing.T) {
		t.Setenv("API_TOKEN_ENCRYPTION_KEY", "12345678901234567890123456789012")
		if got, err := DecryptAPIToken("not base64!"); err == nil || got != "" {
			t.Fatalf("DecryptAPIToken invalid base64 = %q, %v; want empty error", got, err)
		}
	})

	t.Run("decrypt rejects short ciphertext", func(t *testing.T) {
		t.Setenv("API_TOKEN_ENCRYPTION_KEY", "12345678901234567890123456789012")
		short := base64.RawURLEncoding.EncodeToString([]byte("short"))
		if got, err := DecryptAPIToken(short); err == nil || got != "" {
			t.Fatalf("DecryptAPIToken short ciphertext = %q, %v; want empty error", got, err)
		}
	})

	t.Run("decrypt rejects tampered ciphertext", func(t *testing.T) {
		t.Setenv("API_TOKEN_ENCRYPTION_KEY", "12345678901234567890123456789012")
		encrypted, err := EncryptAPIToken("api-token")
		if err != nil {
			t.Fatalf("EncryptAPIToken returned error: %v", err)
		}
		raw, err := base64.RawURLEncoding.DecodeString(encrypted)
		if err != nil {
			t.Fatalf("DecodeString returned error: %v", err)
		}
		raw[len(raw)-1] ^= 1
		tampered := base64.RawURLEncoding.EncodeToString(raw)
		if got, err := DecryptAPIToken(tampered); err == nil || got != "" {
			t.Fatalf("DecryptAPIToken tampered ciphertext = %q, %v; want empty error", got, err)
		}
	})
}

func TestAPITokenConfigHelpers(t *testing.T) {
	t.Run("hash secret helper returns configured value", func(t *testing.T) {
		t.Setenv("API_TOKEN_HASH_SECRET", "hash-secret")
		got, err := getAPITokenHashSecret()
		if err != nil {
			t.Fatalf("getAPITokenHashSecret returned error: %v", err)
		}
		if got != "hash-secret" {
			t.Fatalf("getAPITokenHashSecret() = %q, want hash-secret", got)
		}
	})

	t.Run("encryption key helper returns raw key", func(t *testing.T) {
		t.Setenv("API_TOKEN_ENCRYPTION_KEY", "12345678901234567890123456789012")
		got, err := getAPITokenEncryptionKey()
		if err != nil {
			t.Fatalf("getAPITokenEncryptionKey returned error: %v", err)
		}
		if string(got) != "12345678901234567890123456789012" {
			t.Fatalf("getAPITokenEncryptionKey() = %q", string(got))
		}
	})
}
