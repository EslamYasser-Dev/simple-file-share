package xhttp

import (
	"net/http"

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
