package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/EslamYasser-Dev/simple-file-share/application/services"
	xhttp "github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/primary/http"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/primary/http/dto"
	secondaryauth "github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/auth"
)

const oauthStateCookie = "fs_oauth_state"

// TokenHandler issues access tokens for password credentials (public endpoint).
// The token is returned for API clients and also set as an HttpOnly cookie
// for browsers (credentials: 'include').
type TokenHandler struct {
	tokens *services.TokenService
}

func NewTokenHandler(tokens *services.TokenService) *TokenHandler {
	return &TokenHandler{tokens: tokens}
}

func (h *TokenHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req dto.TokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	pair, err := h.tokens.Login(req.Username, req.Password)
	if err != nil {
		respondWithError(w, err)
		return
	}
	maxAge := int(time.Until(pair.ExpiresAt).Seconds())
	if maxAge < 0 {
		maxAge = 0
	}
	xhttp.SetSessionCookie(w, r, pair.AccessToken, maxAge)
	respondJSON(w, http.StatusOK, dto.FromTokenPair(pair))
}

// RefreshHandler rotates a valid access token into a new one.
type RefreshHandler struct {
	tokens *services.TokenService
}

func NewRefreshHandler(tokens *services.TokenService) *RefreshHandler {
	return &RefreshHandler{tokens: tokens}
}

func (h *RefreshHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	token := sessionTokenFromRequest(r)
	if token == "" {
		respondError(w, http.StatusBadRequest, "access token is required")
		return
	}

	pair, err := h.tokens.Refresh(token)
	if err != nil {
		xhttp.ClearSessionCookie(w, r)
		respondWithError(w, err)
		return
	}
	maxAge := int(time.Until(pair.ExpiresAt).Seconds())
	if maxAge < 0 {
		maxAge = 0
	}
	xhttp.SetSessionCookie(w, r, pair.AccessToken, maxAge)
	respondJSON(w, http.StatusOK, dto.FromTokenPair(pair))
}

// RevokeHandler deny-lists a currently valid access token and clears the
// browser session cookie.
type RevokeHandler struct {
	tokens *services.TokenService
}

func NewRevokeHandler(tokens *services.TokenService) *RevokeHandler {
	return &RevokeHandler{tokens: tokens}
}

func (h *RevokeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	token := sessionTokenFromRequest(r)
	// Clear the cookie regardless so the browser cannot keep a dead session.
	xhttp.ClearSessionCookie(w, r)
	if token == "" {
		// Nothing to revoke server-side; cookie already cleared.
		respondJSON(w, http.StatusOK, map[string]string{"status": "revoked"})
		return
	}

	if err := h.tokens.Revoke(token); err != nil {
		respondWithError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "revoked"})
}

// sessionTokenFromRequest prefers Authorization: Bearer, then a JSON body
// token, then the HttpOnly session cookie (browser logout/refresh).
func sessionTokenFromRequest(r *http.Request) string {
	if authz := r.Header.Get("Authorization"); len(authz) > 7 && strings.EqualFold(authz[:7], "Bearer ") {
		return strings.TrimSpace(authz[7:])
	}
	if r.Body != nil {
		var body dto.RefreshRequest
		dec := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
		if err := dec.Decode(&body); err == nil {
			if t := strings.TrimSpace(body.AccessToken); t != "" {
				return t
			}
		}
	}
	if c, err := r.Cookie(xhttp.SessionCookieName); err == nil && c.Value != "" {
		return c.Value
	}
	return ""
}

// OAuthStartHandler redirects the browser to the provider's consent page and
// binds the CSRF state to an HttpOnly cookie on this browser.
type OAuthStartHandler struct {
	oauth *services.OAuthLoginService
}

func NewOAuthStartHandler(oauth *services.OAuthLoginService) *OAuthStartHandler {
	return &OAuthStartHandler{oauth: oauth}
}

func (h *OAuthStartHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	provider := r.PathValue("provider")
	base, err := oauthPublicBase(r)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "oauth is not configured for this origin")
		return
	}
	callbackURI := base + "/api/auth/oauth/" + provider + "/callback"
	authURL, state, err := h.oauth.Begin(provider, callbackURI)
	if err != nil {
		respondWithError(w, err)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookie,
		Value:    state,
		Path:     "/api/auth/oauth",
		MaxAge:   600,
		HttpOnly: true,
		Secure:   strings.HasPrefix(base, "https://") || r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, authURL, http.StatusFound)
}

// OAuthCallbackHandler finishes the code exchange, sets the HttpOnly session
// cookie, and redirects to the SPA with no token in the URL.
type OAuthCallbackHandler struct {
	oauth *services.OAuthLoginService
}

func NewOAuthCallbackHandler(oauth *services.OAuthLoginService) *OAuthCallbackHandler {
	return &OAuthCallbackHandler{oauth: oauth}
}

func (h *OAuthCallbackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	provider := r.PathValue("provider")
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	browserState := ""
	if c, err := r.Cookie(oauthStateCookie); err == nil {
		browserState = c.Value
	}
	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookie,
		Value:    "",
		Path:     "/api/auth/oauth",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	if code == "" || state == "" {
		respondError(w, http.StatusBadRequest, "missing oauth code or state")
		return
	}

	base, err := oauthPublicBase(r)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "oauth is not configured for this origin")
		return
	}
	callbackURI := base + "/api/auth/oauth/" + provider + "/callback"

	spaBase, err := oauthSPABase(r, base)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "oauth is not configured for this origin")
		return
	}

	pair, err := h.oauth.Complete(provider, code, state, browserState, callbackURI)
	if err != nil {
		http.Redirect(w, r, spaBase+"/#oauthError=1", http.StatusFound)
		return
	}

	maxAge := int(time.Until(pair.ExpiresAt).Seconds())
	if maxAge < 0 {
		maxAge = 0
	}
	xhttp.SetSessionCookie(w, r, pair.AccessToken, maxAge)
	http.Redirect(w, r, spaBase+"/", http.StatusFound)
}

func oauthPublicBase(r *http.Request) (string, error) {
	base := strings.TrimRight(strings.TrimSpace(os.Getenv("OAUTH_REDIRECT_BASE")), "/")
	if base != "" {
		if err := secondaryauth.ValidateRedirectBase(base); err != nil {
			return "", err
		}
		return base, nil
	}
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if proto := r.Header.Get("X-Forwarded-Proto"); proto == "https" {
		scheme = "https"
	}
	return scheme + "://" + r.Host, nil
}

func oauthSPABase(r *http.Request, apiBase string) (string, error) {
	if spa := strings.TrimRight(strings.TrimSpace(os.Getenv("OAUTH_SPA_BASE")), "/"); spa != "" {
		if err := secondaryauth.ValidateRedirectBase(spa); err != nil {
			return "", err
		}
		return spa, nil
	}
	return apiBase, nil
}
