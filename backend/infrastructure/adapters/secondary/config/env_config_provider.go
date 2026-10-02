package config

import (
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
	"os"
	"strconv"
)

type EnvConfigProvider struct {
	port                string
	username            string
	password            string
	rootDir             string
	grpcPort            string
	maxUploadBytes      int64
	defaultQuotaBytes   int64
	enableTLS           bool
	enableGRPCTLS       bool
	enableAuth          bool
	enableSignup        bool
	enableGRPC          bool
	jwtSecret           string
	jwtTtlSeconds       int
	storageBackend      string
	s3                  ports.S3Settings
	enableMetrics       bool
	metricsToken        string
	enablePprof         bool
	enableIndexSnapshot bool
	enableAuditLog      bool
	auditMaxBytes       int64
	auditKeep           int
	enableTwoFactor     bool
	enableAPIKeys       bool
	webhookURLs         []string
	webhookSecret       string
}

func NewEnvConfigProvider() (*EnvConfigProvider, error) {
	rootDir, err := resolveRootDir()
	if err != nil {
		return nil, err
	}

	enableTLS := resolveBoolEnv("ENABLE_TLS", true)

	return &EnvConfigProvider{
		port:                resolvePort(defaultPort),
		username:            resolveUsername(),
		password:            resolvePassword(),
		rootDir:             rootDir,
		grpcPort:            resolveGRPCPort(),
		maxUploadBytes:      resolveMaxUploadBytes(),
		defaultQuotaBytes:   resolveDefaultQuotaBytes(),
		enableTLS:           enableTLS,
		enableGRPCTLS:       resolveBoolEnv("ENABLE_GRPC_TLS", enableTLS),
		enableAuth:          resolveBoolEnv("ENABLE_AUTH", true),
		enableSignup:        resolveEnableSignup(),
		enableGRPC:          resolveEnableGRPC(),
		jwtSecret:           resolveJWTSecret(),
		jwtTtlSeconds:       resolveJWTTTLSeconds(),
		storageBackend:      resolveStorageBackend(),
		s3:                  resolveS3Settings(),
		enableMetrics:       resolveEnableMetrics(),
		metricsToken:        resolveMetricsToken(),
		enablePprof:         resolveEnablePprof(),
		enableIndexSnapshot: resolveEnableIndexSnapshot(),
		enableAuditLog:      resolveEnableAuditLog(),
		auditMaxBytes:       resolveAuditMaxBytes(),
		auditKeep:           resolveAuditKeep(),
		enableTwoFactor:     resolveEnableTwoFactor(),
		enableAPIKeys:       resolveEnableAPIKeys(),
		webhookURLs:         resolveWebhookURLs(),
		webhookSecret:       resolveWebhookSecret(),
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
func (p *EnvConfigProvider) EnableGRPCTLS() bool         { return p.enableGRPCTLS }
func (p *EnvConfigProvider) EnableAuth() bool            { return p.enableAuth }
func (p *EnvConfigProvider) EnableSignup() bool          { return p.enableSignup }
func (p *EnvConfigProvider) GetJWTSecret() string        { return p.jwtSecret }
func (p *EnvConfigProvider) GetJWTTTLSeconds() int       { return p.jwtTtlSeconds }
func (p *EnvConfigProvider) GetStorageBackend() string   { return p.storageBackend }
func (p *EnvConfigProvider) EnableMetrics() bool         { return p.enableMetrics }
func (p *EnvConfigProvider) GetMetricsToken() string     { return p.metricsToken }
func (p *EnvConfigProvider) EnablePprof() bool           { return p.enablePprof }
func (p *EnvConfigProvider) EnableIndexSnapshot() bool   { return p.enableIndexSnapshot }
func (p *EnvConfigProvider) EnableAuditLog() bool        { return p.enableAuditLog }
func (p *EnvConfigProvider) GetAuditMaxBytes() int64     { return p.auditMaxBytes }
func (p *EnvConfigProvider) GetAuditKeep() int           { return p.auditKeep }
func (p *EnvConfigProvider) EnableTwoFactor() bool       { return p.enableTwoFactor }
func (p *EnvConfigProvider) EnableAPIKeys() bool         { return p.enableAPIKeys }
func (p *EnvConfigProvider) GetWebhookURLs() []string    { return p.webhookURLs }
func (p *EnvConfigProvider) GetWebhookSecret() string    { return p.webhookSecret }
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
