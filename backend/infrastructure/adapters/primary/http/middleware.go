package xhttp

import (
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/EslamYasser-Dev/simple-file-share/application/services"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/primary/authctx"
)

// SessionCookieName carries the HttpOnly access token for browser clients.
// XSS cannot read it; API clients keep using Authorization: Bearer.
const SessionCookieName = "fs_session"

// AuthMiddleware enforces Bearer JWT, the HttpOnly session cookie, or Basic
// Auth and stores the authenticated account in the request context.
func AuthMiddleware(auth *services.AuthenticateService, tokens *services.TokenService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, err := resolveAuthenticatedUser(r, auth, tokens)
			if err != nil {
				if tokens != nil {
					w.Header().Set("WWW-Authenticate", `Bearer realm="api", Basic realm="api"`)
				} else {
					w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
				}
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r.WithContext(authctx.WithUser(r.Context(), user)))
		})
	}
}

func resolveAuthenticatedUser(r *http.Request, auth *services.AuthenticateService, tokens *services.TokenService) (*models.User, error) {
	if authz := r.Header.Get("Authorization"); len(authz) > 7 && strings.EqualFold(authz[:7], "Bearer ") {
		if tokens == nil {
			return nil, ports.ErrUnauthorized
		}
		return tokens.Authenticate(strings.TrimSpace(authz[7:]))
	}

	// Browser session: HttpOnly cookie (preferred over any JS-readable store).
	if tokens != nil {
		if c, err := r.Cookie(SessionCookieName); err == nil && c.Value != "" {
			user, err := tokens.Authenticate(c.Value)
			if err == nil {
				return user, nil
			}
			// Fall through: expired/revoked cookie may still allow Basic.
		}
	}

	username, password, ok := r.BasicAuth()
	if !ok {
		return nil, ports.ErrUnauthorized
	}
	return auth.Execute(username, password)
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
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
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
		logger.Info("request", "method", r.Method, "path", r.URL.Path)
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)
		logger.Info("response", "method", r.Method, "path", r.URL.Path, "status", rw.status)
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
