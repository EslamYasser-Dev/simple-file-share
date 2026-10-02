package handlers

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

// MetricsHandler serves GET /metrics in Prometheus text format.
//
// Access control (checked in order): the optional METRICS_TOKEN bearer; then,
// when authentication is off, the local system view; then an authenticated
// admin. Unauthenticated callers get 401, authenticated non-admins 403.
type MetricsHandler struct {
	exporter    ports.MetricsExporter
	token       string
	authEnabled bool
	resolveUser func(*http.Request) (*models.User, error)
}

// NewMetricsHandler builds the scrape endpoint. resolveUser may be nil only
// when authEnabled is false.
func NewMetricsHandler(
	exporter ports.MetricsExporter,
	token string,
	authEnabled bool,
	resolveUser func(*http.Request) (*models.User, error),
) *MetricsHandler {
	return &MetricsHandler{
		exporter:    exporter,
		token:       token,
		authEnabled: authEnabled,
		resolveUser: resolveUser,
	}
}

func (h *MetricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if h.exporter == nil {
		http.Error(w, "metrics unavailable", http.StatusServiceUnavailable)
		return
	}

	switch {
	case h.tokenAuthorized(r):
		h.write(w)
		return
	case !h.authEnabled:
		// System view: no accounts exist, so admin-ness is implicit.
		h.write(w)
		return
	case h.resolveUser == nil:
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.resolveUser(r)
	if err != nil || user == nil {
		w.Header().Set("WWW-Authenticate", `Bearer realm="metrics"`)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if !user.IsAdmin {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	h.write(w)
}

func (h *MetricsHandler) write(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_ = h.exporter.WriteText(w)
}

// tokenAuthorized accepts METRICS_TOKEN as a constant-time bearer comparison.
func (h *MetricsHandler) tokenAuthorized(r *http.Request) bool {
	if h.token == "" {
		return false
	}
	authz := r.Header.Get("Authorization")
	const prefix = "Bearer "
	if len(authz) <= len(prefix) || !strings.EqualFold(authz[:len(prefix)], prefix) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(strings.TrimSpace(authz[len(prefix):])), []byte(h.token)) == 1
}
