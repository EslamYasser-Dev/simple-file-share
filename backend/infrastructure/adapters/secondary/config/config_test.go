package config

import (
	"strings"
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

func TestResolveDefaultQuotaBytes(t *testing.T) {
	t.Setenv("QUOTA_DEFAULT_BYTES", "")
	if got := resolveDefaultQuotaBytes(); got != 0 {
		t.Fatalf("default quota should be unlimited, got %d", got)
	}

	t.Setenv("QUOTA_DEFAULT_BYTES", "100MB")
	if got := resolveDefaultQuotaBytes(); got != 100<<20 {
		t.Fatalf("resolveDefaultQuotaBytes(\"100MB\") = %d", got)
	}

	t.Setenv("QUOTA_DEFAULT_BYTES", "unlimited")
	if got := resolveDefaultQuotaBytes(); got != 0 {
		t.Fatalf("resolveDefaultQuotaBytes(\"unlimited\") = %d, want 0", got)
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

func TestResolveStorageBackend(t *testing.T) {
	cases := map[string]string{
		"":           "local",
		"local":      "local",
		"filesystem": "local",
		"s3":         "s3",
		"S3":         "s3",
		"unknown":    "unknown",
	}
	for in, want := range cases {
		t.Setenv("STORAGE_BACKEND", in)
		if got := resolveStorageBackend(); got != want {
			t.Fatalf("resolveStorageBackend(%q)=%q, want %q", in, got, want)
		}
	}
}

func TestResolveS3Settings(t *testing.T) {
	t.Setenv("S3_ENDPOINT", "http://127.0.0.1:9000/")
	t.Setenv("S3_BUCKET", "files")
	t.Setenv("S3_REGION", "eu-west-1")
	t.Setenv("S3_ACCESS_KEY", "ak")
	t.Setenv("S3_SECRET_KEY", "sk")
	t.Setenv("S3_PREFIX", "/tenant-a/")
	t.Setenv("S3_PATH_STYLE", "")

	s := resolveS3Settings()
	if s.Endpoint != "http://127.0.0.1:9000" {
		t.Fatalf("endpoint = %q", s.Endpoint)
	}
	if s.Bucket != "files" || s.Region != "eu-west-1" {
		t.Fatalf("bucket/region = %q/%q", s.Bucket, s.Region)
	}
	if s.AccessKey != "ak" || s.SecretKey != "sk" {
		t.Fatal("credentials not resolved")
	}
	if s.Prefix != "tenant-a" {
		t.Fatalf("prefix = %q", s.Prefix)
	}
	if !s.PathStyle {
		t.Fatal("custom endpoint should force path-style")
	}
}

func TestResolveRootDirRejectsTilde(t *testing.T) {
	t.Setenv("FILE_SHARE_ROOT", "")
	t.Setenv("ROOT_DIR", "~/.local/share/file-share")
	_, err := resolveRootDir()
	if err == nil {
		t.Fatal("expected an error for a tilde ROOT_DIR")
	}
	if !strings.Contains(err.Error(), "absolute path") {
		t.Fatalf("error should point at absolute paths, got: %v", err)
	}
}

func TestEnvConfigGRPCTLSFallsBackToTLS(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("ROOT_DIR", t.TempDir())
	t.Setenv("ENABLE_TLS", "false")
	t.Setenv("ENABLE_GRPC_TLS", "")

	cfg, err := NewEnvConfigProvider()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.EnableGRPCTLS() {
		t.Fatal("expected gRPC TLS to follow ENABLE_TLS=false")
	}
}

func TestEnvConfigGRPCTLSIndependentOverride(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("ROOT_DIR", t.TempDir())
	t.Setenv("ENABLE_TLS", "false")
	t.Setenv("ENABLE_GRPC_TLS", "true")

	cfg, err := NewEnvConfigProvider()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.EnableTLS() {
		t.Fatal("HTTP TLS must stay disabled for edge-terminated deployments")
	}
	if !cfg.EnableGRPCTLS() {
		t.Fatal("expected gRPC TLS enabled independently of HTTP TLS")
	}
}

func TestEnvConfigGRPCTLSDisableOverride(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("ROOT_DIR", t.TempDir())
	t.Setenv("ENABLE_TLS", "true")
	t.Setenv("ENABLE_GRPC_TLS", "0")

	cfg, err := NewEnvConfigProvider()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.EnableTLS() {
		t.Fatal("expected HTTP TLS enabled")
	}
	if cfg.EnableGRPCTLS() {
		t.Fatal("expected gRPC TLS disabled by override")
	}
}

func TestDevConfigGRPCTLS(t *testing.T) {
	t.Setenv("ENABLE_TLS", "false")
	t.Setenv("ENABLE_GRPC_TLS", "")
	cfg, err := NewDevConfigProvider()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.EnableGRPCTLS() {
		t.Fatal("expected gRPC TLS to follow ENABLE_TLS=false in dev")
	}
	t.Setenv("ENABLE_GRPC_TLS", "true")
	if !cfg.EnableGRPCTLS() {
		t.Fatal("expected dev provider to read ENABLE_GRPC_TLS live")
	}
}
