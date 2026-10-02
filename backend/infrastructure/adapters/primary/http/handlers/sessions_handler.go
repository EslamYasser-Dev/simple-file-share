package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/EslamYasser-Dev/simple-file-share/application/services"
)

// SessionsHandler serves the session list for the current account:
//
//	GET    /api/auth/sessions          -> {"items":[{...,"current":true}]}
//	DELETE /api/auth/sessions/{jti}    -> revoke one of your sessions
type SessionsHandler struct {
	sessions *services.SessionService
}

// NewSessionsHandler constructs the session endpoints (nil disables them).
func NewSessionsHandler(sessions *services.SessionService) *SessionsHandler {
	return &SessionsHandler{sessions: sessions}
}

type sessionItem struct {
	JTI       string    `json:"jti"`
	IssuedAt  time.Time `json:"issuedAt"`
	ExpiresAt time.Time `json:"expiresAt"`
	Remote    string    `json:"remote,omitempty"`
	UserAgent string    `json:"userAgent,omitempty"`
	Current   bool      `json:"current"`
}

func (h *SessionsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleList(w, r)
	case http.MethodDelete:
		h.handleRevoke(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *SessionsHandler) handleList(w http.ResponseWriter, r *http.Request) {
	sessions, err := h.sessions.List(currentUser(r))
	if err != nil {
		respondWithError(w, err)
		return
	}
	current := h.sessions.CurrentJTI(sessionTokenFromRequest(r))
	items := make([]sessionItem, 0, len(sessions))
	for _, s := range sessions {
		items = append(items, sessionItem{
			JTI:       s.JTI,
			IssuedAt:  s.IssuedAt,
			ExpiresAt: s.ExpiresAt,
			Remote:    s.Remote,
			UserAgent: s.UserAgent,
			Current:   s.JTI == current && current != "",
		})
	}
	respondJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *SessionsHandler) handleRevoke(w http.ResponseWriter, r *http.Request) {
	jti := strings.TrimSpace(r.PathValue("jti"))
	if jti == "" {
		respondError(w, http.StatusBadRequest, "session id is required")
		return
	}
	if err := h.sessions.Revoke(currentUser(r), jti); err != nil {
		respondWithError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "revoked"})
}
