package config

import (
	"os"
	"path/filepath"

	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

// DevConfigProvider provides development configuration
// that disables HTTPS and uses HTTP for local development
type DevConfigProvider struct {
	port     string
	username string
	password string
	rootDir  string
	enableTLS bool
}

// NewDevConfigProvider creates a development configuration provider
// that disables TLS by default for local development
func NewDevConfigProvider() (*DevConfigProvider, error) {
	rootDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// For development, use the frontend directory as the root for file operations.
	// When running via `make run` from the project root, the correct path is "frontend".
	// If that doesn't exist, fall back to "../frontend".
	candidateA := filepath.Join(rootDir, "frontend")
	candidateB := filepath.Join(rootDir, "..", "frontend")
	devRoot := candidateA
	if _, statErr := os.Stat(candidateA); os.IsNotExist(statErr) {
		if _, statErrB := os.Stat(candidateB); !os.IsNotExist(statErrB) {
			devRoot = candidateB
		} else {
			// As a last resort, keep using rootDir to avoid crashing; logs elsewhere will warn.
			devRoot = rootDir
		}
	}

	return &DevConfigProvider{
		port:     getEnv("PORT", "3000"),
		username: getEnv("USERNAME", "admin"),
		password: getEnv("PASSWORD", "admin"),
		rootDir:  devRoot,
		enableTLS: false, // Disable TLS in development
	}, nil
}

func (p *DevConfigProvider) GetPort() string     { return p.port }
func (p *DevConfigProvider) GetUsername() string { return p.username }
func (p *DevConfigProvider) GetPassword() string { return p.password }
func (p *DevConfigProvider) GetRootDir() string  { return p.rootDir }
func (p *DevConfigProvider) EnableTLS() bool     { 
	// Always disable TLS in development
	return false 
}

var _ ports.ConfigProvider = (*DevConfigProvider)(nil)
