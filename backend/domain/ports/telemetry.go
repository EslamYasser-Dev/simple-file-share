package ports

import (
	"io"
	"time"
)

// MetricsRecorder records request and RPC telemetry. Implementations must be
// safe for concurrent use and must never panic or block on a slow consumer:
// instrumentation must not be able to take down the request path.
type MetricsRecorder interface {
	// ObserveHTTP records one finished HTTP request. route is the matched
	// route pattern (never a raw path) so label cardinality stays bounded.
	ObserveHTTP(method, route string, status int, duration time.Duration, respBytes int64)
	// ObserveRPC records one finished gRPC call; code is the status code name.
	ObserveRPC(method, code string, duration time.Duration)
	// AddActiveStreams adjusts the live server-sent-events subscriber gauge.
	AddActiveStreams(delta int)
}

// MetricsExporter renders the scrape payload for GET /metrics.
type MetricsExporter interface {
	WriteText(w io.Writer) error
}
