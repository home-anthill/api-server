package initialization

import (
	"testing"

	"go.uber.org/zap"
)

func setValidEnv(t *testing.T) {
	t.Helper()

	t.Setenv("ENV", "prod")
	t.Setenv("JWT_PASSWORD", "0123456789abcdef0123456789abcdef")
	t.Setenv("JWT_REFRESH_PASSWORD", "fedcba9876543210fedcba9876543210")
	t.Setenv("REFRESH_TOKEN_HASH_SECRET", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	t.Setenv("COOKIE_SECRET", "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
	t.Setenv("OAUTH2_CLIENTID", "web-client-id")
	t.Setenv("OAUTH2_SECRETID", "web-client-secret")
	t.Setenv("OAUTH2_APP_CLIENTID", "app-client-id")
	t.Setenv("OAUTH2_APP_SECRETID", "app-client-secret")
	t.Setenv("OAUTH2_CALLBACK", "https://example.com/api/oauth/callback")
	t.Setenv("OAUTH2_APP_CALLBACK", "https://example.com/api/oauth/app/callback")
	t.Setenv("HTTP_CORS", "false")
}

func TestPrintEnvAcceptsValidConfig(t *testing.T) {
	setValidEnv(t)

	if err := printEnv(zap.NewNop().Sugar()); err != nil {
		t.Fatalf("expected valid env, got error: %v", err)
	}
}

func TestPrintEnvRejectsMissingOauthSecret(t *testing.T) {
	setValidEnv(t)
	t.Setenv("OAUTH2_SECRETID", "")

	if err := printEnv(zap.NewNop().Sugar()); err == nil {
		t.Fatal("expected error for missing OAUTH2_SECRETID")
	}
}

func TestPrintEnvRejectsShortCookieSecret(t *testing.T) {
	setValidEnv(t)
	t.Setenv("COOKIE_SECRET", "short")

	if err := printEnv(zap.NewNop().Sugar()); err == nil {
		t.Fatal("expected error for short COOKIE_SECRET")
	}
}

func TestPrintEnvRejectsProdCors(t *testing.T) {
	setValidEnv(t)
	t.Setenv("HTTP_CORS", "true")

	if err := printEnv(zap.NewNop().Sugar()); err == nil {
		t.Fatal("expected error for prod HTTP_CORS=true")
	}
}
