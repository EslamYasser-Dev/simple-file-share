package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/EslamYasser-Dev/simple-file-share/application/services"
	xhttp "github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/primary/http"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/primary/http/dto"
)

// GoogleExchangeHandler trades a Google OIDC ID token for an app session:
// POST /api/auth/google/exchange with {"idToken": "..."}. First-time users
// get an account through the shared OAuth upsert, so one button covers both
// sign-in and sign-up. Public, like the password token endpoint.
type GoogleExchangeHandler struct {
	svc      *services.GoogleLoginService
	audit    *services.AuditService
	sessions *services.SessionService
}

func NewGoogleExchangeHandler(svc *services.GoogleLoginService) *GoogleExchangeHandler {
	return &GoogleExchangeHandler{svc: svc}
}

// SetAudit attaches the security audit trail (nil disables recording).
func (h *GoogleExchangeHandler) SetAudit(a *services.AuditService) { h.audit = a }

// SetSessions attaches the session registry for client enrichment.
func (h *GoogleExchangeHandler) SetSessions(s *services.SessionService) { h.sessions = s }

func (h *GoogleExchangeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req dto.GoogleExchangeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.IDToken == "" {
		respondError(w, http.StatusBadRequest, "id token is required")
		return
	}

	result, err := h.svc.Execute(r.Context(), req.IDToken)
	if err != nil {
		// Generic like password login: the response must not distinguish
		// bad tokens from unknown Google identities.
		h.audit.Record(services.AuditLoginFail, "google", clientIP(r), "", "invalid id token")
		respondWithError(w, err)
		return
	}
	detail := "google"
	if result.IsNew {
		detail = "google (new account)"
	}
	h.audit.Record(services.AuditLoginOK, result.User.Username, clientIP(r), "", detail)
	if h.sessions != nil {
		h.sessions.Observe(result.Pair, clientIP(r), r.UserAgent())
	}
	maxAge := int(time.Until(result.Pair.ExpiresAt).Seconds())
	if maxAge < 0 {
		maxAge = 0
	}
	xhttp.SetSessionCookie(w, r, result.Pair.AccessToken, maxAge)
	respondJSON(w, http.StatusOK, dto.GoogleExchangeResponse{
		AccessToken:  result.Pair.AccessToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(maxAge),
		IsNewAccount: result.IsNew,
	})
}
