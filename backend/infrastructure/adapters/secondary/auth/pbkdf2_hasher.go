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

	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

// PBKDF2Hasher encodes passwords with salted PBKDF2-HMAC-SHA256 using the
// standard library only, preserving the project's zero-dependency constraint.
// Encoded format: pbkdf2-sha256$<iterations>$<salthx>$<keyhex>
type PBKDF2Hasher struct {
	iterations int
	keyLen     int
	saltLen    int
}

const (
	defaultIterations = 600000
	defaultKeyLen     = 32
	defaultSaltLen    = 16
)

var _ ports.PasswordHasher = (*PBKDF2Hasher)(nil)

func NewPBKDF2Hasher() *PBKDF2Hasher {
	return &PBKDF2Hasher{
		iterations: defaultIterations,
		keyLen:     defaultKeyLen,
		saltLen:    defaultSaltLen,
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
// constant time.
func (h *PBKDF2Hasher) Verify(password, encoded string) bool {
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
