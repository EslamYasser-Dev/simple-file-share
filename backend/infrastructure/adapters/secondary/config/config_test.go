package config

import (
	"testing"
)

func TestResolveRootDirProductionDefault(t *testing.T) {
	t.Setenv("ROOT_DIR", "")
	t.Setenv("FILE_SHARE_ROOT", "")
	t.Setenv("APP_ENV", "production")

	root, err := resolveRootDir()
	if err != nil {
		t.Fatal(err)
	}
	if root != "/data" {
		t.Fatalf("resolveRootDir() = %q, want /data", root)
	}
}

func TestResolveUsernameAliases(t *testing.T) {
	t.Setenv("USERNAME", "")
	t.Setenv("FILE_SHARE_USERNAME", "from-alias")

	if got := resolveUsername(); got != "from-alias" {
		t.Fatalf("resolveUsername() = %q", got)
	}
}

func TestResolveMaxUploadBytes(t *testing.T) {
	t.Setenv("MAX_UPLOAD_BYTES", "2048")
	if got := resolveMaxUploadBytes(); got != 2048 {
		t.Fatalf("resolveMaxUploadBytes() = %d", got)
	}

	t.Setenv("MAX_UPLOAD_BYTES", "invalid")
	if got := resolveMaxUploadBytes(); got != defaultMaxUploadBytes {
		t.Fatalf("expected default for invalid value, got %d", got)
	}
}

func TestEnvConfigProvider(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("ROOT_DIR", t.TempDir())
	t.Setenv("ENABLE_TLS", "false")
	t.Setenv("ENABLE_AUTH", "false")

	cfg, err := NewEnvConfigProvider()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.EnableTLS() {
		t.Fatal("expected TLS disabled")
	}
	if cfg.EnableAuth() {
		t.Fatal("expected auth disabled")
	}
	if cfg.GetMaxUploadBytes() <= 0 {
		t.Fatal("expected positive upload limit")
	}
}
