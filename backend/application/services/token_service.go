package services

import (
	"time"

	"github.com/EslamYasser-Dev/simple-file-share/application/events"
	domainerrors "github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

// TokenPair is a freshly issued access token and its expiry.
type TokenPair struct {
	AccessToken string
	ExpiresAt   time.Time
	// ID is the token's jti for session bookkeeping (not serialized to
	// API clients — see dto.FromTokenPair).
	ID string
	// Subject is the account the token was issued for.
	Subject string
}

// TokenService issues access tokens for authenticated accounts and resolves
// bearer tokens back to accounts, including revocation checks.
type TokenService struct {
	auth      *AuthenticateService
	tokens    ports.TokenManager
	users     ports.UserRepository
	bus       *events.Bus
	sessions  *SessionService
	twoFactor *TOTPService
}

func NewTokenService(auth *AuthenticateService, tokens ports.TokenManager, users ports.UserRepository) *TokenService {
	return &TokenService{auth: auth, tokens: tokens, users: users}
}

// SetEventBus attaches a live-update bus (nil disables publishing).
func (s *TokenService) SetEventBus(bus *events.Bus) { s.bus = bus }

// SetSessions attaches the session registry so every minted token shows up
// in the account's session list (nil disables tracking).
func (s *TokenService) SetSessions(sess *SessionService) { s.sessions = sess }

// SetTwoFactor attaches the TOTP service so password logins enforce the
// second factor for enrolled accounts (nil skips enforcement — used by
// tests that do not model 2FA).
func (s *TokenService) SetTwoFactor(tf *TOTPService) { s.twoFactor = tf }

func (s *TokenService) Login(username, password string) (*TokenPair, error) {
	return s.LoginWithOTP(username, password, "")
}

// LoginWithOTP authenticates the password and, for enrolled accounts,
// requires otp (a 6-digit TOTP code or an unused backup code). Missing otp
// returns domainerrors.ErrTwoFactorRequired; a wrong one returns the generic
// invalid-credentials error so responses never reveal which factor failed.
func (s *TokenService) LoginWithOTP(username, password, otp string) (*TokenPair, error) {
	user, err := s.auth.Execute(username, password)
	if err != nil {
		return nil, err
	}
	if err := s.CheckTwoFactor(user, otp); err != nil {
		return nil, err
	}
	pair, err := s.IssueFor(user)
	if err != nil {
		return nil, err
	}
	publishEvent(s.bus, events.TypeLogin, "", user)
	return pair, nil
}

// CheckTwoFactor enforces the second factor when the account is enrolled.
// An empty otp yields ErrTwoFactorRequired — the challenge response used by
// Basic auth (which has no OTP channel and must fall back to the token
// endpoint).
func (s *TokenService) CheckTwoFactor(user *models.User, otp string) error {
	if user == nil || !user.TOTPEnabled || s.twoFactor == nil {
		return nil
	}
	return s.twoFactor.VerifyLogin(user, otp)
}

func (s *TokenService) IssueFor(user *models.User) (*TokenPair, error) {
	token, claims, err := s.tokens.Issue(user.Username)
	if err != nil {
		return nil, err
	}
	pair := &TokenPair{
		AccessToken: token,
		ExpiresAt:   claims.ExpiresAt,
		ID:          claims.ID,
		Subject:     user.Username,
	}
	// Base record for every minted token (HTTP handlers enrich it with the
	// client's remote address and user agent afterwards).
	s.sessions.Observe(pair, "", "")
	return pair, nil
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
	if !user.Enabled {
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
	if !user.Enabled {
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
