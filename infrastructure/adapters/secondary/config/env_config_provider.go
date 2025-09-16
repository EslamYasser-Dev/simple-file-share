package config

import (
	"os"

	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

// EnvConfigProvider reads configuration from environment variables.
type EnvConfigProvider struct {
	port     string
	username string
	password string
	rootDir  string
}

// NewEnvConfigProvider creates a config provider with defaults.
func NewEnvConfigProvider() (*EnvConfigProvider, error) {
	rootDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	return &EnvConfigProvider{
		port:     getEnv("PORT", "22010"),
		username: getEnv("USERNAME", "admin"),
		password: getEnv("PASSWORD", "thisone"),
		rootDir:  rootDir,
	}, nil
}

func (p *EnvConfigProvider) GetPort() string     { return p.port }
func (p *EnvConfigProvider) GetUsername() string { return p.username }
func (p *EnvConfigProvider) GetPassword() string { return p.password }
func (p *EnvConfigProvider) GetRootDir() string  { return p.rootDir }

// getEnv returns env var value or fallback.
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

var _ ports.ConfigProvider = (*EnvConfigProvider)(nil)
