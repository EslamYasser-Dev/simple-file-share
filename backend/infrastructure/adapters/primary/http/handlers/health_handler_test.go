package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandler(t *testing.T) {
	h := NewHealthHandler()

	t.Run("GET returns 200 with status ok", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("GET /health = %d, want 200", rec.Code)
		}
		if body := rec.Body.String(); body == "" {
			t.Fatal("GET /health returned empty body")
		}
	})

	t.Run("HEAD returns 200 for standard probes", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodHead, "/health", nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("HEAD /health = %d, want 200", rec.Code)
		}
	})

	t.Run("other methods are rejected", func(t *testing.T) {
		for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodDelete} {
			req := httptest.NewRequest(method, "/health", nil)
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if rec.Code != http.StatusMethodNotAllowed {
				t.Fatalf("%s /health = %d, want 405", method, rec.Code)
			}
		}
	})
}
