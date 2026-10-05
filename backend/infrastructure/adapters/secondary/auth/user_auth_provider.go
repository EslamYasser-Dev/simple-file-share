package auth

import (
	"sync"

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

// burnHash is a decoy hash used to equalize timing between "no such user"
// and "wrong password": without it, missing users skip the KDF entirely and
// response latency reveals which usernames exist. Built lazily so startup
// does not pay the derivation.
var (
	burnHashOnce sync.Once
	burnHash     string
)

func (p *UserAuthProvider) Authenticate(username, password string) (*models.User, error) {
	user, err := p.users.FindByUsername(username)
	if err != nil {
		// Collapse "user missing" into the same generic failure as a bad
		// password so the response never reveals whether an account exists.
		p.burnVerification(password)
		return nil, domainerrors.ErrInvalidCredentials
	}
	if !user.Enabled {
		p.burnVerification(password)
		return nil, domainerrors.ErrInvalidCredentials
	}
	if !p.hasher.Verify(password, user.PasswordHash) {
		return nil, domainerrors.ErrInvalidCredentials
	}
	return user, nil
}

// burnVerification spends the same KDF work a real password check would, so
// unknown and disabled accounts answer on the same clock as a bad password.
// The result is deliberately ignored.
func (p *UserAuthProvider) burnVerification(password string) {
	burnHashOnce.Do(func() {
		if h, err := p.hasher.Hash(""); err == nil {
			burnHash = h
		}
	})
	if burnHash != "" {
		_ = p.hasher.Verify(password, burnHash)
	}
}
