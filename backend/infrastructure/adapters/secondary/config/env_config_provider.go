package config

import (
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
	"os"
	"strconv"
)

type EnvConfigProvider struct {
	port              string
	username          string
	password          string
	rootDir           string
	grpcPort          string
	maxUploadBytes    int64
	defaultQuotaBytes int64
	enableTLS         bool
	enableAuth        bool
	enableSignup      bool
	enableGRPC        bool
	jwtSecret         string
	jwtTtlSeconds     int
	storageBackend    string
	s3                ports.S3Settings
}

func NewEnvConfigProvider() (*EnvConfigProvider, error) {
	rootDir, err := resolveRootDir()
	if err != nil {
		return nil, err
	}

	return &EnvConfigProvider{
		port:              resolvePort(defaultPort),
		username:          resolveUsername(),
		password:          resolvePassword(),
		rootDir:           rootDir,
		grpcPort:          resolveGRPCPort(),
		maxUploadBytes:    resolveMaxUploadBytes(),
		defaultQuotaBytes: resolveDefaultQuotaBytes(),
		enableTLS:         resolveBoolEnv("ENABLE_TLS", true),
		enableAuth:        resolveBoolEnv("ENABLE_AUTH", true),
		enableSignup:      resolveEnableSignup(),
		enableGRPC:        resolveEnableGRPC(),
		jwtSecret:         resolveJWTSecret(),
		jwtTtlSeconds:     resolveJWTTTLSeconds(),
		storageBackend:    resolveStorageBackend(),
		s3:                resolveS3Settings(),
	}, nil
}

func (p *EnvConfigProvider) GetPort() string             { return p.port }
func (p *EnvConfigProvider) GetUsername() string         { return p.username }
func (p *EnvConfigProvider) GetPassword() string         { return p.password }
func (p *EnvConfigProvider) GetRootDir() string          { return p.rootDir }
func (p *EnvConfigProvider) GetMaxUploadBytes() int64    { return p.maxUploadBytes }
func (p *EnvConfigProvider) GetDefaultQuotaBytes() int64 { return p.defaultQuotaBytes }
func (p *EnvConfigProvider) GetGRPCPort() string         { return p.grpcPort }
func (p *EnvConfigProvider) EnableGRPC() bool            { return p.enableGRPC }
func (p *EnvConfigProvider) EnableTLS() bool             { return p.enableTLS }
func (p *EnvConfigProvider) EnableAuth() bool            { return p.enableAuth }
func (p *EnvConfigProvider) EnableSignup() bool          { return p.enableSignup }
func (p *EnvConfigProvider) GetJWTSecret() string        { return p.jwtSecret }
func (p *EnvConfigProvider) GetJWTTTLSeconds() int       { return p.jwtTtlSeconds }
func (p *EnvConfigProvider) GetStorageBackend() string   { return p.storageBackend }
func (p *EnvConfigProvider) GetS3Settings() ports.S3Settings {
	return p.s3
}

var _ ports.ConfigProvider = (*EnvConfigProvider)(nil)

func resolveJWTSecret() string {
	defaultSecret := "change-me-in-production"
	v := os.Getenv("JWT_SECRET")
	if v != "" {
		return v
	}
	return defaultSecret
}

func resolveJWTTTLSeconds() int {
	v := os.Getenv("JWT_TTL_SECONDS")
	if v != "" {
		n, err := strconv.Atoi(v)
		if err == nil && n > 0 {
			return n
		}
	}
	return 3600 // 1 hour default
}
