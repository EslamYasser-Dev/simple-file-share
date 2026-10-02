package services

import (
	"time"

	domainerrors "github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

// SessionService tracks issued access tokens so users can list and revoke
// their active sessions. Recording is a convenience view — actual token
// rejection always goes through TokenManager.Revoke, so a lost sessions.json
// can never resurrect a revoked token.
type SessionService struct {
	store  ports.SessionRegistry
	tokens ports.TokenManager
}

// NewSessionService wires the registry and the token manager (both may be
// nil, which disables the feature).
func NewSessionService(store ports.SessionRegistry, tokens ports.TokenManager) *SessionService {
	return &SessionService{store: store, tokens: tokens}
}

// Enabled reports whether session tracking is available.
func (s *SessionService) Enabled() bool {
	return s != nil && s.store != nil
}

// Observe records or refreshes the session for an issued pair. The HTTP
// handlers pass the client's remote address and user agent; the base minting
// path (gRPC, OAuth redirect) passes empty strings. Never fails the caller.
func (s *SessionService) Observe(pair *TokenPair, remote, userAgent string) {
	if !s.Enabled() || pair == nil || pair.ID == "" || pair.Subject == "" {
		return
	}
	issuedAt := time.Now().UTC()
	// Enrichment re-record keeps the original issue time when the session is
	// already known; only brand-new jtis stamp "now".
	if existing, ok, err := s.store.Find(pair.ID); err == nil && ok && !existing.IssuedAt.IsZero() {
		issuedAt = existing.IssuedAt
	}
	_ = s.store.Record(ports.Session{
		JTI:       pair.ID,
		Subject:   pair.Subject,
		IssuedAt:  issuedAt,
		ExpiresAt: pair.ExpiresAt,
		Remote:    remote,
		UserAgent: userAgent,
	})
}

// List returns the actor's sessions newest first. A nil actor (auth
// disabled) or a system view returns every tracked session.
func (s *SessionService) List(actor *models.User) ([]ports.Session, error) {
	if !s.Enabled() {
		return []ports.Session{}, nil
	}
	subject := ""
	if actor != nil && !actor.IsSystemView() {
		subject = actor.Username
	}
	sessions, err := s.store.List(subject)
	if err != nil {
		return nil, err
	}
	if sessions == nil {
		sessions = []ports.Session{}
	}
	return sessions, nil
}

// CurrentJTI returns the JTI of the presented bearer/cookie token ("" when
// the token is missing or invalid). Handlers use it to flag the active
// session in list responses.
func (s *SessionService) CurrentJTI(token string) string {
	if s == nil || s.tokens == nil || token == "" {
		return ""
	}
	claims, err := s.tokens.Verify(token)
	if err != nil {
		return ""
	}
	return claims.ID
}

// Revoke removes one session. Users may only revoke their own; a system view
// (auth disabled) may revoke any.
func (s *SessionService) Revoke(actor *models.User, jti string) error {
	if !s.Enabled() {
		return &domainerrors.NotFoundError{Path: "session"}
	}
	sess, ok, err := s.store.Find(jti)
	if err != nil {
		return err
	}
	if !ok {
		return &domainerrors.NotFoundError{Path: "session"}
	}
	if actor != nil && !actor.IsSystemView() && sess.Subject != actor.Username {
		// Do not reveal whether someone else's session id exists.
		return &domainerrors.ForbiddenError{Action: "revoke session"}
	}
	if s.tokens != nil {
		s.tokens.Revoke(sess.JTI, sess.ExpiresAt)
	}
	return s.store.Remove(sess.JTI)
}
