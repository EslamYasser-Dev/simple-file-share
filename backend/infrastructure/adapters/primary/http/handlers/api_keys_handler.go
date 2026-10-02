package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/EslamYasser-Dev/simple-file-share/application/services"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
)

// APIKeysHandler serves the self-service credential endpoints:
//
//	GET    /api/auth/api-keys        -> {"items":[{...}]}
//	POST   /api/auth/api-keys        -> mint (plaintext returned once, 201)
//	DELETE /api/auth/api-keys/{id}   -> revoke one of your keys
type APIKeysHandler struct {
	svc   *services.APIKeyService
	audit *services.AuditService
}

// NewAPIKeysHandler constructs the endpoints (nil service disables them).
func NewAPIKeysHandler(svc *services.APIKeyService) *APIKeysHandler {
	return &APIKeysHandler{svc: svc}
}

// SetAudit injects the audit recorder (nil-safe).
func (h *APIKeysHandler) SetAudit(audit *services.AuditService) {
	h.audit = audit
}

type apiKeyItem struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Scope      string    `json:"scope"`
	KeyPrefix  string    `json:"keyPrefix"`
	CreatedAt  time.Time `json:"createdAt"`
	ExpiresAt  time.Time `json:"expiresAt,omitempty"`
	LastUsedAt time.Time `json:"lastUsedAt,omitempty"`
}

func (h *APIKeysHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleList(w, r)
	case http.MethodPost:
		h.handleCreate(w, r)
	case http.MethodDelete:
		h.handleRevoke(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *APIKeysHandler) handleList(w http.ResponseWriter, r *http.Request) {
	actor := currentUser(r)
	if actor == nil {
		respondError(w, http.StatusForbidden, "authentication required")
		return
	}
	keys, err := h.svc.List(actor)
	if err != nil {
		respondWithError(w, err)
		return
	}
	items := make([]apiKeyItem, 0, len(keys))
	for _, k := range keys {
		items = append(items, apiKeyItemFor(k))
	}
	respondJSON(w, http.StatusOK, map[string]any{"items": items})
}

type createAPIKeyRequest struct {
	Name      string `json:"name"`
	Scope     string `json:"scope"`
	ExpiresIn int64  `json:"expiresIn"`
}

func (h *APIKeysHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	actor := currentUser(r)
	if actor == nil {
		respondError(w, http.StatusForbidden, "authentication required")
		return
	}
	// An empty body is allowed: scope then defaults to read and the missing
	// name fails validation below.
	var req createAPIKeyRequest
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	plaintext, key, err := h.svc.Issue(
		actor, req.Name, req.Scope, time.Duration(req.ExpiresIn)*time.Second,
	)
	if err != nil {
		respondWithError(w, err)
		return
	}
	h.audit.Record(
		services.AuditAPIKeyCreate, actorName(r), clientIP(r),
		"apikey:"+key.ID, "scope="+key.Scope,
	)
	respondJSON(w, http.StatusCreated, map[string]any{
		"key":    plaintext, // shown exactly once — only the hash is stored
		"id":     key.ID,
		"name":   key.Name,
		"scope":  key.Scope,
		"prefix": keyPrefixOf(key),
	})
}

func (h *APIKeysHandler) handleRevoke(w http.ResponseWriter, r *http.Request) {
	actor := currentUser(r)
	if actor == nil {
		respondError(w, http.StatusForbidden, "authentication required")
		return
	}
	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		respondError(w, http.StatusBadRequest, "key id is required")
		return
	}
	if err := h.svc.Revoke(actor, id); err != nil {
		respondWithError(w, err)
		return
	}
	h.audit.Record(services.AuditAPIKeyRevoke, actorName(r), clientIP(r), "apikey:"+id, "")
	respondJSON(w, http.StatusOK, map[string]string{"status": "revoked"})
}

func apiKeyItemFor(k models.APIKey) apiKeyItem {
	return apiKeyItem{
		ID:         k.ID,
		Name:       k.Name,
		Scope:      k.Scope,
		KeyPrefix:  keyPrefixOf(k),
		CreatedAt:  k.CreatedAt,
		ExpiresAt:  k.ExpiresAt,
		LastUsedAt: k.LastUsedAt,
	}
}

// keyPrefixOf renders "sfs_<id>_…" for display. The id already identifies
// the key, so no secret characters need to be retained.
func keyPrefixOf(k models.APIKey) string {
	return services.APIKeyPrefix + k.ID + "_…"
}
