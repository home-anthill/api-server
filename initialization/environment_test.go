package initialization

import (
	"os"
	"path/filepath"
	"strings"
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
	t.Setenv("API_TOKEN_HASH_SECRET", "cccccccccccccccccccccccccccccccc")
	t.Setenv("API_TOKEN_ENCRYPTION_KEY", "dddddddddddddddddddddddddddddddd")
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

func TestPrintEnvRejectsMissingAPITokenHashSecret(t *testing.T) {
	setValidEnv(t)
	t.Setenv("API_TOKEN_HASH_SECRET", "")

	if err := printEnv(zap.NewNop().Sugar()); err == nil {
		t.Fatal("expected error for missing API_TOKEN_HASH_SECRET")
	}
}

func TestPrintEnvRejectsShortAPITokenHashSecret(t *testing.T) {
	setValidEnv(t)
	t.Setenv("API_TOKEN_HASH_SECRET", "short")

	if err := printEnv(zap.NewNop().Sugar()); err == nil {
		t.Fatal("expected error for short API_TOKEN_HASH_SECRET")
	}
}

func TestPrintEnvRejectsProdCors(t *testing.T) {
	setValidEnv(t)
	t.Setenv("HTTP_CORS", "true")

	if err := printEnv(zap.NewNop().Sugar()); err == nil {
		t.Fatal("expected error for prod HTTP_CORS=true")
	}
}

func TestPrintEnvRejectsShortSecretsAndInvalidURLs(t *testing.T) {
	tests := []struct {
		name   string
		envKey string
		value  string
	}{
		{name: "short JWT_PASSWORD", envKey: "JWT_PASSWORD", value: "short"},
		{name: "short JWT_REFRESH_PASSWORD", envKey: "JWT_REFRESH_PASSWORD", value: "short"},
		{name: "short REFRESH_TOKEN_HASH_SECRET", envKey: "REFRESH_TOKEN_HASH_SECRET", value: "short"},
		{name: "callback missing host", envKey: "OAUTH2_CALLBACK", value: "callback"},
		{name: "app callback missing host", envKey: "OAUTH2_APP_CALLBACK", value: "callback"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setValidEnv(t)
			t.Setenv(tt.envKey, tt.value)

			if err := printEnv(zap.NewNop().Sugar()); err == nil {
				t.Fatalf("expected error for %s=%q", tt.envKey, tt.value)
			}
		})
	}
}

func TestPrintEnvRejectsMissingRequiredVariables(t *testing.T) {
	required := []string{
		"JWT_PASSWORD",
		"JWT_REFRESH_PASSWORD",
		"REFRESH_TOKEN_HASH_SECRET",
		"COOKIE_SECRET",
		"API_TOKEN_HASH_SECRET",
		"API_TOKEN_ENCRYPTION_KEY",
		"OAUTH2_CLIENTID",
		"OAUTH2_SECRETID",
		"OAUTH2_APP_CLIENTID",
		"OAUTH2_APP_SECRETID",
		"OAUTH2_CALLBACK",
		"OAUTH2_APP_CALLBACK",
	}

	for _, name := range required {
		t.Run(name, func(t *testing.T) {
			setValidEnv(t)
			t.Setenv(name, " ")

			err := printEnv(zap.NewNop().Sugar())
			if err == nil {
				t.Fatalf("expected missing %s error", name)
			}
			if !strings.Contains(err.Error(), name) {
				t.Fatalf("error %q does not mention %s", err.Error(), name)
			}
		})
	}
}

func TestReadEnv(t *testing.T) {
	t.Run("loads env file from project root in working directory path", func(t *testing.T) {
		root := filepath.Join(t.TempDir(), "api-server")
		nested := filepath.Join(root, "nested")
		if err := os.MkdirAll(nested, 0o755); err != nil {
			t.Fatalf("MkdirAll returned error: %v", err)
		}
		const envName = "READ_ENV_TEST_VALUE"
		previous, existed := os.LookupEnv(envName)
		t.Cleanup(func() {
			if existed {
				_ = os.Setenv(envName, previous)
				return
			}
			_ = os.Unsetenv(envName)
		})
		_ = os.Unsetenv(envName)
		if err := os.WriteFile(filepath.Join(root, ".env"), []byte(envName+"=loaded\n"), 0o600); err != nil {
			t.Fatalf("WriteFile returned error: %v", err)
		}
		t.Chdir(nested)

		envFile, err := readEnv()
		if err != nil {
			t.Fatalf("readEnv returned error: %v", err)
		}
		if envFile != filepath.Join(root, ".env") {
			t.Fatalf("readEnv envFile = %q, want %q", envFile, filepath.Join(root, ".env"))
		}
		if got := os.Getenv(envName); got != "loaded" {
			t.Fatalf("loaded env = %q, want loaded", got)
		}
	})

	t.Run("returns error when env file is missing", func(t *testing.T) {
		t.Chdir(t.TempDir())

		envFile, err := readEnv()
		if err == nil {
			t.Fatal("expected missing .env error")
		}
		if envFile == "" {
			t.Fatal("expected env file path to be returned")
		}
	})
}

func TestInitEnv(t *testing.T) {
	setValidEnv(t)

	root := filepath.Join(t.TempDir(), "api-server")
	nested := filepath.Join(root, "cmd")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}

	envContent := strings.Join([]string{
		"ENV=testing",
		"JWT_PASSWORD=0123456789abcdef0123456789abcdef",
		"JWT_REFRESH_PASSWORD=fedcba9876543210fedcba9876543210",
		"REFRESH_TOKEN_HASH_SECRET=aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"COOKIE_SECRET=bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		"API_TOKEN_HASH_SECRET=cccccccccccccccccccccccccccccccc",
		"API_TOKEN_ENCRYPTION_KEY=dddddddddddddddddddddddddddddddd",
		"OAUTH2_CLIENTID=web-client-id",
		"OAUTH2_SECRETID=web-client-secret",
		"OAUTH2_APP_CLIENTID=app-client-id",
		"OAUTH2_APP_SECRETID=app-client-secret",
		"OAUTH2_CALLBACK=https://example.com/api/oauth/callback",
		"OAUTH2_APP_CALLBACK=https://example.com/api/oauth/app/callback",
		"HTTP_CORS=false",
	}, "\n")
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte(envContent), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
	t.Chdir(nested)

	if err := InitEnv(zap.NewNop().Sugar()); err != nil {
		t.Fatalf("InitEnv returned error: %v", err)
	}
}

func TestInitEnvReturnsReadEnvError(t *testing.T) {
	t.Chdir(t.TempDir())

	if err := InitEnv(zap.NewNop().Sugar()); err == nil {
		t.Fatal("expected InitEnv to return readEnv error")
	}
}
