package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/EslamYasser-Dev/simple-file-share/application/services"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/primary/http/dto"
)

// SharesHandler manages public share links: create (POST), list (GET) and
// revoke (DELETE) on /api/shares. All methods require authentication.
type SharesHandler struct {
	create *services.CreateShareService
	list   *services.ListSharesService
	revoke *services.RevokeShareService
	audit  *services.AuditService
}

func NewSharesHandler(create *services.CreateShareService, list *services.ListSharesService, revoke *services.RevokeShareService) *SharesHandler {
	return &SharesHandler{create: create, list: list, revoke: revoke}
}

// SetAudit attaches the security audit trail (nil disables recording).
func (h *SharesHandler) SetAudit(a *services.AuditService) { h.audit = a }

func (h *SharesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.handleCreate(w, r)
	case http.MethodGet:
		h.handleList(w, r)
	case http.MethodDelete:
		h.handleRevoke(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *SharesHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateShareRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	share, err := h.create.Execute(currentUser(r), req.Path, services.SharePolicy{
		ExpiresInSeconds: req.ExpiresInSeconds,
		Password:         req.Password,
		MaxDownloads:     req.MaxDownloads,
	})
	if err != nil {
		respondWithError(w, err)
		return
	}
	// Short feature summary — never the password itself.
	features := []string{}
	if share.PasswordProtected() {
		features = append(features, "password")
	}
	if share.Limited() {
		features = append(features, "limit")
	}
	h.audit.Record(services.AuditShareCreate, actorName(r), clientIP(r), share.Path, strings.Join(features, ","))
	respondJSON(w, http.StatusCreated, dto.FromShare(share))
}

func (h *SharesHandler) handleList(w http.ResponseWriter, r *http.Request) {
	page, err := parsePaging(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid cursor")
		return
	}
	shares, err := h.list.Execute(currentUser(r))
	if err != nil {
		respondWithError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, window(page, dto.FromShares(shares)))
}

func (h *SharesHandler) handleRevoke(w http.ResponseWriter, r *http.Request) {
	var req dto.RevokeShareRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.revoke.Execute(currentUser(r), req.Token); err != nil {
		respondWithError(w, err)
		return
	}
	// The share token is a secret; log its prefix only.
	target := req.Token
	if len(target) > 8 {
		target = target[:8] + "…"
	}
	h.audit.Record(services.AuditShareRevoke, actorName(r), clientIP(r), target, "")
	respondJSON(w, http.StatusOK, dto.MessageResponse{Message: "revoked"})
}
