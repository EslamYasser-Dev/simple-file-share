package ports

// ConfigProvider defines the interface for configuration providers.
type ConfigProvider interface {
	GetPort() string
	GetUsername() string
	GetPassword() string
	GetRootDir() string
	GetJWTSecret() string
	GetMaxUploadBytes() int64
	EnableTLS() bool
	EnableAuth() bool
}
