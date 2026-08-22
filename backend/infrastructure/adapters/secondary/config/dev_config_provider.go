package config

import (
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

type DevConfigProvider struct {
	port           string
	username       string
	password       string
	rootDir        string
	jwtSecret      string
	maxUploadBytes int64
}

func NewDevConfigProvider() (*DevConfigProvider, error) {
	rootDir, err := resolveRootDir()
	if err != nil {
		return nil, err
	}

	return &DevConfigProvider{
		port:           resolvePort("3000"),
		username:       resolveUsername(),
		password:       resolvePassword(),
		rootDir:        rootDir,
		jwtSecret:      resolveJWTSecret("dev-secret-change-me"),
		maxUploadBytes: resolveMaxUploadBytes(),
	}, nil
}

func (p *DevConfigProvider) GetPort() string           { return p.port }
func (p *DevConfigProvider) GetUsername() string       { return p.username }
func (p *DevConfigProvider) GetPassword() string       { return p.password }
func (p *DevConfigProvider) GetRootDir() string       { return p.rootDir }
func (p *DevConfigProvider) GetJWTSecret() string      { return p.jwtSecret }
func (p *DevConfigProvider) GetMaxUploadBytes() int64  { return p.maxUploadBytes }
func (p *DevConfigProvider) EnableTLS() bool           { return resolveBoolEnv("ENABLE_TLS", false) }
func (p *DevConfigProvider) EnableAuth() bool          { return resolveBoolEnv("ENABLE_AUTH", false) }

var _ ports.ConfigProvider = (*DevConfigProvider)(nil)
