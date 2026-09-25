package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type stubCertSource struct {
	pem []byte
	ok  bool
}

func (s stubCertSource) CertPEM() ([]byte, bool) { return s.pem, s.ok }

func TestGRPCCertHandler(t *testing.T) {
	t.Run("serves PEM when a key pair exists", func(t *testing.T) {
		h := NewGRPCCertHandler(stubCertSource{pem: []byte("-----BEGIN CERTIFICATE-----\nabc\n"), ok: true})
		req := httptest.NewRequest(http.MethodGet, "/api/grpc/cert", nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("GET = %d, want 200", rec.Code)
		}
		if got := rec.Header().Get("Content-Type"); got != "application/x-pem-file" {
			t.Fatalf("Content-Type = %q", got)
		}
		if rec.Body.String() != "-----BEGIN CERTIFICATE-----\nabc\n" {
			t.Fatalf("body = %q", rec.Body.String())
		}
	})

	t.Run("404 when gRPC TLS is disabled", func(t *testing.T) {
		h := NewGRPCCertHandler(stubCertSource{})
		req := httptest.NewRequest(http.MethodGet, "/api/grpc/cert", nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("GET = %d, want 404", rec.Code)
		}
	})

	t.Run("other methods are rejected", func(t *testing.T) {
		h := NewGRPCCertHandler(stubCertSource{pem: []byte("x"), ok: true})
		for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodDelete} {
			req := httptest.NewRequest(method, "/api/grpc/cert", nil)
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if rec.Code != http.StatusMethodNotAllowed {
				t.Fatalf("%s = %d, want 405", method, rec.Code)
			}
		}
	})
}
