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
	iat := now.Unix()
	if floor := m.revocations.subjectIssueFloor(subject); floor > iat {
		iat = floor
	}
	exp := time.Unix(iat, 0).Add(m.ttl)
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
		IssuedAt:  iat,
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
	if m.revocations.isSubjectRevoked(claims.Subject, claims.IssuedAt) {
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

// RevokeSubject invalidates tokens issued at or before now for subject until
// the until horizon (used as a record-pruning deadline). Tokens issued after
// the revoke still verify, so re-login works immediately.
func (m *JWTManager) RevokeSubject(subject string, until time.Time) {
	if subject == "" {
		return
	}
	m.revocations.revokeSubject(subject, until)
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
	Entries  map[string]string     `json:"entries"`
	Subjects map[string]subjectDoc `json:"subjects,omitempty"`
}

type subjectDoc struct {
	// CutoffUnix rejects tokens with iat <= cutoff (second precision).
	CutoffUnix int64 `json:"cutoffUnix"`
	// KeepUntil is when the record may be pruned (not when access resumes).
	KeepUntil string `json:"keepUntil"`
}

// revocationList is a bounded set of token IDs whose expiry has not passed.
// When path is non-empty, mutations are flushed to disk so restarts do not
// resurrect revoked sessions.
type revocationList struct {
	mu       sync.Mutex
	path     string
	entries  map[string]time.Time
	subjects map[string]subjectDoc
}

func newRevocationList(path string) *revocationList {
	l := &revocationList{
		path:     path,
		entries:  make(map[string]time.Time),
		subjects: make(map[string]subjectDoc),
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

func (l *revocationList) revokeSubject(subject string, until time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.subjects == nil {
		l.subjects = make(map[string]subjectDoc)
	}
	now := time.Now()
	cutoff := now.Unix()
	keep := until
	if !keep.After(now) {
		keep = now.Add(time.Hour)
	}
	if cur, ok := l.subjects[subject]; ok && cur.CutoffUnix > cutoff {
		cutoff = cur.CutoffUnix
	}
	l.subjects[subject] = subjectDoc{CutoffUnix: cutoff, KeepUntil: keep.UTC().Format(time.RFC3339)}
	l.saveLocked()
}

// isSubjectRevoked reports whether a token for subject issued at iat is
// covered by a subject revoke (iat <= cutoff) while the record is still live.
func (l *revocationList) isSubjectRevoked(subject string, iat int64) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	rev, ok := l.subjects[subject]
	if !ok {
		return false
	}
	keep, err := time.Parse(time.RFC3339, rev.KeepUntil)
	if err != nil || time.Now().After(keep) {
		delete(l.subjects, subject)
		return false
	}
	return iat <= rev.CutoffUnix
}

// subjectIssueFloor returns the minimum iat a newly issued token may use so
// it is never covered by an active subject revoke.
func (l *revocationList) subjectIssueFloor(subject string) int64 {
	l.mu.Lock()
	defer l.mu.Unlock()
	rev, ok := l.subjects[subject]
	if !ok {
		return 0
	}
	keep, err := time.Parse(time.RFC3339, rev.KeepUntil)
	if err != nil || time.Now().After(keep) {
		delete(l.subjects, subject)
		return 0
	}
	return rev.CutoffUnix + 1
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
	if l.subjects != nil {
		for k, rev := range l.subjects {
			keep, err := time.Parse(time.RFC3339, rev.KeepUntil)
			if err != nil || now.After(keep) {
				delete(l.subjects, k)
			}
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
	if err := json.Unmarshal(data, &doc); err != nil {
		return
	}
	now := time.Now()
	if doc.Entries != nil {
		for id, raw := range doc.Entries {
			exp, err := time.Parse(time.RFC3339, raw)
			if err != nil || now.After(exp) {
				continue
			}
			l.entries[id] = exp
		}
	}
	if doc.Subjects != nil {
		if l.subjects == nil {
			l.subjects = make(map[string]subjectDoc)
		}
		for sub, sd := range doc.Subjects {
			keep, err := time.Parse(time.RFC3339, sd.KeepUntil)
			if err != nil || now.After(keep) {
				continue
			}
			l.subjects[sub] = sd
		}
	}
}

func (l *revocationList) saveLocked() {
	if l.path == "" {
		return
	}
	doc := revocationDoc{
		Entries:  make(map[string]string, len(l.entries)),
		Subjects: make(map[string]subjectDoc, len(l.subjects)),
	}
	for id, exp := range l.entries {
		doc.Entries[id] = exp.UTC().Format(time.RFC3339)
	}
	for sub, sd := range l.subjects {
		doc.Subjects[sub] = sd
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
