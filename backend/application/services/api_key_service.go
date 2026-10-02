package services

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"time"

	domainerrors "github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

// APIKeyPrefix marks bearer credentials that are API keys rather than JWTs.
const APIKeyPrefix = "sfs_"

// apiKeyAlphabet is 64 characters (a-zA-Z0-9); 256 divides evenly by 64 so
// mapping random bytes with i%64 is unbiased.
const apiKeyAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// apiKeyTouchWindow throttles LastUsedAt rewrites to at most one per window
// so a hot key does not rewrite the file on every request.
const apiKeyTouchWindow = 5 * time.Minute

// APIKeyService mints, authenticates, lists, and revokes API keys.
// The plaintext key is returned exactly once at issue time; authentication
// only ever verifies a PBKDF2 hash.
type APIKeyService struct {
	keys  ports.APIKeyRepository
	users ports.UserRepository
	hash  ports.PasswordHasher

	// now is stubbed in tests.
	now func() time.Time
}

// NewAPIKeyService wires the service with its stores.
func NewAPIKeyService(keys ports.APIKeyRepository, users ports.UserRepository, hash ports.PasswordHasher) *APIKeyService {
	return &APIKeyService{keys: keys, users: users, hash: hash, now: time.Now}
}

// Issue mints a new key for the actor and returns the full plaintext
// "sfs_<id>_<secret>" alongside the stored record (secret hash filled in).
// An empty scope defaults to read; expiresIn <= 0 means the key never
// expires.
func (s *APIKeyService) Issue(actor *models.User, name, scope string, expiresIn time.Duration) (string, models.APIKey, error) {
	owner := ""
	if actor != nil {
		owner = strings.TrimSpace(actor.Username)
	}
	if owner == "" {
		return "", models.APIKey{}, domainerrors.NewValidationError("actor", owner, "authentication required")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return "", models.APIKey{}, domainerrors.NewValidationError("name", name, "name is required")
	}
	if len(name) > 80 {
		return "", models.APIKey{}, domainerrors.NewValidationError("name", name, "name must be at most 80 characters")
	}
	scope = strings.TrimSpace(scope)
	if scope == "" {
		scope = models.APIKeyScopeRead
	}
	if models.APIKeyScopeRank(scope) == 0 {
		return "", models.APIKey{}, domainerrors.NewValidationError("scope", scope, "scope must be one of read, write, admin")
	}
	if expiresIn < 0 {
		return "", models.APIKey{}, domainerrors.NewValidationError("expiresIn", expiresIn, "must be a positive number of seconds (0 = never)")
	}

	idBytes := make([]byte, 8)
	if _, err := rand.Read(idBytes); err != nil {
		return "", models.APIKey{}, err
	}
	secret := randomKeySecret()
	key := models.APIKey{
		ID:        hex.EncodeToString(idBytes),
		Name:      name,
		Owner:     owner,
		Scope:     scope,
		CreatedAt: s.now().UTC(),
	}
	if expiresIn > 0 {
		key.ExpiresAt = s.now().UTC().Add(expiresIn)
	}
	hash, err := s.hash.Hash(secret)
	if err != nil {
		return "", models.APIKey{}, err
	}
	key.SecretHash = hash
	if err := s.keys.Create(key); err != nil {
		return "", models.APIKey{}, err
	}
	plaintext := APIKeyPrefix + key.ID + "_" + secret
	return plaintext, key, nil
}

// Authenticate verifies a presented key string and resolves it to its owner's
// account plus the key's scope. Every failure mode (unknown id, bad secret,
// expired key, missing/disabled owner) maps to ports.ErrUnauthorized so
// callers learn nothing about which part was wrong.
func (s *APIKeyService) Authenticate(presented string) (*models.User, string, error) {
	id, secret, ok := parseAPIKey(presented)
	if !ok {
		return nil, "", ports.ErrUnauthorized
	}
	key, err := s.keys.Find(id)
	if err != nil {
		return nil, "", ports.ErrUnauthorized
	}
	if !s.hash.Verify(secret, key.SecretHash) {
		return nil, "", ports.ErrUnauthorized
	}
	now := s.now().UTC()
	if !key.ExpiresAt.IsZero() && !key.ExpiresAt.After(now) {
		return nil, "", ports.ErrUnauthorized
	}
	user, err := s.users.FindByUsername(key.Owner)
	if err != nil || !user.Enabled {
		return nil, "", ports.ErrUnauthorized
	}
	s.touch(key, now)
	return user, key.Scope, nil
}

// List returns the owner's keys (metadata; hashes are never copied out).
func (s *APIKeyService) List(actor *models.User) ([]models.APIKey, error) {
	if actor == nil || strings.TrimSpace(actor.Username) == "" {
		return nil, domainerrors.NewValidationError("actor", "", "authentication required")
	}
	return s.keys.List(actor.Username)
}

// Revoke deletes one of the owner's keys; another owner's key is reported
// as not found so callers cannot probe for ids they do not hold.
func (s *APIKeyService) Revoke(actor *models.User, id string) error {
	if actor == nil || strings.TrimSpace(actor.Username) == "" {
		return domainerrors.NewValidationError("actor", "", "authentication required")
	}
	key, err := s.keys.Find(id)
	if err != nil {
		return err
	}
	if key.Owner != actor.Username {
		return domainerrors.ErrNotFound
	}
	return s.keys.Delete(id)
}

// touch records last use at most once per apiKeyTouchWindow (best effort —
// a failed write never fails authentication).
func (s *APIKeyService) touch(key *models.APIKey, now time.Time) {
	if !key.LastUsedAt.IsZero() && now.Sub(key.LastUsedAt) < apiKeyTouchWindow {
		return
	}
	updated := *key
	updated.LastUsedAt = now
	_ = s.keys.Save(updated)
}

// parseAPIKey splits "sfs_<16-hex-id>_<secret>".
func parseAPIKey(presented string) (id, secret string, ok bool) {
	rest, found := strings.CutPrefix(strings.TrimSpace(presented), APIKeyPrefix)
	if !found {
		return "", "", false
	}
	id, secret, found = strings.Cut(rest, "_")
	if !found || len(id) != 16 || secret == "" {
		return "", "", false
	}
	if _, err := hex.DecodeString(id); err != nil {
		return "", "", false
	}
	return id, secret, true
}

// randomKeySecret returns 32 characters from apiKeyAlphabet.
func randomKeySecret() string {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		panic("crypto/rand unavailable: " + err.Error())
	}
	out := make([]byte, 32)
	for i, b := range buf {
		out[i] = apiKeyAlphabet[int(b)%len(apiKeyAlphabet)]
	}
	return string(out)
}
