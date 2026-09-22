package xhttp

import "time"

const (
	// Server timeouts. Both Read and Write timeouts are disabled because
	// uploads may be very large (unlimited by default) and legitimately take
	// hours; only header reads remain bounded to prevent slowloris attacks.
	DefaultReadTimeout     = 0
	DefaultWriteTimeout    = 0
	DefaultIdleTimeout     = 120 * time.Second
	DefaultShutdownTimeout = 30 * time.Second

	// HTTP limits
	DefaultMaxHeaderBytes = 1 << 20 // 1MB
)
