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
	// EnableGRPCTLS toggles TLS on the gRPC listener independently of the HTTP
	// listener, so an edge-terminated deployment can serve plain HTTP while the
	// TCP-proxied gRPC port still speaks TLS. Defaults to EnableTLS.
	EnableGRPCTLS() bool
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
	// EnableMetrics gates the GET /metrics scrape endpoint. The endpoint is
	// always admin/token guarded when enabled.
	EnableMetrics() bool
	// GetMetricsToken is an optional static bearer token accepted for
	// GET /metrics ("" means token auth is off; admin JWT still works).
	GetMetricsToken() string
	// EnablePprof serves net/http/pprof on a loopback-only side listener
	// (never the public port).
	EnablePprof() bool
	// EnableIndexSnapshot persists the full-text index between boots so
	// startup only re-extracts files whose size/mtime changed.
	EnableIndexSnapshot() bool
	// EnableAuditLog writes security events (logins, admin changes) to
	// ROOT_DIR/.file-share/audit.jsonl. Defaults to enabled.
	EnableAuditLog() bool
	// GetAuditMaxBytes is the rotation budget of the active audit file.
	GetAuditMaxBytes() int64
	// GetAuditKeep is how many rotated audit generations are retained.
	GetAuditKeep() int
	// EnableTwoFactor serves the TOTP enrollment endpoints. Login
	// enforcement follows account state regardless of this flag, so
	// turning it off only blocks new enrollments and the admin reset —
	// enrolled accounts keep their second factor at login. Defaults on.
	EnableTwoFactor() bool
	// EnableAPIKeys serves the self-service API key endpoints and accepts
	// "sfs_…" bearer credentials. Defaults to enabled; existing keys stop
	// authenticating when turned off.
	EnableAPIKeys() bool
	// GetWebhookURLs is the comma-list of HTTP(S) endpoints that receive
	// signed copies of application events (empty = webhooks disabled).
	GetWebhookURLs() []string
	// GetWebhookSecret signs webhook payloads with HMAC-SHA256
	// ("" = deliveries are unsigned).
	GetWebhookSecret() string
	// GetGoogleClientID lists the Google OAuth client IDs accepted as an
	// ID-token audience for mobile sign-in ("" = Google sign-in disabled).
	GetGoogleClientID() string
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
