package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/EslamYasser-Dev/simple-file-share/application/services"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/primary/http/dto"
)

// SharesHandler manages public share links: create (POST), list (GET) and
// revoke (DELETE) on /api/shares. All methods require authentication.
type SharesHandler struct {
	create *services.CreateShareService
	list   *services.ListSharesService
	revoke *services.RevokeShareService
}

func NewSharesHandler(create *services.CreateShareService, list *services.ListSharesService, revoke *services.RevokeShareService) *SharesHandler {
	return &SharesHandler{create: create, list: list, revoke: revoke}
}

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

	share, err := h.create.Execute(currentUser(r), req.Path, req.ExpiresInSeconds)
	if err != nil {
		respondWithError(w, err)
		return
	}
	respondJSON(w, http.StatusCreated, dto.FromShare(share))
}

func (h *SharesHandler) handleList(w http.ResponseWriter, r *http.Request) {
	shares, err := h.list.Execute(currentUser(r))
	if err != nil {
		respondWithError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, dto.FromShares(shares))
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
	respondJSON(w, http.StatusOK, dto.MessageResponse{Message: "revoked"})
}
