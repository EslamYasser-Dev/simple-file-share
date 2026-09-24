package xhttp

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSecurityHeadersPresent(t *testing.T) {
	h := securityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Error("missing nosniff")
	}
	if rec.Header().Get("Referrer-Policy") != "same-origin" {
		t.Error("missing referrer policy")
	}
	if rec.Header().Get("X-Frame-Options") != "DENY" {
		t.Error("missing frame options")
	}
	if rec.Header().Get("Permissions-Policy") == "" {
		t.Error("missing permissions policy")
	}
	if rec.Header().Get("Content-Security-Policy") == "" {
		t.Error("missing content security policy")
	}
	if rec.Header().Get("Strict-Transport-Security") != "" {
		t.Error("HSTS must not be sent over plain HTTP")
	}
}

func TestSecurityHeadersHSTSWhenForwardedHTTPS(t *testing.T) {
	h := securityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	h.ServeHTTP(rec, req)
	if got := rec.Header().Get("Strict-Transport-Security"); got != "max-age=31536000" {
		t.Errorf("HSTS = %q, want max-age=31536000", got)
	}
}

func TestSecurityHeadersCSPSkippedForSwagger(t *testing.T) {
	h := securityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/swagger", nil))
	if rec.Header().Get("Content-Security-Policy") != "" {
		t.Error("CSP must be skipped for the swagger UI (CDN assets)")
	}
	if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Error("other security headers must still apply on /swagger")
	}
}

func TestReadHeaderTimeoutIsBounded(t *testing.T) {
	server := NewServer("0", noopTLS{}, dumbLogger{}, RouteHandlers{}, nil, nil, false)
	if server.httpServer.ReadHeaderTimeout <= 0 {
		t.Fatalf("ReadHeaderTimeout = %v, want > 0 (slowloris protection)", server.httpServer.ReadHeaderTimeout)
	}
}

func TestPublicAuthRoutesAreRateLimited(t *testing.T) {
	server := NewServer("0", noopTLS{}, dumbLogger{}, RouteHandlers{
		Token: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}),
	}, nil, nil, false)
	server.authLimiter = NewIPLimiter(1, 2)

	mux := server.registerRoutes()
	srv := httptest.NewServer(mux)
	defer srv.Close()

	limited := false
	for i := 0; i < 10; i++ {
		resp, err := http.Post(srv.URL+"/api/auth/token", "application/json", nil)
		if err != nil {
			t.Fatal(err)
		}
		code := resp.StatusCode
		resp.Body.Close()
		if code == http.StatusTooManyRequests {
			limited = true
		}
	}
	if !limited {
		t.Fatal("expected 429 on /api/auth/token after burst")
	}
}

func TestSwaggerGatedByEnv(t *testing.T) {
	server := NewServer("0", noopTLS{}, dumbLogger{}, RouteHandlers{}, nil, nil, false)

	t.Setenv("ENABLE_SWAGGER", "false")
	t.Setenv("APP_ENV", "development")
	mux := server.registerRoutes()
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/swagger", nil))
	if rec.Code != http.StatusNotFound {
		t.Errorf("swagger disabled status = %d, want 404", rec.Code)
	}
	recYAML := httptest.NewRecorder()
	mux.ServeHTTP(recYAML, httptest.NewRequest(http.MethodGet, "/swagger.yaml", nil))
	if recYAML.Code != http.StatusNotFound {
		t.Errorf("swagger.yaml disabled status = %d, want 404", recYAML.Code)
	}

	t.Setenv("ENABLE_SWAGGER", "true")
	mux2 := server.registerRoutes()
	rec2 := httptest.NewRecorder()
	mux2.ServeHTTP(rec2, httptest.NewRequest(http.MethodGet, "/swagger", nil))
	if rec2.Code != http.StatusOK {
		t.Errorf("swagger enabled status = %d, want 200", rec2.Code)
	}
}
