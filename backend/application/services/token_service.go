package services

import (
	"time"

	domainerrors "github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

// TokenPair is a freshly issued access token and its expiry.
type TokenPair struct {
	AccessToken string
	ExpiresAt   time.Time
}

// TokenService issues access tokens for authenticated accounts and resolves
// bearer tokens back to accounts, including revocation checks.
type TokenService struct {
	auth   *AuthenticateService
	tokens ports.TokenManager
	users  ports.UserRepository
}

func NewTokenService(auth *AuthenticateService, tokens ports.TokenManager, users ports.UserRepository) *TokenService {
	return &TokenService{auth: auth, tokens: tokens, users: users}
}

func (s *TokenService) Login(username, password string) (*TokenPair, error) {
	user, err := s.auth.Execute(username, password)
	if err != nil {
		return nil, err
	}
	return s.IssueFor(user)
}

func (s *TokenService) IssueFor(user *models.User) (*TokenPair, error) {
	token, claims, err := s.tokens.Issue(user.Username)
	if err != nil {
		return nil, err
	}
	return &TokenPair{AccessToken: token, ExpiresAt: claims.ExpiresAt}, nil
}

func (s *TokenService) Authenticate(token string) (*models.User, error) {
	claims, err := s.tokens.Verify(token)
	if err != nil {
		return nil, domainerrors.ErrInvalidCredentials
	}
	user, err := s.users.FindByUsername(claims.Subject)
	if err != nil {
		return nil, domainerrors.ErrInvalidCredentials
	}
	return user, nil
}

func (s *TokenService) Refresh(token string) (*TokenPair, error) {
	claims, err := s.tokens.Verify(token)
	if err != nil {
		return nil, domainerrors.ErrInvalidCredentials
	}
	user, err := s.users.FindByUsername(claims.Subject)
	if err != nil {
		return nil, domainerrors.ErrInvalidCredentials
	}
	s.tokens.Revoke(claims.ID, claims.ExpiresAt)
	return s.IssueFor(user)
}

func (s *TokenService) Revoke(token string) error {
	claims, err := s.tokens.Verify(token)
	if err != nil {
		return domainerrors.ErrInvalidCredentials
	}
	s.tokens.Revoke(claims.ID, claims.ExpiresAt)
	return nil
}
