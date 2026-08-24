package auth

import (
	"crypto/subtle"

	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

// StaticAuthProvider implements simple static credential authentication.
type StaticAuthProvider struct {
	username string
	password string
}

// NewStaticAuthProvider creates a new auth provider with given credentials.
func NewStaticAuthProvider(username, password string) *StaticAuthProvider {
	return &StaticAuthProvider{
		username: username,
		password: password,
	}
}

// Authenticate validates username and password in constant time
// to prevent timing attacks on credential comparison.
func (p *StaticAuthProvider) Authenticate(username, password string) bool {
	userOK := subtle.ConstantTimeCompare([]byte(username), []byte(p.username)) == 1
	passOK := subtle.ConstantTimeCompare([]byte(password), []byte(p.password)) == 1
	return userOK && passOK
}

var _ ports.AuthProvider = (*StaticAuthProvider)(nil)
