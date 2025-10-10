package xhttp

import "time"

const (
	// Server timeouts
	DefaultReadTimeout  = 30 * time.Second
	DefaultWriteTimeout = 30 * time.Second
	DefaultShutdownTimeout = 30 * time.Second
	
	// HTTP limits
	DefaultMaxHeaderBytes = 1 << 20 // 1MB
	
	// TLS configuration
	DefaultTLSMinVersion = 1.3
)
