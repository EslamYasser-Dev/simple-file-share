package xhttp

import "time"

const (
	// Server timeouts. Read and Write timeouts are disabled because uploads
	// may be very large (unlimited by default) and legitimately take hours;
	// header reads are bounded separately so slowloris cannot pin sockets
	// open with a trickle of request headers.
	DefaultReadTimeout       = 0
	DefaultReadHeaderTimeout = 10 * time.Second
	DefaultWriteTimeout      = 0
	DefaultIdleTimeout       = 120 * time.Second
	DefaultShutdownTimeout   = 30 * time.Second

	// HTTP limits
	DefaultMaxHeaderBytes = 1 << 20 // 1MB
)
