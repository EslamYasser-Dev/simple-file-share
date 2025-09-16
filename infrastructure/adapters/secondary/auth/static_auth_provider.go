package auth

import "github.com/EslamYasser-Dev/simple-file-share/domain/ports"

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

// Authenticate validates username and password.
func (p *StaticAuthProvider) Authenticate(username, password string) bool {
	return username == p.username && password == p.password
}

var _ ports.AuthProvider = (*StaticAuthProvider)(nil)
