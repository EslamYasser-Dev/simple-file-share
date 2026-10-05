package auth

import (
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

// PBKDF2Hasher encodes passwords with salted PBKDF2-HMAC-SHA256 using the
// standard library only, preserving the project's zero-dependency constraint.
// Encoded format: pbkdf2-sha256$<iterations>$<salthx>$<keyhex>
//
// Successful verifications are memoized for a short TTL so per-request Basic
// auth, API-key checks, and share-password checks do not re-run the 600k
// iteration derivation on every call. Only successes are cached, the key
// includes the stored hash (so a password change invalidates it), and
// failures always pay full price.
type PBKDF2Hasher struct {
	iterations int
	keyLen     int
	saltLen    int

	cacheMu    sync.Mutex
	cache      map[[32]byte]time.Time
	cacheOrder [][32]byte
}

const (
	defaultIterations = 600000
	defaultKeyLen     = 32
	defaultSaltLen    = 16

	verifyCacheTTL = 10 * time.Minute
	verifyCacheMax = 2048
)

var _ ports.PasswordHasher = (*PBKDF2Hasher)(nil)

func NewPBKDF2Hasher() *PBKDF2Hasher {
	return &PBKDF2Hasher{
		iterations: defaultIterations,
		keyLen:     defaultKeyLen,
		saltLen:    defaultSaltLen,
		cache:      make(map[[32]byte]time.Time),
	}
}

// Hash derives a salted key from the password and encodes it for storage.
func (h *PBKDF2Hasher) Hash(password string) (string, error) {
	salt := make([]byte, h.saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate salt: %w", err)
	}

	key, err := pbkdf2.Key(sha256.New, password, salt, h.iterations, h.keyLen)
	if err != nil {
		return "", fmt.Errorf("derive key: %w", err)
	}
	return fmt.Sprintf("pbkdf2-sha256$%d$%x$%x", h.iterations, salt, key), nil
}

// Verify parses the encoded hash and compares a freshly derived key in
// constant time, consulting the success memo first.
func (h *PBKDF2Hasher) Verify(password, encoded string) bool {
	cacheKey := verifyCacheKey(password, encoded)
	now := time.Now()
	h.cacheMu.Lock()
	if exp, ok := h.cache[cacheKey]; ok && now.Before(exp) {
		h.cacheMu.Unlock()
		return true
	}
	h.cacheMu.Unlock()

	if !h.verifyUncached(password, encoded) {
		return false
	}
	h.rememberVerified(cacheKey, now)
	return true
}

func (h *PBKDF2Hasher) verifyUncached(password, encoded string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 4 || parts[0] != "pbkdf2-sha256" {
		return false
	}

	iterations, err := strconv.Atoi(parts[1])
	if err != nil || iterations < 1 {
		return false
	}

	salt, err := hex.DecodeString(parts[2])
	if err != nil || len(salt) == 0 {
		return false
	}

	expected, err := hex.DecodeString(parts[3])
	if err != nil {
		return false
	}

	key, err := pbkdf2.Key(sha256.New, password, salt, iterations, len(expected))
	if err != nil {
		return false
	}
	return subtle.ConstantTimeCompare(key, expected) == 1
}

// verifyCacheKey binds the presented password to the specific stored hash, so
// changing the password (new salt) naturally retires every old entry.
func verifyCacheKey(password, encoded string) [32]byte {
	hsh := sha256.New()
	_, _ = hsh.Write([]byte(password))
	_, _ = hsh.Write([]byte{0})
	_, _ = hsh.Write([]byte(encoded))
	var key [32]byte
	copy(key[:], hsh.Sum(nil))
	return key
}

func (h *PBKDF2Hasher) rememberVerified(key [32]byte, now time.Time) {
	h.cacheMu.Lock()
	defer h.cacheMu.Unlock()

	if h.cache == nil {
		h.cache = make(map[[32]byte]time.Time)
	}
	if _, ok := h.cache[key]; !ok {
		h.cacheOrder = append(h.cacheOrder, key)
	}
	h.cache[key] = now.Add(verifyCacheTTL)

	// Evict from the oldest end: expired entries always, live ones once over
	// capacity. Map and order stay in lockstep because keys enter both
	// together above.
	for len(h.cacheOrder) > 0 {
		oldest := h.cacheOrder[0]
		exp, ok := h.cache[oldest]
		expired := !ok || !now.Before(exp)
		if !expired && len(h.cache) <= verifyCacheMax {
			break
		}
		h.cacheOrder = h.cacheOrder[1:]
		delete(h.cache, oldest)
	}
}
