package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadReadsDotEnv(t *testing.T) {
	chdirTemp(t)
	unsetEnvKeys(t, "PORT", "MEDIA_MAX_UPLOAD_MB", "MEDIA_STORAGE_PATH", "MEDIA_PUBLIC_BASE_URL", "ADMIN_USERNAME", "ADMIN_EMAIL", "ADMIN_NAME", "ADMIN_PASSWORD", "SESSION_SECRET", "SESSION_TTL_HOURS", "COOKIE_SECURE", "USER_DEFAULT_QUOTA_GB", "SHARE_DEFAULT_TTL_DAYS")
	writeDotEnv(t, `
PORT=9090
MEDIA_MAX_UPLOAD_MB=4
MEDIA_STORAGE_PATH=uploads
MEDIA_PUBLIC_BASE_URL=http://localhost:9090/
ADMIN_USERNAME=admin-local
ADMIN_EMAIL=admin@example.test
ADMIN_NAME=Admin Local
ADMIN_PASSWORD=admin-local-password
SESSION_SECRET=local-session-secret-with-more-than-32-chars
SESSION_TTL_HOURS=24
COOKIE_SECURE=true
USER_DEFAULT_QUOTA_GB=20
SHARE_DEFAULT_TTL_DAYS=14
`)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load fallo: %v", err)
	}

	if cfg.Port != "9090" {
		t.Fatalf("port esperado 9090, recibido %s", cfg.Port)
	}
	if cfg.Media.MaxUploadMB != 4 {
		t.Fatalf("max upload esperado 4, recibido %d", cfg.Media.MaxUploadMB)
	}
	if cfg.Media.MaxUploadBytes != 4*1024*1024 {
		t.Fatalf("max upload bytes inesperado: %d", cfg.Media.MaxUploadBytes)
	}
	if cfg.Media.StoragePath != "uploads" {
		t.Fatalf("storage esperado uploads, recibido %s", cfg.Media.StoragePath)
	}
	if cfg.Media.PublicBaseURL != "http://localhost:9090" {
		t.Fatalf("public base url inesperada: %s", cfg.Media.PublicBaseURL)
	}
	if cfg.Media.SQLitePath != filepath.Join("uploads", defaultSQLiteStoreName) {
		t.Fatalf("sqlite path inesperado: %s", cfg.Media.SQLitePath)
	}
	if cfg.Auth.AdminUsername != "admin-local" {
		t.Fatalf("admin username inesperado: %s", cfg.Auth.AdminUsername)
	}
	if cfg.Auth.AdminEmail != "admin@example.test" {
		t.Fatalf("admin email inesperado: %s", cfg.Auth.AdminEmail)
	}
	if cfg.Auth.AdminName != "Admin Local" {
		t.Fatalf("admin name inesperado: %s", cfg.Auth.AdminName)
	}
	if cfg.Auth.AdminPassword != "admin-local-password" {
		t.Fatalf("admin password inesperado: %s", cfg.Auth.AdminPassword)
	}
	if cfg.Auth.SessionTTLHours != 24 {
		t.Fatalf("session ttl esperado 24, recibido %d", cfg.Auth.SessionTTLHours)
	}
	if !cfg.Auth.CookieSecure {
		t.Fatal("cookie secure esperado true")
	}
	if cfg.Product.DefaultQuotaGB != 20 {
		t.Fatalf("quota gb esperada 20, recibida %d", cfg.Product.DefaultQuotaGB)
	}
	if cfg.Product.DefaultQuotaBytes != 20*1024*1024*1024 {
		t.Fatalf("quota bytes inesperada: %d", cfg.Product.DefaultQuotaBytes)
	}
	if cfg.Product.ShareTTLDays != 14 {
		t.Fatalf("share ttl esperado 14, recibido %d", cfg.Product.ShareTTLDays)
	}
}

func TestLoadDoesNotOverrideExistingEnvironment(t *testing.T) {
	chdirTemp(t)
	unsetEnvKeys(t, "PORT", "MEDIA_MAX_UPLOAD_MB", "MEDIA_STORAGE_PATH", "MEDIA_PUBLIC_BASE_URL", "ADMIN_USERNAME", "ADMIN_EMAIL", "ADMIN_NAME", "ADMIN_PASSWORD", "SESSION_SECRET", "SESSION_TTL_HOURS", "COOKIE_SECURE", "USER_DEFAULT_QUOTA_GB", "SHARE_DEFAULT_TTL_DAYS")
	t.Setenv("ADMIN_PASSWORD", "from-env")
	writeDotEnv(t, `
ADMIN_PASSWORD=from-file
SESSION_SECRET=local-session-secret-with-more-than-32-chars
`)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load fallo: %v", err)
	}

	if cfg.Auth.AdminPassword != "from-env" {
		t.Fatalf("admin password esperado from-env, recibido %s", cfg.Auth.AdminPassword)
	}
}

func TestLoadReturnsDotEnvError(t *testing.T) {
	chdirTemp(t)
	unsetEnvKeys(t, "PORT", "MEDIA_MAX_UPLOAD_MB", "MEDIA_STORAGE_PATH", "MEDIA_PUBLIC_BASE_URL", "ADMIN_USERNAME", "ADMIN_EMAIL", "ADMIN_NAME", "ADMIN_PASSWORD", "SESSION_SECRET", "SESSION_TTL_HOURS", "COOKIE_SECURE", "USER_DEFAULT_QUOTA_GB", "SHARE_DEFAULT_TTL_DAYS")
	writeDotEnv(t, "ADMIN_PASSWORD\n")

	if _, err := Load(); err == nil {
		t.Fatal("se esperaba error por .env invalido")
	}
}

func chdirTemp(t *testing.T) {
	t.Helper()

	previousDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("leer cwd fallo: %v", err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatalf("chdir fallo: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(previousDir); err != nil {
			t.Fatalf("restaurar cwd fallo: %v", err)
		}
	})
}

func unsetEnvKeys(t *testing.T, keys ...string) {
	t.Helper()

	for _, key := range keys {
		previousValue, existed := os.LookupEnv(key)
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("unset %s fallo: %v", key, err)
		}
		t.Cleanup(func() {
			if existed {
				if err := os.Setenv(key, previousValue); err != nil {
					t.Fatalf("restaurar %s fallo: %v", key, err)
				}
				return
			}
			if err := os.Unsetenv(key); err != nil {
				t.Fatalf("limpiar %s fallo: %v", key, err)
			}
		})
	}
}

func writeDotEnv(t *testing.T, content string) {
	t.Helper()

	if err := os.WriteFile(".env", []byte(content), 0o600); err != nil {
		t.Fatalf("escribir .env fallo: %v", err)
	}
}
