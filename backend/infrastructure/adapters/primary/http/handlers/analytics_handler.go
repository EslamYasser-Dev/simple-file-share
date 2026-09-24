package handlers

import (
	"net/http"
	"strconv"

	"github.com/EslamYasser-Dev/simple-file-share/application/services"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

// AdminAnalyticsHandler serves:
//
//	GET /api/admin/analytics/overview
//	GET /api/admin/analytics/timeline
//	GET /api/admin/analytics/top-files
type AdminAnalyticsHandler struct {
	svc *services.AnalyticsService
}

func NewAdminAnalyticsHandler(svc *services.AnalyticsService) *AdminAnalyticsHandler {
	return &AdminAnalyticsHandler{svc: svc}
}

func (h *AdminAnalyticsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	days := queryInt(r, "days", 30)
	actor := currentUser(r)

	switch r.PathValue("report") {
	case "overview":
		ov, err := h.svc.Overview(actor, days)
		if err != nil {
			respondWithError(w, err)
			return
		}
		respondJSON(w, http.StatusOK, ov)
	case "timeline":
		buckets, err := h.svc.Timeline(actor, days)
		if err != nil {
			respondWithError(w, err)
			return
		}
		if buckets == nil {
			buckets = make([]ports.AnalyticsBucket, 0)
		}
		respondJSON(w, http.StatusOK, buckets)
	case "top-files":
		limit := queryInt(r, "limit", 10)
		top, err := h.svc.TopFiles(actor, days, limit)
		if err != nil {
			respondWithError(w, err)
			return
		}
		respondJSON(w, http.StatusOK, top)
	default:
		respondError(w, http.StatusNotFound, "unknown analytics report")
	}
}

func queryInt(r *http.Request, key string, fallback int) int {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return fallback
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}
