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

func TestAuditConfigDefaultsAndOverrides(t *testing.T) {
	t.Setenv("ROOT_DIR", t.TempDir())
	cfg, err := NewEnvConfigProvider()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.EnableAuditLog() {
		t.Fatal("audit log must default to enabled")
	}
	if cfg.GetAuditMaxBytes() != 8<<20 {
		t.Fatalf("audit max bytes default = %d, want %d", cfg.GetAuditMaxBytes(), 8<<20)
	}
	if cfg.GetAuditKeep() != 3 {
		t.Fatalf("audit keep default = %d, want 3", cfg.GetAuditKeep())
	}

	t.Setenv("AUDIT_LOG", "false")
	t.Setenv("AUDIT_MAX_BYTES", "1MB")
	t.Setenv("AUDIT_KEEP", "5")
	cfg, err = NewEnvConfigProvider()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.EnableAuditLog() {
		t.Fatal("AUDIT_LOG=false must disable the audit trail")
	}
	if cfg.GetAuditMaxBytes() != 1<<20 {
		t.Fatalf("AUDIT_MAX_BYTES = %d, want %d", cfg.GetAuditMaxBytes(), 1<<20)
	}
	if cfg.GetAuditKeep() != 5 {
		t.Fatalf("AUDIT_KEEP = %d, want 5", cfg.GetAuditKeep())
	}

	// Out-of-range keeps clamp instead of failing boot.
	t.Setenv("AUDIT_KEEP", "not-a-number")
	cfg, err = NewEnvConfigProvider()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.GetAuditKeep() != 3 {
		t.Fatalf("bad AUDIT_KEEP must fall back to 3, got %d", cfg.GetAuditKeep())
	}
}

func TestTwoFactorConfigDefault(t *testing.T) {
	t.Setenv("ROOT_DIR", t.TempDir())
	cfg, err := NewEnvConfigProvider()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.EnableTwoFactor() {
		t.Fatal("ENABLE_2FA must default to enabled (enrollment is opt-in per account)")
	}
	t.Setenv("ENABLE_2FA", "false")
	cfg, err = NewEnvConfigProvider()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.EnableTwoFactor() {
		t.Fatal("ENABLE_2FA=false must disable enrollment routes")
	}
}

func TestAPIKeysConfigDefault(t *testing.T) {
	t.Setenv("ROOT_DIR", t.TempDir())
	cfg, err := NewEnvConfigProvider()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.EnableAPIKeys() {
		t.Fatal("ENABLE_API_KEYS must default to enabled")
	}
	dev, err := NewDevConfigProvider()
	if err != nil {
		t.Fatal(err)
	}
	if !dev.EnableAPIKeys() {
		t.Fatal("dev provider must also default API keys to enabled")
	}
	t.Setenv("ENABLE_API_KEYS", "false")
	cfg, err = NewEnvConfigProvider()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.EnableAPIKeys() {
		t.Fatal("ENABLE_API_KEYS=false must disable the endpoints and key auth")
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

func TestWebhookConfigDefaultsAndParsing(t *testing.T) {
	t.Setenv("WEBHOOK_URLS", "")
	t.Setenv("WEBHOOK_SECRET", "")
	if urls := resolveWebhookURLs(); urls != nil {
		t.Errorf("default urls = %v, want nil (webhooks off)", urls)
	}
	if s := resolveWebhookSecret(); s != "" {
		t.Errorf("default secret = %q, want empty", s)
	}

	t.Setenv("WEBHOOK_URLS", " https://a.example/hook , http://b.example/hook , ftp://bad/x , not-a-url , ")
	got := resolveWebhookURLs()
	want := []string{"https://a.example/hook", "http://b.example/hook"}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("urls = %v, want %v (http(s) with host only)", got, want)
	}

	t.Setenv("WEBHOOK_SECRET", "  s3cr3t  ")
	if s := resolveWebhookSecret(); s != "s3cr3t" {
		t.Errorf("secret = %q, want trimmed value", s)
	}
}

func TestProvidersExposeWebhookConfig(t *testing.T) {
	t.Setenv("WEBHOOK_URLS", "https://hooks.example/x")
	t.Setenv("WEBHOOK_SECRET", "k1")

	env, err := NewEnvConfigProvider()
	if err != nil {
		t.Fatal(err)
	}
	if len(env.GetWebhookURLs()) != 1 || env.GetWebhookURLs()[0] != "https://hooks.example/x" {
		t.Errorf("env urls = %v", env.GetWebhookURLs())
	}
	if env.GetWebhookSecret() != "k1" {
		t.Errorf("env secret = %q", env.GetWebhookSecret())
	}

	dev, err := NewDevConfigProvider()
	if err != nil {
		t.Fatal(err)
	}
	if len(dev.GetWebhookURLs()) != 1 || dev.GetWebhookSecret() != "k1" {
		t.Errorf("dev = %v / %q", dev.GetWebhookURLs(), dev.GetWebhookSecret())
	}
}
