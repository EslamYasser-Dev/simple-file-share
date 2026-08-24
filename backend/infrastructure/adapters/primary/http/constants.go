it cpackage xhttp

import "time"

const (
	// Server timeouts
	DefaultReadTimeout     = 30 * time.Second
	DefaultWriteTimeout    = 60 * time.Second // allow large downloads/uploads to finish
	DefaultIdleTimeout     = 120 * time.Second
	DefaultShutdownTimeout = 30 * time.Second

	// HTTP limits
	DefaultMaxHeaderBytes = 1 << 20 // 1MB
)
