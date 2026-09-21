// Package authctx carries the authenticated user through request contexts. It
// is shared by every primary adapter (HTTP, gRPC) so they agree on the key.
package authctx

import (
	"context"

	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
)

type contextKey string

const userContextKey contextKey = "user"

// WithUser returns a copy of ctx carrying the authenticated user.
func WithUser(ctx context.Context, user *models.User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

// UserFromContext returns the authenticated user, or nil when requests are
// served without auth (auth disabled, or a middleware that does not attach a
// user). A nil user is treated as a system/admin view by the scoping layer.
func UserFromContext(ctx context.Context) *models.User {
	u, _ := ctx.Value(userContextKey).(*models.User)
	return u
}
