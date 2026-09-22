package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
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
	// default to the working directory itself or a source folder (e.g.
	// frontend/), because the server deletes and rewrites paths under the root.
	devStorageDirName = ".file-share-data"
)

// resolveRootDir returns the storage root: an explicit ROOT_DIR/FILE_SHARE_ROOT
// when set, /data in production, or a dedicated app directory in development.
// It only resolves the path; callers must validate and create it via
// fs.PrepareStorageRoot.
func resolveRootDir() (string, error) {
	for _, key := range []string{"ROOT_DIR", "FILE_SHARE_ROOT"} {
		if v := os.Getenv(key); v != "" {
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

func getEnvFirst(keys []string, fallback string) string {
	for _, key := range keys {
		if value := os.Getenv(key); value != "" {
			return value
		}
	}
	return fallback
}
