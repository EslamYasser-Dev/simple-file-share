package handlers

import (
	"net/http"
	"strings"

	"github.com/EslamYasser-Dev/simple-file-share/application/services"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

// AdminAuditHandler serves GET /api/admin/audit — one page of the security
// trail, newest first, with an opaque cursor for the next page:
//
//	GET /api/admin/audit?action=login.fail&actor=alice&limit=50&cursor=...
//	=> {"items":[...],"nextCursor":"..."}
type AdminAuditHandler struct {
	audit *services.AuditService
}

// NewAdminAuditHandler constructs the query endpoint (nil disables it).
func NewAdminAuditHandler(audit *services.AuditService) *AdminAuditHandler {
	return &AdminAuditHandler{audit: audit}
}

type auditPage struct {
	Items      []ports.AuditEntry `json:"items"`
	NextCursor string             `json:"nextCursor,omitempty"`
}

func (h *AdminAuditHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	q := ports.AuditQuery{
		Action: strings.TrimSpace(r.URL.Query().Get("action")),
		Actor:  strings.TrimSpace(r.URL.Query().Get("actor")),
		Limit:  queryInt(r, "limit", 50),
		Cursor: strings.TrimSpace(r.URL.Query().Get("cursor")),
	}
	entries, next, err := h.audit.Query(currentUser(r), q)
	if err != nil {
		respondWithError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, auditPage{Items: entries, NextCursor: next})
}
