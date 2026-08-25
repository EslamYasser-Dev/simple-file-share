package config

import "github.com/EslamYasser-Dev/simple-file-share/domain/ports"

type EnvConfigProvider struct {
	port           string
	username       string
	password       string
	rootDir        string
	jwtSecret      string
	maxUploadBytes int64
	enableTLS      bool
	enableAuth     bool
}

func NewEnvConfigProvider() (*EnvConfigProvider, error) {
	rootDir, err := resolveRootDir()
	if err != nil {
		return nil, err
	}

	return &EnvConfigProvider{
		port:           resolvePort(defaultPort),
		username:       resolveUsername(),
		password:       resolvePassword(),
		rootDir:        rootDir,
		jwtSecret:      resolveJWTSecret("change-me-in-production"),
		maxUploadBytes: resolveMaxUploadBytes(),
		enableTLS:      resolveBoolEnv("ENABLE_TLS", true),
		enableAuth:     resolveBoolEnv("ENABLE_AUTH", true),
	}, nil
}

func (p *EnvConfigProvider) GetPort() string          { return p.port }
func (p *EnvConfigProvider) GetUsername() string      { return p.username }
func (p *EnvConfigProvider) GetPassword() string      { return p.password }
func (p *EnvConfigProvider) GetRootDir() string       { return p.rootDir }
func (p *EnvConfigProvider) GetJWTSecret() string     { return p.jwtSecret }
func (p *EnvConfigProvider) GetMaxUploadBytes() int64 { return p.maxUploadBytes }
func (p *EnvConfigProvider) EnableTLS() bool          { return p.enableTLS }
func (p *EnvConfigProvider) EnableAuth() bool         { return p.enableAuth }

var _ ports.ConfigProvider = (*EnvConfigProvider)(nil)
