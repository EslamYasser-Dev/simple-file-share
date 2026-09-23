package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
)

type jwtHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

type jwtClaims struct {
	Subject   string `json:"sub"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
	ID        string `json:"jti"`
}

// JWTManager signs and verifies HS256 access tokens using only the standard
// library, preserving the project's zero-dependency constraint.
type JWTManager struct {
	secret      []byte
	ttl         time.Duration
	revocations *revocationList
	now         func() time.Time
}

var _ ports.TokenManager = (*JWTManager)(nil)

// NewJWTManager creates a manager with in-memory revocation only.
func NewJWTManager(secret string, ttl time.Duration) *JWTManager {
	if ttl <= 0 {
		ttl = time.Hour
	}
	return &JWTManager{
		secret:      []byte(secret),
		ttl:         ttl,
		revocations: newRevocationList(""),
		now:         time.Now,
	}
}

// NewJWTManagerWithStore persists revocations to path (JSON) so they survive
// process restarts. An empty path keeps revocations in memory only.
func NewJWTManagerWithStore(secret string, ttl time.Duration, path string) *JWTManager {
	if ttl <= 0 {
		ttl = time.Hour
	}
	return &JWTManager{
		secret:      []byte(secret),
		ttl:         ttl,
		revocations: newRevocationList(path),
		now:         time.Now,
	}
}

func (m *JWTManager) Issue(subject string) (string, ports.TokenClaims, error) {
	now := m.now()
	exp := now.Add(m.ttl)
	jti, err := randomID()
	if err != nil {
		return "", ports.TokenClaims{}, err
	}

	header, err := json.Marshal(jwtHeader{Alg: "HS256", Typ: "JWT"})
	if err != nil {
		return "", ports.TokenClaims{}, err
	}
	payload, err := json.Marshal(jwtClaims{
		Subject:   subject,
		IssuedAt:  now.Unix(),
		ExpiresAt: exp.Unix(),
		ID:        jti,
	})
	if err != nil {
		return "", ports.TokenClaims{}, err
	}

	enc := base64.RawURLEncoding
	signingInput := enc.EncodeToString(header) + "." + enc.EncodeToString(payload)
	sig := m.sign(signingInput)
	token := signingInput + "." + enc.EncodeToString(sig)

	return token, ports.TokenClaims{Subject: subject, ID: jti, ExpiresAt: exp}, nil
}

func (m *JWTManager) Verify(token string) (*ports.TokenClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}
	enc := base64.RawURLEncoding
	signingInput := parts[0] + "." + parts[1]
	sig, err := enc.DecodeString(parts[2])
	if err != nil {
		return nil, ErrInvalidToken
	}
	if !hmac.Equal(sig, m.sign(signingInput)) {
		return nil, ErrInvalidToken
	}

	headerJSON, err := enc.DecodeString(parts[0])
	if err != nil {
		return nil, ErrInvalidToken
	}
	var header jwtHeader
	if err := json.Unmarshal(headerJSON, &header); err != nil || header.Alg != "HS256" {
		return nil, ErrInvalidToken
	}

	payloadJSON, err := enc.DecodeString(parts[1])
	if err != nil {
		return nil, ErrInvalidToken
	}
	var claims jwtClaims
	if err := json.Unmarshal(payloadJSON, &claims); err != nil {
		return nil, ErrInvalidToken
	}
	if claims.Subject == "" || claims.ID == "" {
		return nil, ErrInvalidToken
	}

	exp := time.Unix(claims.ExpiresAt, 0)
	if !m.now().Before(exp) {
		return nil, ErrExpiredToken
	}
	if m.revocations.isRevoked(claims.ID) {
		return nil, ErrInvalidToken
	}

	return &ports.TokenClaims{
		Subject:   claims.Subject,
		ID:        claims.ID,
		ExpiresAt: exp,
	}, nil
}

func (m *JWTManager) Revoke(id string, expiresAt time.Time) {
	m.revocations.revoke(id, expiresAt)
}

func (m *JWTManager) sign(signingInput string) []byte {
	mac := hmac.New(sha256.New, m.secret)
	mac.Write([]byte(signingInput))
	return mac.Sum(nil)
}

func randomID() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

type revocationDoc struct {
	Entries map[string]string `json:"entries"`
}

// revocationList is a bounded set of token IDs whose expiry has not passed.
// When path is non-empty, mutations are flushed to disk so restarts do not
// resurrect revoked sessions.
type revocationList struct {
	mu      sync.Mutex
	path    string
	entries map[string]time.Time
}

func newRevocationList(path string) *revocationList {
	l := &revocationList{
		path:    path,
		entries: make(map[string]time.Time),
	}
	l.load()
	return l
}

func (l *revocationList) revoke(id string, expiresAt time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.pruneLocked(time.Now())
	l.entries[id] = expiresAt
	l.saveLocked()
}

func (l *revocationList) isRevoked(id string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	exp, ok := l.entries[id]
	if !ok {
		return false
	}
	if time.Now().After(exp) {
		delete(l.entries, id)
		return false
	}
	return true
}

func (l *revocationList) pruneLocked(now time.Time) {
	if len(l.entries) <= 10000 {
		return
	}
	for k, exp := range l.entries {
		if now.After(exp) {
			delete(l.entries, k)
		}
	}
}

func (l *revocationList) load() {
	if l.path == "" {
		return
	}
	data, err := os.ReadFile(l.path)
	if err != nil {
		return
	}
	var doc revocationDoc
	if err := json.Unmarshal(data, &doc); err != nil || doc.Entries == nil {
		return
	}
	now := time.Now()
	for id, raw := range doc.Entries {
		exp, err := time.Parse(time.RFC3339, raw)
		if err != nil || now.After(exp) {
			continue
		}
		l.entries[id] = exp
	}
}

func (l *revocationList) saveLocked() {
	if l.path == "" {
		return
	}
	doc := revocationDoc{Entries: make(map[string]string, len(l.entries))}
	for id, exp := range l.entries {
		doc.Entries[id] = exp.UTC().Format(time.RFC3339)
	}
	data, err := json.Marshal(doc)
	if err != nil {
		return
	}
	if err := os.MkdirAll(filepath.Dir(l.path), 0o700); err != nil {
		return
	}
	tmp, err := os.CreateTemp(filepath.Dir(l.path), ".revocations-*.tmp")
	if err != nil {
		return
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return
	}
	if err := tmp.Close(); err != nil {
		return
	}
	_ = os.Rename(tmp.Name(), l.path)
}
