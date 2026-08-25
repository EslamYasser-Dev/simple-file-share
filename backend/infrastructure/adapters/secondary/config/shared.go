package config

import (
	"os"
	"path/filepath"
	"strconv"
)

const (
	defaultPort           = "22010"
	defaultMaxUploadBytes = 100 << 20 // 100 MiB
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
	return getEnvFirst([]string{"USERNAME", "FILE_SHARE_USERNAME"}, "admin")
}

func resolvePassword() string {
	return getEnvFirst([]string{"PASSWORD", "FILE_SHARE_PASSWORD"}, "admin")
}

func resolveJWTSecret(fallback string) string {
	return getEnvFirst([]string{"JWT_SECRET"}, fallback)
}

func resolveMaxUploadBytes() int64 {
	raw := os.Getenv("MAX_UPLOAD_BYTES")
	if raw == "" {
		return defaultMaxUploadBytes
	}
	n, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || n <= 0 {
		return defaultMaxUploadBytes
	}
	return n
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
