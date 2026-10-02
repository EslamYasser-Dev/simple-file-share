package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

const (
	defaultPort = "22010"
	// defaultMaxUploadBytes = 0 means unlimited: uploads are only bounded by
	// available disk space. Set MAX_UPLOAD_BYTES (e.g. "2GB", "500MB") to cap.
	defaultMaxUploadBytes int64 = 0
	// defaultDefaultQuotaBytes = 0 means newly registered accounts are
	// unlimited. Set QUOTA_DEFAULT_BYTES (e.g. "100MB") to give every new
	// account a storage quota.
	defaultDefaultQuotaBytes int64 = 0
	// devStorageDirName is the dedicated, app-owned directory created under the
	// working directory when ROOT_DIR is unset in development. Storage must never
	// default to the working directory itself or a source folder, because the
	// server deletes and rewrites paths under the root.
	devStorageDirName = ".file-share-data"
)

// resolveRootDir returns the storage root: an explicit ROOT_DIR/FILE_SHARE_ROOT
// when set, /data in production, or a dedicated app directory in development.
// It only resolves the path; callers must validate and create it via
// fs.PrepareStorageRoot.
func resolveRootDir() (string, error) {
	for _, key := range []string{"ROOT_DIR", "FILE_SHARE_ROOT"} {
		if v := os.Getenv(key); v != "" {
			if strings.HasPrefix(v, "~") {
				return "", fmt.Errorf("%s %q: '~' is never expanded in environment variables — use an absolute path (containers: /data)", key, v)
			}
			return filepath.Abs(v)
		}
	}

	if os.Getenv("APP_ENV") == "production" {
		return "/data", nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(cwd, devStorageDirName), nil
}

func resolvePort(fallback string) string {
	return getEnvFirst([]string{"PORT"}, fallback)
}

func resolveUsername() string {
	return getEnvFirst([]string{"ADMIN_USERNAME", "USERNAME", "FILE_SHARE_USERNAME"}, "admin")
}

func resolvePassword() string {
	return getEnvFirst([]string{"ADMIN_PASSWORD", "PASSWORD", "FILE_SHARE_PASSWORD"}, "admin")
}

// resolveEnableSignup gates public self-registration. Defaults to enabled.
func resolveEnableSignup() bool {
	return resolveBoolEnv("ENABLE_SIGNUP", true)
}

// resolveEnableMetrics gates GET /metrics. Defaults to enabled: the endpoint
// guards itself (admin JWT or METRICS_TOKEN), so exposing it is safe and
// production observability should not depend on remembering a flag.
func resolveEnableMetrics() bool {
	return resolveBoolEnv("ENABLE_METRICS", true)
}

// resolveMetricsToken is the optional static bearer token for /metrics.
func resolveMetricsToken() string {
	return strings.TrimSpace(os.Getenv("METRICS_TOKEN"))
}

// resolveEnablePprof gates the loopback pprof side listener. Defaults to off:
// it is a debugging tool and must never be enabled by accident.
func resolveEnablePprof() bool {
	return resolveBoolEnv("ENABLE_PPROF", false)
}

// resolveEnableIndexSnapshot gates the persisted text-index snapshot.
// Defaults on: boot skips re-extraction of unchanged files. INDEX_SNAPSHOT
// =false restores a plain full rebuild every start.
func resolveEnableIndexSnapshot() bool {
	return resolveBoolEnv("INDEX_SNAPSHOT", true)
}

// resolveEnableAuditLog gates the security audit trail. Defaults on: an
// ops deployment should not lose login/admin history to a forgotten flag.
func resolveEnableAuditLog() bool {
	return resolveBoolEnv("AUDIT_LOG", true)
}

// resolveAuditMaxBytes is the rotation budget of the active audit file
// (human sizes such as "16MB" are accepted).
func resolveAuditMaxBytes() int64 {
	return resolveSizeEnv("AUDIT_MAX_BYTES", 8<<20)
}

// resolveAuditKeep is the number of rotated audit generations retained.
func resolveAuditKeep() int {
	raw := strings.TrimSpace(os.Getenv("AUDIT_KEEP"))
	if raw == "" {
		return 3
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 {
		return 3
	}
	if n > 10 {
		n = 10
	}
	return n
}

// resolveEnableTwoFactor gates the TOTP enrollment routes. Defaults on:
// accounts are only protected after opting in, so offering the feature is
// safe; =false blocks new enrollments (existing enrolled accounts still
// enforce their second factor at login).
func resolveEnableTwoFactor() bool {
	return resolveBoolEnv("ENABLE_2FA", true)
}

// resolveEnableAPIKeys gates the API key endpoints and "sfs_…" bearer
// credentials. Defaults on; =false hides the routes and rejects every key.
func resolveEnableAPIKeys() bool {
	return resolveBoolEnv("ENABLE_API_KEYS", true)
}

// resolveWebhookURLs parses WEBHOOK_URLS (comma separated). Only http(s)
// endpoints with a host survive parsing, so a typo or an unexpected scheme is
// dropped at startup instead of becoming a surprising outbound request.
// Empty (the default) disables webhooks entirely.
func resolveWebhookURLs() []string {
	raw := strings.TrimSpace(os.Getenv("WEBHOOK_URLS"))
	if raw == "" {
		return nil
	}
	var out []string
	for _, part := range strings.Split(raw, ",") {
		candidate := strings.TrimSpace(part)
		if candidate == "" {
			continue
		}
		parsed, err := url.Parse(candidate)
		if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
			continue
		}
		out = append(out, candidate)
	}
	return out
}

// resolveWebhookSecret is the HMAC-SHA256 signing key for webhook deliveries.
// Empty means deliveries are sent unsigned.
func resolveWebhookSecret() string {
	return strings.TrimSpace(os.Getenv("WEBHOOK_SECRET"))
}

// resolveGRPCPort returns the gRPC listen port. Defaults to 50051.
func resolveGRPCPort() string {
	return getEnvFirst([]string{"GRPC_PORT"}, "50051")
}

// resolveEnableGRPC gates the gRPC primary adapter. Defaults to enabled.
func resolveEnableGRPC() bool {
	return resolveBoolEnv("ENABLE_GRPC", true)
}

// resolveMaxUploadBytes parses MAX_UPLOAD_BYTES. The value may be a plain
// number of bytes or a human size such as "500MB", "2GB", "1TB". A value of 0,
// "unlimited", or an unparseable value means no limit.
func resolveMaxUploadBytes() int64 {
	return resolveSizeEnv("MAX_UPLOAD_BYTES", defaultMaxUploadBytes)
}

// resolveDefaultQuotaBytes parses QUOTA_DEFAULT_BYTES with the same rules as
// resolveMaxUploadBytes. 0 means newly registered accounts are unlimited.
func resolveDefaultQuotaBytes() int64 {
	return resolveSizeEnv("QUOTA_DEFAULT_BYTES", defaultDefaultQuotaBytes)
}

func resolveSizeEnv(key string, fallback int64) int64 {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	lower := strings.ToLower(raw)
	if lower == "unlimited" || lower == "0" || lower == "0b" {
		return 0
	}
	if n, err := strconv.ParseInt(raw, 10, 64); err == nil {
		if n < 0 {
			return 0
		}
		return n
	}
	if n, ok := parseSize(raw); ok {
		return n
	}
	return fallback
}

// parseSize converts a human size (e.g. "2GB", "512MB", "10TB") to bytes.
func parseSize(raw string) (int64, bool) {
	s := strings.TrimSpace(strings.ToUpper(raw))
	units := []struct {
		suffix string
		mul    int64
	}{
		{"TB", 1 << 40}, {"GB", 1 << 30}, {"MB", 1 << 20},
		{"KB", 1 << 10}, {"KIB", 1 << 10}, {"B", 1},
	}
	for _, u := range units {
		if strings.HasSuffix(s, u.suffix) {
			num, err := strconv.ParseInt(strings.TrimSpace(strings.TrimSuffix(s, u.suffix)), 10, 64)
			if err == nil && num >= 0 {
				return num * u.mul, true
			}
			return 0, false
		}
	}
	return 0, false
}

func resolveBoolEnv(key string, defaultValue bool) bool {
	raw := os.Getenv(key)
	if raw == "" {
		return defaultValue
	}
	return raw != "false" && raw != "0"
}

// resolveStorageBackend returns "local" or "s3". Unknown values fall back to
// "local"; callers that need a hard failure validate the raw env themselves.
func resolveStorageBackend() string {
	raw := strings.ToLower(strings.TrimSpace(os.Getenv("STORAGE_BACKEND")))
	switch raw {
	case "", "local", "fs", "filesystem":
		return "local"
	case "s3":
		return "s3"
	default:
		return raw
	}
}

func resolveS3Settings() ports.S3Settings {
	endpoint := strings.TrimRight(strings.TrimSpace(os.Getenv("S3_ENDPOINT")), "/")
	region := strings.TrimSpace(os.Getenv("S3_REGION"))
	if region == "" {
		region = "us-east-1"
	}
	pathStyle := resolveBoolEnv("S3_PATH_STYLE", false)
	if endpoint != "" && !pathStyle {
		// Custom endpoints (MinIO, R2, Ceph) almost always need path-style.
		if !strings.Contains(endpoint, "amazonaws.com") {
			pathStyle = true
		}
	}
	return ports.S3Settings{
		Endpoint:  endpoint,
		Bucket:    strings.TrimSpace(os.Getenv("S3_BUCKET")),
		Region:    region,
		AccessKey: strings.TrimSpace(os.Getenv("S3_ACCESS_KEY")),
		SecretKey: strings.TrimSpace(os.Getenv("S3_SECRET_KEY")),
		Prefix:    strings.Trim(strings.TrimSpace(os.Getenv("S3_PREFIX")), "/"),
		PathStyle: pathStyle,
	}
}

func getEnvFirst(keys []string, fallback string) string {
	for _, key := range keys {
		if value := os.Getenv(key); value != "" {
			return value
		}
	}
	return fallback
}
