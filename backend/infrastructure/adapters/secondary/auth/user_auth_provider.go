package auth

import (
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
		// Collapse "user missing" into the same generic failure as a bad
		// password so the response never reveals whether an account exists.
		return nil, domainerrors.ErrInvalidCredentials
	}
	if !user.Enabled {
		return nil, domainerrors.ErrInvalidCredentials
	}
	if !p.hasher.Verify(password, user.PasswordHash) {
		return nil, domainerrors.ErrInvalidCredentials
	}
	return user, nil
}
