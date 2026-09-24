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
	// GetStorageBackend returns "local" (default) or "s3".
	GetStorageBackend() string
	// GetS3Settings returns S3-compatible object-store credentials. Fields are
	// meaningful only when GetStorageBackend is "s3".
	GetS3Settings() S3Settings
}

// S3Settings describes an S3-compatible object store (AWS, MinIO, R2, GCS).
type S3Settings struct {
	Endpoint  string
	Bucket    string
	Region    string
	AccessKey string
	SecretKey string
	// Prefix is prepended to every object key (optional).
	Prefix string
	// PathStyle forces path-style addressing (required by most MinIO setups).
	PathStyle bool
}
