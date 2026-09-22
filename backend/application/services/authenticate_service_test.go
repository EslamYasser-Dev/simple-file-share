package services

import (
	"errors"
	"testing"

	domainerrors "github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

type stubAuthenticateProvider struct {
	user *models.User
	err  error
}

func (s stubAuthenticateProvider) Authenticate(_, _ string) (*models.User, error) {
	return s.user, s.err
}

var _ ports.AuthProvider = stubAuthenticateProvider{}

func TestAuthenticateServiceResolvesValidCredentials(t *testing.T) {
	alice := &models.User{Username: "alice", IsAdmin: true}
	service := NewAuthenticateService(stubAuthenticateProvider{user: alice})

	user, err := service.Execute("alice", "secret")
	if err != nil {
		t.Fatalf("authenticate = %v", err)
	}
	if user != alice {
		t.Fatalf("user = %+v, want %+v", user, alice)
	}
}

func TestAuthenticateServicePropagatesInvalidCredentials(t *testing.T) {
	service := NewAuthenticateService(stubAuthenticateProvider{err: domainerrors.ErrInvalidCredentials})

	if _, err := service.Execute("alice", "wrong"); !errors.Is(err, domainerrors.ErrInvalidCredentials) {
		t.Fatalf("authenticate = %v, want ErrInvalidCredentials", err)
	}
}
