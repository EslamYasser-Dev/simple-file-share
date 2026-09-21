package ports

import "github.com/EslamYasser-Dev/simple-file-share/domain/models"

// AuthProvider authenticates credentials and returns the account, or
// domainerrors.ErrInvalidCredentials (or domainerrors.ErrUserNotFound) when
// the credentials are rejected.
type AuthProvider interface {
	Authenticate(username, password string) (*models.User, error)
}
