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
)

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

	candidates := []string{
		filepath.Join(cwd, "data"),
		filepath.Join(cwd, "frontend"),
		filepath.Join(cwd, "..", "data"),
		cwd,
	}
	for _, candidate := range candidates {
		if _, statErr := os.Stat(candidate); statErr == nil {
			return filepath.Abs(candidate)
		}
	}
	return cwd, nil
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

func resolveJWTSecret(fallback string) string {
	return getEnvFirst([]string{"JWT_SECRET"}, fallback)
}

// resolveMaxUploadBytes parses MAX_UPLOAD_BYTES. The value may be a plain
// number of bytes or a human size such as "500MB", "2GB", "1TB". A value of 0,
// "unlimited", or an unparseable value means no limit.
func resolveMaxUploadBytes() int64 {
	raw := strings.TrimSpace(os.Getenv("MAX_UPLOAD_BYTES"))
	if raw == "" {
		return defaultMaxUploadBytes
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
	return defaultMaxUploadBytes
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
