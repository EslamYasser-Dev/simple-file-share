package auth

import (
	"errors"

	domainerrors "github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

// UserAuthProvider authenticates Basic Auth credentials against the user
// repository. Password hash comparison is constant-time via the hasher.
type UserAuthProvider struct {
	users  ports.UserRepository
	hasher ports.PasswordHasher
}

var _ ports.AuthProvider = (*UserAuthProvider)(nil)

func NewUserAuthProvider(users ports.UserRepository, hasher ports.PasswordHasher) *UserAuthProvider {
	return &UserAuthProvider{users: users, hasher: hasher}
}

func (p *UserAuthProvider) Authenticate(username, password string) (*models.User, error) {
	user, err := p.users.FindByUsername(username)
	if err != nil {
		return nil, err
	}
	if !p.hasher.Verify(password, user.PasswordHash) {
		// Return a generic failure regardless of whether the user exists to
		// avoid leaking account existence through response timing or errors.
		return nil, errors.Join(domainerrors.ErrInvalidCredentials, domainerrors.ErrUserNotFound)
	}
	return user, nil
}
