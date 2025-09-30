package ports

// ConfigProvider defines the interface for configuration providers
type ConfigProvider interface {
	GetPort() string
	GetUsername() string
	GetPassword() string
	GetRootDir() string
	// EnableTLS returns whether TLS should be enabled
	EnableTLS() bool
}
