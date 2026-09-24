package services

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"strings"
	"sync"
	"time"

	domainerrors "github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/auth"
)

// OAuthLoginService completes the OAuth dance: validates the state cookie
// value, exchanges the code, upserts an account, and issues an access token.
type OAuthLoginService struct {
	providers         map[string]*auth.OAuthProvider
	users             ports.UserRepository
	fileRepo          ports.FileRepository
	scoper            ports.PathScoper
	tokens            *TokenService
	signupEnabled     bool
	defaultQuotaBytes int64
	states            *oauthStateStore
}

func NewOAuthLoginService(
	providers map[string]*auth.OAuthProvider,
	users ports.UserRepository,
	fileRepo ports.FileRepository,
	scoper ports.PathScoper,
	tokens *TokenService,
	signupEnabled bool,
	defaultQuotaBytes int64,
) *OAuthLoginService {
	return &OAuthLoginService{
		providers:         providers,
		users:             users,
		fileRepo:          fileRepo,
		scoper:            scoper,
		tokens:            tokens,
		signupEnabled:     signupEnabled,
		defaultQuotaBytes: defaultQuotaBytes,
		states:            newOAuthStateStore(),
	}
}

// NewOAuthLoginServiceForTest builds a service with no providers configured,
// for handler tests that only exercise failure paths.
func NewOAuthLoginServiceForTest() *OAuthLoginService {
	return NewOAuthLoginService(map[string]*auth.OAuthProvider{}, nil, nil, nil, nil, true, 0)
}

func (s *OAuthLoginService) ProviderNames() []string {
	return auth.OAuthProviderNames(s.providers)
}

func (s *OAuthLoginService) GetProvider(name string) *auth.OAuthProvider {
	return s.providers[name]
}

// Begin creates a state value bound to the provider and returns the
// authorization URL the browser should open. The state must also be stored in
// an HttpOnly cookie by the primary adapter and verified on callback.
func (s *OAuthLoginService) Begin(provider, callbackURI string) (string, string, error) {
	p := s.providers[provider]
	if p == nil {
		return "", "", domainerrors.NewValidationError("provider", provider, "oauth provider not configured")
	}
	state, err := randomState()
	if err != nil {
		return "", "", err
	}
	if !s.states.put(state, provider) {
		return "", "", domainerrors.NewValidationError("state", "", "too many pending oauth logins")
	}
	return p.AuthURL(callbackURI, state), state, nil
}

// Complete exchanges the code, upserts the account, and returns a token pair.
// browserState must be the state cookie value issued to this browser; it is
// compared against the query state to block login CSRF.
func (s *OAuthLoginService) Complete(provider, code, state, browserState, callbackURI string) (*TokenPair, error) {
	if browserState == "" || state == "" || !constantTimeEqual(browserState, state) {
		return nil, domainerrors.NewValidationError("state", "", "invalid oauth state")
	}
	if !s.states.consume(state, provider) {
		return nil, domainerrors.NewValidationError("state", "", "invalid oauth state")
	}
	p := s.providers[provider]
	if p == nil {
		return nil, domainerrors.NewValidationError("provider", provider, "oauth provider not configured")
	}
	profile, err := p.Exchange(code, callbackURI)
	if err != nil {
		return nil, domainerrors.NewValidationError("oauth", provider, "code exchange failed")
	}

	user, err := s.upsert(provider, profile)
	if err != nil {
		return nil, err
	}
	return s.tokens.IssueFor(user)
}

func (s *OAuthLoginService) upsert(provider string, profile *auth.OAuthProfile) (*models.User, error) {
	if profile.Subject == "" {
		return nil, domainerrors.NewValidationError("subject", provider, "oauth profile missing stable subject")
	}

	if existing, err := s.users.FindByOAuth(provider, profile.Subject); err == nil {
		return existing, nil
	} else if !errors.Is(err, domainerrors.ErrUserNotFound) {
		return nil, err
	}

	username := sanitizeOAuthUsername(profile.Username)
	if username == "" {
		return nil, domainerrors.NewValidationError("username", profile.Username, "could not derive a valid username from oauth profile")
	}

	existing, err := s.users.FindByUsername(username)
	if err == nil {
		// Never claim a password-protected username (squatter / homograph risk).
		if existing.PasswordHash != "" {
			return nil, &domainerrors.ForbiddenError{Action: "register via oauth with username " + username}
		}
		sameIdentity := existing.OAuthProvider == provider && existing.OAuthSubject != "" && existing.OAuthSubject == profile.Subject
		if sameIdentity {
			return existing, nil
		}
		// Different OAuth identity (or unbound legacy) owns this name: pick a
		// fresh one rather than signing into their account.
		username, err = disambiguateOAuthUsername(username, provider, profile.Subject, func(candidate string) bool {
			_, findErr := s.users.FindByUsername(candidate)
			return errors.Is(findErr, domainerrors.ErrUserNotFound)
		})
		if err != nil {
			return nil, err
		}
	} else if !errors.Is(err, domainerrors.ErrUserNotFound) {
		return nil, err
	}

	if !s.signupEnabled {
		return nil, &domainerrors.ForbiddenError{Action: "register via oauth"}
	}

	user := &models.User{
		Username:      username,
		PasswordHash:  "",
		Role:          models.RoleMember,
		IsAdmin:       false,
		Enabled:       true,
		QuotaBytes:    s.defaultQuotaBytes,
		CreatedAt:     time.Now().UTC(),
		OAuthProvider: provider,
		OAuthSubject:  profile.Subject,
	}
	count, err := s.users.CountUsers()
	if err != nil {
		return nil, err
	}
	if count == 0 {
		user.Role = models.RoleAdmin
		user.IsAdmin = true
	}
	if err := s.users.CreateUser(user); err != nil {
		if errors.Is(err, domainerrors.ErrUserAlreadyExists) {
			// Concurrent create with the same name — re-check OAuth binding.
			if again, ferr := s.users.FindByOAuth(provider, profile.Subject); ferr == nil {
				return again, nil
			}
		}
		return nil, err
	}
	_ = s.fileRepo.CreateDirectory(s.scoper.PrivatePrefix(user.Username))
	user.PasswordHash = ""
	return user, nil
}

// disambiguateOAuthUsername appends a provider tag and a short subject digest
// until the candidate is free, keeping the shared username pattern.
func disambiguateOAuthUsername(base, provider, subject string, available func(string) bool) (string, error) {
	tag := "id"
	switch provider {
	case "github":
		tag = "gh"
	case "google":
		tag = "go"
	}
	sum := sha256.Sum256([]byte(provider + ":" + subject))
	suffix := hex.EncodeToString(sum[:4])
	for i := 0; i < 8; i++ {
		candidate := truncateUsername(base + "-" + tag + "-" + suffix)
		if usernamePattern.MatchString(candidate) && available(candidate) {
			return candidate, nil
		}
		sum = sha256.Sum256(sum[:])
		suffix = hex.EncodeToString(sum[:4])
	}
	return "", domainerrors.NewValidationError("username", base, "could not allocate a unique oauth username")
}

func truncateUsername(s string) string {
	if len(s) > 32 {
		s = strings.TrimRight(s[:32], ".-_")
	}
	return s
}

func constantTimeEqual(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// sanitizeOAuthUsername lowercases the raw name, replaces invalid characters,
// and enforces the shared username pattern (3-32 chars).
func sanitizeOAuthUsername(raw string) string {
	raw = strings.TrimSpace(strings.ToLower(raw))
	var b strings.Builder
	for _, r := range raw {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '.', r == '-', r == '_':
			b.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r + ('a' - 'A'))
		case r == ' ', r == '@':
			b.WriteRune('-')
		}
	}
	out := strings.Trim(b.String(), ".-_")
	if out == "" {
		return ""
	}
	if len(out) > 32 {
		out = out[:32]
		out = strings.TrimRight(out, ".-_")
	}
	if len(out) < 3 {
		return ""
	}
	if out[0] < 'a' || out[0] > 'z' {
		if out[0] < '0' || out[0] > '9' {
			return ""
		}
	}
	if !usernamePattern.MatchString(out) {
		return ""
	}
	return out
}

func randomState() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

type oauthStateEntry struct {
	provider  string
	createdAt time.Time
}

type oauthStateStore struct {
	mu      sync.Mutex
	entries map[string]oauthStateEntry
}

func newOAuthStateStore() *oauthStateStore {
	return &oauthStateStore{entries: make(map[string]oauthStateEntry)}
}

func (s *oauthStateStore) put(state, provider string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for k, e := range s.entries {
		if now.Sub(e.createdAt) > 10*time.Minute {
			delete(s.entries, k)
		}
	}
	// Fail closed when full: wiping every live state would let an attacker
	// knock out in-flight logins (and would not reduce forgery risk).
	if len(s.entries) >= 10000 {
		return false
	}
	s.entries[state] = oauthStateEntry{provider: provider, createdAt: now}
	return true
}

func (s *oauthStateStore) consume(state, provider string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.entries[state]
	if !ok {
		return false
	}
	delete(s.entries, state)
	return e.provider == provider && time.Since(e.createdAt) <= 10*time.Minute
}
