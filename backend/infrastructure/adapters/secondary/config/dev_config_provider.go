package config

import (
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

type DevConfigProvider struct {
	port              string
	username          string
	password          string
	rootDir           string
	grpcPort          string
	enableGRPC        bool
	maxUploadBytes    int64
	defaultQuotaBytes int64
	jwtSecret         string
	jwtTtlSeconds     int
	storageBackend    string
	s3                ports.S3Settings
}

func NewDevConfigProvider() (*DevConfigProvider, error) {
	rootDir, err := resolveRootDir()
	if err != nil {
		return nil, err
	}

	return &DevConfigProvider{
		port:              resolvePort("3000"),
		username:          resolveUsername(),
		password:          resolvePassword(),
		rootDir:           rootDir,
		grpcPort:          resolveGRPCPort(),
		enableGRPC:        resolveEnableGRPC(),
		maxUploadBytes:    resolveMaxUploadBytes(),
		defaultQuotaBytes: resolveDefaultQuotaBytes(),
		jwtSecret:         resolveJWTSecret(),
		jwtTtlSeconds:     resolveJWTTTLSeconds(),
		storageBackend:    resolveStorageBackend(),
		s3:                resolveS3Settings(),
	}, nil
}

func (p *DevConfigProvider) GetPort() string             { return p.port }
func (p *DevConfigProvider) GetUsername() string         { return p.username }
func (p *DevConfigProvider) GetPassword() string         { return p.password }
func (p *DevConfigProvider) GetRootDir() string          { return p.rootDir }
func (p *DevConfigProvider) GetMaxUploadBytes() int64    { return p.maxUploadBytes }
func (p *DevConfigProvider) GetDefaultQuotaBytes() int64 { return p.defaultQuotaBytes }
func (p *DevConfigProvider) GetGRPCPort() string         { return p.grpcPort }
func (p *DevConfigProvider) EnableGRPC() bool            { return p.enableGRPC }
func (p *DevConfigProvider) EnableTLS() bool             { return resolveBoolEnv("ENABLE_TLS", false) }
func (p *DevConfigProvider) EnableGRPCTLS() bool         { return resolveBoolEnv("ENABLE_GRPC_TLS", resolveBoolEnv("ENABLE_TLS", false)) }
func (p *DevConfigProvider) EnableAuth() bool            { return resolveBoolEnv("ENABLE_AUTH", false) }
func (p *DevConfigProvider) EnableSignup() bool          { return resolveEnableSignup() }
func (p *DevConfigProvider) GetJWTSecret() string        { return p.jwtSecret }
func (p *DevConfigProvider) GetJWTTTLSeconds() int       { return p.jwtTtlSeconds }
func (p *DevConfigProvider) GetStorageBackend() string   { return p.storageBackend }
func (p *DevConfigProvider) GetS3Settings() ports.S3Settings {
	return p.s3
}

var _ ports.ConfigProvider = (*DevConfigProvider)(nil)
