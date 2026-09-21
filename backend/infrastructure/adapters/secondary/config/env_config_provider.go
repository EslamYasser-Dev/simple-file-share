package config

import "github.com/EslamYasser-Dev/simple-file-share/domain/ports"

type EnvConfigProvider struct {
	port           string
	username       string
	password       string
	rootDir        string
	grpcPort       string
	maxUploadBytes int64
	enableTLS      bool
	enableAuth     bool
	enableSignup   bool
	enableGRPC     bool
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
		grpcPort:       resolveGRPCPort(),
		maxUploadBytes: resolveMaxUploadBytes(),
		enableTLS:      resolveBoolEnv("ENABLE_TLS", true),
		enableAuth:     resolveBoolEnv("ENABLE_AUTH", true),
		enableSignup:   resolveEnableSignup(),
		enableGRPC:     resolveEnableGRPC(),
	}, nil
}

func (p *EnvConfigProvider) GetPort() string          { return p.port }
func (p *EnvConfigProvider) GetUsername() string      { return p.username }
func (p *EnvConfigProvider) GetPassword() string      { return p.password }
func (p *EnvConfigProvider) GetRootDir() string       { return p.rootDir }
func (p *EnvConfigProvider) GetMaxUploadBytes() int64 { return p.maxUploadBytes }
func (p *EnvConfigProvider) GetGRPCPort() string      { return p.grpcPort }
func (p *EnvConfigProvider) EnableGRPC() bool         { return p.enableGRPC }
func (p *EnvConfigProvider) EnableTLS() bool          { return p.enableTLS }
func (p *EnvConfigProvider) EnableAuth() bool         { return p.enableAuth }
func (p *EnvConfigProvider) EnableSignup() bool       { return p.enableSignup }

var _ ports.ConfigProvider = (*EnvConfigProvider)(nil)
