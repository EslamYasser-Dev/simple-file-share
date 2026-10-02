package xhttp

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

type recordingRecorder struct {
	route    string
	status   int
	bytes    int64
	duration time.Duration
	rpc      string
}

func (r *recordingRecorder) ObserveHTTP(method, route string, status int, d time.Duration, b int64) {
	r.route = route
	r.status = status
	r.bytes = b
	r.duration = d
}

func (r *recordingRecorder) ObserveRPC(method, code string, d time.Duration) {
	r.rpc = method + "/" + code
}

func (r *recordingRecorder) AddActiveStreams(int) {}

var _ ports.MetricsRecorder = (*recordingRecorder)(nil)

func TestObserveRequestsStampsIDAndRecords(t *testing.T) {
	rec := &recordingRecorder{}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/ping", func(w http.ResponseWriter, r *http.Request) {
		if RequestIDFromContext(r.Context()) == "" {
			t.Error("handler saw empty request id")
		}
		w.WriteHeader(http.StatusTeapot)
		_, _ = w.Write([]byte("hello"))
	})
	h := observeRequests(rec, mux)

	req := httptest.NewRequest(http.MethodGet, "/api/ping", nil)
	req.Header.Set("X-Request-Id", "trace-abc-123")
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, req)

	if got := rec2.Header().Get("X-Request-Id"); got != "trace-abc-123" {
		t.Fatalf("echoed request id = %q", got)
	}
	if rec.route != "GET /api/ping" {
		t.Fatalf("route label = %q, want mux pattern", rec.route)
	}
	if rec.status != http.StatusTeapot {
		t.Fatalf("status = %d", rec.status)
	}
	if rec.bytes != int64(len("hello")) {
		t.Fatalf("bytes = %d", rec.bytes)
	}
}

func TestObserveRequestsGeneratesIDWhenMissingOrUnsafe(t *testing.T) {
	for _, inbound := range []string{"", "bad id with spaces", "x", strings.Repeat("a", 65)} {
		rec := &recordingRecorder{}
		mux := http.NewServeMux()
		mux.HandleFunc("GET /x", func(w http.ResponseWriter, r *http.Request) {})
		h := observeRequests(rec, mux)

		req := httptest.NewRequest(http.MethodGet, "/x", nil)
		if inbound != "" {
			req.Header.Set("X-Request-Id", inbound)
		}
		rec2 := httptest.NewRecorder()
		h.ServeHTTP(rec2, req)

		out := rec2.Header().Get("X-Request-Id")
		if out == "" || out == inbound {
			t.Fatalf("inbound %q produced id %q, want a fresh generated id", inbound, out)
		}
	}
}

func TestObserveRequestsNilRecorderStillStamps(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /x", func(w http.ResponseWriter, r *http.Request) {})
	h := observeRequests(nil, mux)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/x", nil))
	if rec.Header().Get("X-Request-Id") == "" {
		t.Fatal("nil recorder must not skip the request id")
	}
}

func TestObserveRequestsUnmatchedRouteLabel(t *testing.T) {
	rec := &recordingRecorder{}
	mux := http.NewServeMux()
	// No match, no catch-all: ServeMux writes 404 with an empty pattern.
	h := observeRequests(rec, mux)
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, httptest.NewRequest(http.MethodGet, "/definitely-not-registered", nil))
	if rec.route != "unmatched" {
		t.Fatalf("route = %q, want unmatched", rec.route)
	}
	if rec.status != http.StatusNotFound {
		t.Fatalf("status = %d", rec.status)
	}
}

func TestResolveRequestIDPreservesValidInbound(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-Id", "abcDEF123_-")
	if got := resolveRequestID(req); got != "abcDEF123_-" {
		t.Fatalf("resolveRequestID = %q", got)
	}
}
