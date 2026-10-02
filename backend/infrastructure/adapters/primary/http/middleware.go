package xhttp

import (
	"errors"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/EslamYasser-Dev/simple-file-share/application/services"
	domainerrors "github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/primary/authctx"
)

// SessionCookieName carries the HttpOnly access token for browser clients.
// XSS cannot read it; API clients keep using Authorization: Bearer.
const SessionCookieName = "fs_session"

// AuthMiddleware enforces API keys, Bearer JWT, the HttpOnly session cookie,
// or Basic Auth and stores the authenticated account in the request context.
// When a request authenticates with an API key, its scope is enforced here:
// safe methods need read, mutating methods need write, /api/admin/* needs
// admin, and the credential surface (/api/auth/* except /api/auth/me) is
// never reachable by keys.
func AuthMiddleware(auth *services.AuthenticateService, tokens *services.TokenService, apiKeys *services.APIKeyService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, scope, err := resolveAuthenticatedUser(r, auth, tokens, apiKeys)
			if err != nil {
				if tokens != nil {
					w.Header().Set("WWW-Authenticate", `Bearer realm="api", Basic realm="api"`)
				} else {
					w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
				}
				if errors.Is(err, domainerrors.ErrTwoFactorRequired) {
					// JSON challenge so API clients can prompt for a code;
					// Basic auth has no OTP channel (use the token endpoint).
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusUnauthorized)
					_, _ = w.Write([]byte(`{"error":"totp_required"}`))
					return
				}
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			if scope != "" && !apiKeyRequestAllowed(scope, r) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				if strings.HasPrefix(r.URL.Path, "/api/auth/") {
					_, _ = w.Write([]byte(`{"error":"api keys cannot manage credentials"}`))
					return
				}
				_, _ = w.Write([]byte(`{"error":"insufficient scope"}`))
				return
			}
			next.ServeHTTP(w, r.WithContext(authctx.WithUser(r.Context(), user)))
		})
	}
}

// apiKeyRequestAllowed reports whether an API key with the given scope may
// perform this request. Scope ranks read < write < admin; /api/admin/*
// always demands admin, everything else demands read for safe methods and
// write for mutations. Keys may verify themselves at /api/auth/me but never
// reach password, session, TOTP, or key-management endpoints — that surface
// requires a real login so a stolen key cannot escalate to account control.
func apiKeyRequestAllowed(scope string, r *http.Request) bool {
	path := r.URL.Path
	if strings.HasPrefix(path, "/api/auth/") && path != "/api/auth/me" {
		return false
	}
	required := models.APIKeyScopeRead
	switch r.Method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
	default:
		required = models.APIKeyScopeWrite
	}
	if strings.HasPrefix(path, "/api/admin/") {
		required = models.APIKeyScopeAdmin
	}
	return models.APIKeyScopeAllows(scope, required)
}

// resolveAuthenticatedUser returns the caller and, when the caller presented
// an API key, that key's scope ("" for JWT/cookie/Basic = unrestricted).
func resolveAuthenticatedUser(r *http.Request, auth *services.AuthenticateService, tokens *services.TokenService, apiKeys *services.APIKeyService) (*models.User, string, error) {
	if authz := r.Header.Get("Authorization"); len(authz) > 7 && strings.EqualFold(authz[:7], "Bearer ") {
		raw := strings.TrimSpace(authz[7:])
		if strings.HasPrefix(raw, services.APIKeyPrefix) {
			if apiKeys == nil {
				return nil, "", ports.ErrUnauthorized
			}
			return apiKeys.Authenticate(raw)
		}
		if tokens == nil {
			return nil, "", ports.ErrUnauthorized
		}
		user, err := tokens.Authenticate(raw)
		return user, "", err
	}

	// Browser session: HttpOnly cookie (preferred over any JS-readable store).
	if tokens != nil {
		if c, err := r.Cookie(SessionCookieName); err == nil && c.Value != "" {
			user, err := tokens.Authenticate(c.Value)
			if err == nil {
				return user, "", nil
			}
			// Fall through: expired/revoked cookie may still allow Basic.
		}
	}

	username, password, ok := r.BasicAuth()
	if !ok {
		return nil, "", ports.ErrUnauthorized
	}
	user, err := auth.Execute(username, password)
	if err != nil {
		return nil, "", err
	}
	// Basic has no OTP channel: enrolled accounts must use the token
	// endpoint (or an API key) instead.
	if tokens != nil {
		if err := tokens.CheckTwoFactor(user, ""); err != nil {
			return nil, "", err
		}
	}
	return user, "", nil
}

// ResolveAuthenticatedUser resolves the caller from Bearer token, session
// cookie, or Basic credentials. It is exported for handlers that authenticate
// individually (e.g. GET /metrics) instead of sitting behind the auth chain.
// API keys are deliberately not accepted here: metrics stays on JWT/Basic.
func ResolveAuthenticatedUser(
	r *http.Request,
	auth *services.AuthenticateService,
	tokens *services.TokenService,
) (*models.User, error) {
	user, _, err := resolveAuthenticatedUser(r, auth, tokens, nil)
	return user, err
}

// corsMiddleware reflects a single allowed Origin with credentials enabled.
// It never sends ACAO:* (which is incompatible with credentialed requests).
// Allowed origins: exact CORS_ORIGINS entries plus the request host itself.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && originAllowed(r, origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Add("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Upload-Offset")
			w.Header().Set("Access-Control-Max-Age", "600")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func originAllowed(r *http.Request, origin string) bool {
	u, err := url.Parse(origin)
	if err != nil || u.Host == "" {
		return false
	}
	if strings.EqualFold(u.Host, r.Host) {
		return true
	}
	raw := strings.TrimSpace(os.Getenv("CORS_ORIGINS"))
	if raw == "" {
		return false
	}
	for _, allowed := range strings.Split(raw, ",") {
		allowed = strings.TrimRight(strings.TrimSpace(allowed), "/")
		if allowed != "" && strings.EqualFold(allowed, strings.TrimRight(origin, "/")) {
			return true
		}
	}
	return false
}

func loggingMiddleware(next http.Handler, logger ports.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)
		logger.Info("request",
			"request_id", RequestIDFromContext(r.Context()),
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.status,
			"bytes", rw.bytesWritten,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}

func chainMiddleware(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

// SetSessionCookie attaches the access token as an HttpOnly cookie browsers
// send automatically (XSS cannot exfiltrate it).
func SetSessionCookie(w http.ResponseWriter, r *http.Request, token string, maxAgeSeconds int) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   maxAgeSeconds,
		HttpOnly: true,
		Secure:   requestIsHTTPS(r),
		SameSite: http.SameSiteLaxMode,
	})
}

// ClearSessionCookie removes the browser session cookie.
func ClearSessionCookie(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   requestIsHTTPS(r),
		SameSite: http.SameSiteLaxMode,
	})
}

func requestIsHTTPS(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	return strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")
}
