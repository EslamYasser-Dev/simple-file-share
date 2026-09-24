package xhttp

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestIPLimiterAllowsBurstThenLimits(t *testing.T) {
	limiter := NewIPLimiter(10, 3)
	now := time.Now()
	limiter.now = func() time.Time { return now }

	// Helper: create a request with the given remote address.
	makeReq := func(ra string) *http.Request {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = ra
		return req
	}

	for i := 0; i < 3; i++ {
		if !limiter.Allow(makeReq("10.0.0.1")) {
			t.Fatalf("request %d: expected allowed within burst", i)
		}
	}
	if limiter.Allow(makeReq("10.0.0.1")) {
		t.Fatal("expected request beyond burst to be rejected")
	}

	// Advance the clock 1s -> tokens refill by the rate.
	now = now.Add(time.Second)
	if !limiter.Allow(makeReq("10.0.0.1")) {
		t.Fatal("expected a token after refill")
	}
}

func TestIPLimiterIsPerAddress(t *testing.T) {
	limiter := NewIPLimiter(1, 1)
	makeReq := func(ra string) *http.Request {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = ra
		return req
	}
	if !limiter.Allow(makeReq("10.0.0.1")) {
		t.Fatal("address 1 must start with its own full budget")
	}
	if limiter.Allow(makeReq("10.0.0.1")) {
		t.Fatal("address 1 must be out of budget")
	}
	if !limiter.Allow(makeReq("10.0.0.2")) {
		t.Fatal("address 2 must start with its own full budget")
	}
}

func TestRateLimitMiddlewareRejectsExcess(t *testing.T) {
	limiter := NewIPLimiter(100, 1)
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := RateLimitMiddleware(limiter)(inner)

	first := httptest.NewRequest(http.MethodGet, "/", nil)
	first.RemoteAddr = "127.0.0.1:5000"
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, first)
	if rec1.Code != http.StatusOK {
		t.Fatalf("first request = %d, want 200", rec1.Code)
	}

	second := httptest.NewRequest(http.MethodGet, "/", nil)
	second.RemoteAddr = "127.0.0.1:5000"
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, second)
	if rec2.Code != http.StatusTooManyRequests {
		t.Fatalf("second request = %d, want 429", rec2.Code)
	}
}

func TestClientIPStripsPort(t *testing.T) {
	cases := []struct {
		remoteAddr string
		want       string
	}{
		{"127.0.0.1:5000", "127.0.0.1"},
		{"10.0.0.7:443", "10.0.0.7"},
		{"[::1]:8080", "::1"},
		{"no-port", "no-port"},
		{"", ""},
	}
	for _, tc := range cases {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = tc.remoteAddr
		if got := ClientIP(req); got != tc.want {
			t.Errorf("ClientIP(%q) = %q, want %q", tc.remoteAddr, got, tc.want)
		}
	}
}

// TestRateLimitMiddlewareIsKeyedByHost not proving the IPv6 case directly, but
// proving different ports on the same host share a bucket.
func TestRateLimitMiddlewareShareBucketAcrossPorts(t *testing.T) {
	limiter := NewIPLimiter(1, 2)
	handler := RateLimitMiddleware(limiter)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "127.0.0.1:5000"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("req %d on port 5000 = %d, want 200", i, rec.Code)
		}
	}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "127.0.0.1:6000" // same host, new port/connection
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("req on new port = %d, want 429 (bucket must be keyed by host)", rec.Code)
	}
}

func TestRateLimitMiddlewarePassesNilLimiterThrough(t *testing.T) {
	handler := RateLimitMiddleware(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("pass-through = %d, want 204", rec.Code)
	}
}
