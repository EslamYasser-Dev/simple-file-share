package ports

// ConfigProvider defines the interface for configuration providers.
type ConfigProvider interface {
	GetPort() string
	GetUsername() string
	GetPassword() string
	GetRootDir() string
	GetMaxUploadBytes() int64
	// GetGRPCPort is the TCP port the gRPC server listens on.
	GetGRPCPort() string
	// EnableGRPC toggles the gRPC primary adapter.
	EnableGRPC() bool
	EnableTLS() bool
	EnableAuth() bool
	// EnableSignup toggles public self-registration via POST /api/auth/register.
	EnableSignup() bool
	// GetDefaultQuotaBytes is the storage quota given to newly registered
	// accounts. 0 means unlimited.
	GetDefaultQuotaBytes() int64
	// GetJWTSecret is the HMAC key for stateless access tokens.
	GetJWTSecret() string
	// GetJWTTTLSeconds is the access-token lifetime in seconds.
	GetJWTTTLSeconds() int
}
