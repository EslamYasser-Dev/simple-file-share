package xhttp

import (
	"context"
	"net/http"
	"strings"

	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

// AuthMiddleware returns an HTTP middleware that enforces Basic Auth.
// It uses the domain.AuthProvider port to validate credentials.
func AuthMiddleware(authProvider ports.AuthProvider) func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			user, pass, ok := r.BasicAuth()
			if !ok || !authProvider.Authenticate(user, pass) {
				w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			next(w, r)
		}
	}
}

// JWTMiddleware returns an HTTP middleware that enforces JWT authentication.
// It uses the domain.JWTProvider port to validate tokens.
func JWTMiddleware(jwtProvider ports.JWTProvider) func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Get the JWT token from the Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
				return
			}

			// The expected format is "Bearer <token>"
			const prefix = "Bearer "
			if !strings.HasPrefix(authHeader, prefix) {
				http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
				return
			}

			tokenString := authHeader[len(prefix):]

			// Validate the token
			claims, err := jwtProvider.ValidateToken(tokenString)
			if err != nil {
				http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
				return
			}

			// Add the claims to the request context
			ctx := context.WithValue(r.Context(), "user", claims)
			r = r.WithContext(ctx)

			next(w, r)
		}
	}
}