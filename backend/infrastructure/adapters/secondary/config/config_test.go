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

	t.Setenv("MAX_UPLOAD_BYTES", "2GB")
	if got := resolveMaxUploadBytes(); got != 2<<30 {
		t.Fatalf("resolveMaxUploadBytes(\"2GB\") = %d", got)
	}

	t.Setenv("MAX_UPLOAD_BYTES", "invalid")
	if got := resolveMaxUploadBytes(); got != 0 {
		t.Fatalf("expected unlimited default for invalid value, got %d", got)
	}

	t.Setenv("MAX_UPLOAD_BYTES", "unlimited")
	if got := resolveMaxUploadBytes(); got != 0 {
		t.Fatalf("resolveMaxUploadBytes(\"unlimited\") = %d, want 0", got)
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
	// Uploads are unlimited by default (0 = no cap).
	if cfg.GetMaxUploadBytes() != 0 {
		t.Fatalf("expected unlimited uploads by default, got %d", cfg.GetMaxUploadBytes())
	}
}
