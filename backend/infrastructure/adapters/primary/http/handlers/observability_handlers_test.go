package handlers

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

type stubExporter struct {
	body string
	err  error
}

func (s *stubExporter) WriteText(w io.Writer) error {
	if s.err != nil {
		return s.err
	}
	_, err := io.WriteString(w, s.body)
	return err
}

func adminUser() *models.User { return &models.User{Username: "root", IsAdmin: true} }

func memberUser() *models.User { return &models.User{Username: "eve", IsAdmin: false} }

func TestMetricsHandlerAccess(t *testing.T) {
	exporter := &stubExporter{body: "# TYPE fs_x gauge\nfs_x 1\n"}
	resolve := func(r *http.Request) (*models.User, error) {
		if r.Header.Get("Authorization") == "Bearer admin-jwt" {
			return adminUser(), nil
		}
		if r.Header.Get("Authorization") == "Bearer member-jwt" {
			return memberUser(), nil
		}
		return nil, ports.ErrUnauthorized
	}

	cases := []struct {
		name       string
		token      string
		authOn     bool
		authHeader string
		want       int
	}{
		{"token match", "scraper-secret", true, "Bearer scraper-secret", http.StatusOK},
		{"token mismatch falls through to auth", "scraper-secret", true, "Bearer wrong", http.StatusUnauthorized},
		{"admin jwt", "", true, "Bearer admin-jwt", http.StatusOK},
		{"member jwt forbidden", "", true, "Bearer member-jwt", http.StatusForbidden},
		{"no credentials", "", true, "", http.StatusUnauthorized},
		{"auth disabled is open", "", false, "", http.StatusOK},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := NewMetricsHandler(exporter, tc.token, tc.authOn, resolve)
			req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if rec.Code != tc.want {
				t.Fatalf("status = %d, want %d", rec.Code, tc.want)
			}
			if tc.want == http.StatusOK {
				if got := rec.Header().Get("Content-Type"); !strings.Contains(got, "text/plain") {
					t.Fatalf("content-type = %q", got)
				}
				if !strings.Contains(rec.Body.String(), "fs_x 1") {
					t.Fatalf("body = %q", rec.Body.String())
				}
			}
		})
	}
}

func TestMetricsHandlerMethodAndExporterFailure(t *testing.T) {
	h := NewMetricsHandler(&stubExporter{body: "x"}, "", false, nil)

	req := httptest.NewRequest(http.MethodPost, "/metrics", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST = %d, want 405", rec.Code)
	}

	nilExporter := NewMetricsHandler(nil, "", false, nil)
	rec = httptest.NewRecorder()
	nilExporter.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("nil exporter = %d, want 503", rec.Code)
	}
}

func TestReadyHandler(t *testing.T) {
	t.Run("all checks pass", func(t *testing.T) {
		h := NewReadyHandler(
			ReadinessCheck{Name: "storage", Check: func() error { return nil }},
			ReadinessCheck{Name: "index", Check: func() error { return nil }},
		)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health/ready", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
		}
		if !strings.Contains(rec.Body.String(), `"status":"ready"`) ||
			!strings.Contains(rec.Body.String(), `"storage":"ok"`) {
			t.Fatalf("body = %s", rec.Body.String())
		}
	})

	t.Run("failed check yields 503", func(t *testing.T) {
		h := NewReadyHandler(
			ReadinessCheck{Name: "storage", Check: func() error { return errors.New("read-only") }},
			ReadinessCheck{Name: "index", Check: func() error { return nil }},
		)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health/ready", nil))
		if rec.Code != http.StatusServiceUnavailable {
			t.Fatalf("status = %d, want 503", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "read-only") {
			t.Fatalf("error detail missing: %s", rec.Body.String())
		}
	})

	t.Run("method not allowed", func(t *testing.T) {
		h := NewReadyHandler()
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/health/ready", nil))
		if rec.Code != http.StatusMethodNotAllowed {
			t.Fatalf("status = %d, want 405", rec.Code)
		}
	})
}
