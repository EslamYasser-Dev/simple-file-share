package services

import (
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

// AuthenticateService is the application-layer account login use case. Every
// primary adapter (HTTP Basic middleware, gRPC interceptor and AuthService)
// funnels through it so credential handling lives in one place.
type AuthenticateService struct {
	provider ports.AuthProvider
}

func NewAuthenticateService(provider ports.AuthProvider) *AuthenticateService {
	return &AuthenticateService{provider: provider}
}

// Execute resolves credentials to an account, returning a domain error (e.g.
// domainerrors.ErrInvalidCredentials) when they do not match.
func (s *AuthenticateService) Execute(username, password string) (*models.User, error) {
	return s.provider.Authenticate(username, password)
}
