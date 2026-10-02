package xhttp

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"regexp"
	"time"

	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

type requestIDKey struct{}

// inboundRequestID is the shape we echo back: short, opaque, header-safe.
var inboundRequestID = regexp.MustCompile(`^[A-Za-z0-9._-]{8,64}$`)

// RequestIDFromContext returns the request id stamped by the observability
// wrapper, or "" when served outside it (raw handler tests).
func RequestIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey{}).(string)
	return id
}

// resolveRequestID echoes a well-formed inbound X-Request-Id (so traces stitch
// across proxies) or mints a fresh 128-bit id.
func resolveRequestID(r *http.Request) string {
	if in := r.Header.Get("X-Request-Id"); inboundRequestID.MatchString(in) {
		return in
	}
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		// crypto/rand failing is fatal for TLS anyway; fall back to a
		// time-derived id so logging never panics.
		return hex.EncodeToString(time.Now().AppendFormat(nil, "20060102150405.000000000"))
	}
	return hex.EncodeToString(b[:])
}

// normalizeRoute maps the mux-matched pattern to a bounded metric label.
// Requests that never match a pattern collapse to "unmatched" instead of
// leaking raw paths (which contain share tokens and usernames).
func normalizeRoute(pattern string) string {
	if pattern == "" {
		return "unmatched"
	}
	return pattern
}

// observeRequests is the outermost wrapper under the security headers: it
// stamps the request id, measures the whole mux, and records telemetry. It is
// applied even without a recorder so every response carries X-Request-Id.
func observeRequests(recorder ports.MetricsRecorder, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := resolveRequestID(r)
		req := r.WithContext(context.WithValue(r.Context(), requestIDKey{}, id))
		w.Header().Set("X-Request-Id", id)

		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, req)

		if recorder != nil {
			recorder.ObserveHTTP(
				req.Method,
				normalizeRoute(req.Pattern),
				rw.status,
				time.Since(start),
				int64(rw.bytesWritten),
			)
		}
	})
}
