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
}
