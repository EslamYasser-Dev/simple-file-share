package xhttp

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestChainMiddlewareOrder sanity-checks that chainMiddleware applies the first
// middleware outermost and reaches the handler last. A future refactor must not
// silently drop the rate limiter or AuthMiddleware from the outer layers.
func TestChainMiddlewareOrder(t *testing.T) {
	var order []string
	appenders := []func(http.Handler) http.Handler{
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, "outer")
				next.ServeHTTP(w, r)
			})
		},
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, "middle")
				next.ServeHTTP(w, r)
			})
		},
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, "inner")
				next.ServeHTTP(w, r)
			})
		},
	}
	handler := chainMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
	}), appenders...)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if got := strings.Join(order, ","); got != "outer,middle,inner,handler" {
		t.Fatalf("middleware order = %q, want outer,middle,inner,handler", got)
	}
}
