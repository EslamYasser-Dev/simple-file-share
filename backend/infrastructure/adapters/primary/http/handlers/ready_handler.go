package handlers

import (
	"net/http"

	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/primary/http/dto"
)

// ReadinessCheck is one named dependency probe for GET /health/ready.
type ReadinessCheck struct {
	Name  string
	Check func() error
}

// ReadyHandler reports whether the server can actually serve traffic, in
// contrast to /health (liveness: process up). Every failed check is listed;
// any failure yields 503 so orchestrators keep the instance out of rotation.
type ReadyHandler struct {
	checks []ReadinessCheck
}

// NewReadyHandler composes readiness from the given checks.
func NewReadyHandler(checks ...ReadinessCheck) *ReadyHandler {
	return &ReadyHandler{checks: checks}
}

func (h *ReadyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	results := make(map[string]string, len(h.checks))
	ready := true
	for _, c := range h.checks {
		if c.Check == nil {
			results[c.Name] = "ok"
			continue
		}
		if err := c.Check(); err != nil {
			results[c.Name] = "error: " + err.Error()
			ready = false
			continue
		}
		results[c.Name] = "ok"
	}

	status := "ready"
	code := http.StatusOK
	if !ready {
		status = "unavailable"
		code = http.StatusServiceUnavailable
	}
	respondJSON(w, code, dto.ReadyResponse{Status: status, Checks: results})
}
