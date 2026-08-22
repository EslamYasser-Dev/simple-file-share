package handlers

import (
	"net/http"
	"time"

	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/primary/http/dto"
)

var serverStartedAt = time.Now()

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	respondJSON(w, http.StatusOK, dto.HealthResponse{
		Status: "ok",
		Uptime: time.Since(serverStartedAt).Round(time.Second).String(),
	})
}
