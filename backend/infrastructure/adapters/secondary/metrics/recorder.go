package metrics

import (
	"io"
	"strconv"
	"time"
)

// Recorder is the drop-in ports.MetricsRecorder built on a Registry. It owns
// the metric families used by the HTTP and gRPC primary adapters so label
// names and buckets stay consistent across transports.
type Recorder struct {
	reg          *Registry
	httpRequests *CounterVec
	httpDuration *HistogramVec
	httpBytes    *CounterVec
	grpcRequests *CounterVec
	grpcDuration *HistogramVec
	streams      *Gauge
}

// durationBuckets cover a file-share workload: cache hits and auth rejects
// in single-digit milliseconds, uploads/downloads up to seconds, SSE streams
// beyond the top bound (they land in +Inf by design).
var durationBuckets = []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30}

// NewRecorder registers every family on reg and returns the recorder.
func NewRecorder(reg *Registry) *Recorder {
	return &Recorder{
		reg: reg,
		httpRequests: reg.NewCounter(
			"fileshare_http_requests_total",
			"HTTP requests handled, by route pattern and status class code.",
			"method", "route", "status",
		),
		httpDuration: reg.NewHistogram(
			"fileshare_http_request_duration_seconds",
			"HTTP request duration in seconds, by route pattern.",
			durationBuckets, "method", "route",
		),
		httpBytes: reg.NewCounter(
			"fileshare_http_response_bytes_total",
			"HTTP response body bytes written, by route pattern.",
			"route",
		),
		grpcRequests: reg.NewCounter(
			"fileshare_grpc_requests_total",
			"gRPC calls handled, by method and status code.",
			"method", "status",
		),
		grpcDuration: reg.NewHistogram(
			"fileshare_grpc_request_duration_seconds",
			"gRPC call duration in seconds, by method.",
			durationBuckets, "method",
		),
		streams: reg.NewGauge(
			"fileshare_event_streams",
			"Live server-sent-events subscribers.",
		).WithLabelValues(),
	}
}

// WriteText satisfies ports.MetricsExporter by delegating to the registry.
func (r *Recorder) WriteText(w io.Writer) error {
	return r.reg.WriteText(w)
}

// ObserveHTTP records one HTTP request.
func (r *Recorder) ObserveHTTP(method, route string, status int, duration time.Duration, respBytes int64) {
	if r == nil {
		return
	}
	r.httpRequests.WithLabelValues(method, route, strconv.Itoa(status)).Inc()
	r.httpDuration.WithLabelValues(method, route).Observe(duration.Seconds())
	if respBytes > 0 {
		r.httpBytes.WithLabelValues(route).Add(float64(respBytes))
	}
}

// ObserveRPC records one gRPC call.
func (r *Recorder) ObserveRPC(method, code string, duration time.Duration) {
	if r == nil {
		return
	}
	r.grpcRequests.WithLabelValues(method, code).Inc()
	r.grpcDuration.WithLabelValues(method).Observe(duration.Seconds())
}

// AddActiveStreams adjusts the live SSE subscriber gauge.
func (r *Recorder) AddActiveStreams(delta int) {
	if r == nil {
		return
	}
	r.streams.Add(float64(delta))
}
