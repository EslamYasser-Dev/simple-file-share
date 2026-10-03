package services

import (
	"bytes"
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"

	domainerrors "github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/auth"
)

// googleCertURL serves Google's current OIDC signing keys as a JWK set.
const googleCertURL = "https://www.googleapis.com/oauth2/v3/certs"

// googleIssuers are the only accepted ID-token issuers.
var googleIssuers = map[string]bool{
	"accounts.google.com":         true,
	"https://accounts.google.com": true,
}

// GoogleIDToken is a verified Google identity: stable subject plus the
// profile fields used to mint or match a local account.
type GoogleIDToken struct {
	Subject string
	Email   string
	Name    string
}

// GoogleCertSource returns signing keys by key ID. The production source
// fetches Google's JWK set with a TTL cache; tests inject a static key.
type GoogleCertSource func(ctx context.Context) (map[string]*rsa.PublicKey, error)

// GoogleIDTokenVerifier validates Google OIDC ID tokens (RS256, audience,
// issuer, expiry) without trusting any client claim.
type GoogleIDTokenVerifier struct {
	clientIDs []string
	source    GoogleCertSource
	now       func() time.Time
}

// NewGoogleIDTokenVerifier accepts a comma-separated audience list; empty
// disables verification (every token is rejected, never accepted).
func NewGoogleIDTokenVerifier(clientIDs string, source GoogleCertSource) *GoogleIDTokenVerifier {
	var ids []string
	for _, id := range strings.Split(clientIDs, ",") {
		if id = strings.TrimSpace(id); id != "" {
			ids = append(ids, id)
		}
	}
	if source == nil {
		source = NewGoogleCertCache(http.DefaultClient, time.Hour)
	}
	return &GoogleIDTokenVerifier{clientIDs: ids, source: source, now: time.Now}
}

// Verify checks signature, issuer, audience, and expiry, returning the
// identity on success.
func (v *GoogleIDTokenVerifier) Verify(ctx context.Context, idToken string) (*GoogleIDToken, error) {
	if len(v.clientIDs) == 0 {
		return nil, domainerrors.NewValidationError("id_token", nil, "google sign-in is not configured on this server")
	}
	if strings.Count(idToken, ".") != 2 {
		return nil, domainerrors.NewValidationError("id_token", nil, "malformed id token")
	}
	keys, err := v.source(ctx)
	if err != nil {
		return nil, err
	}
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(idToken, claims, func(t *jwt.Token) (any, error) {
		if t.Method.Alg() != jwt.SigningMethodRS256.Alg() {
			return nil, domainerrors.NewValidationError("id_token", nil, "unexpected signing algorithm")
		}
		kid, _ := t.Header["kid"].(string)
		key, ok := keys[kid]
		if !ok {
			return nil, domainerrors.NewValidationError("id_token", nil, "unknown signing key")
		}
		return key, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()}))
	if err != nil || !token.Valid {
		return nil, domainerrors.NewValidationError("id_token", nil, "invalid id token signature")
	}

	issuer, _ := claims["iss"].(string)
	if !googleIssuers[issuer] {
		return nil, domainerrors.NewValidationError("id_token", nil, "untrusted token issuer")
	}
	// Aud arrives as a string or a list; accept either shape.
	matched := false
	for _, id := range v.clientIDs {
		if audMatches(claims["aud"], id) {
			matched = true
			break
		}
	}
	if !matched {
		return nil, domainerrors.NewValidationError("id_token", nil, "token audience mismatch")
	}
	if exp, err := claims.GetExpirationTime(); err != nil || exp == nil || v.now().After(exp.Time) {
		return nil, domainerrors.NewValidationError("id_token", nil, "id token expired")
	}
	subject, err := claims.GetSubject()
	if err != nil || subject == "" {
		return nil, domainerrors.NewValidationError("id_token", nil, "id token has no subject")
	}
	email, _ := claims["email"].(string)
	name, _ := claims["name"].(string)
	return &GoogleIDToken{Subject: subject, Email: email, Name: name}, nil
}

// audMatches compares one accepted audience against the claim in either its
// string or list shape.
func audMatches(claim any, accepted string) bool {
	switch aud := claim.(type) {
	case string:
		return aud == accepted
	case []any:
		for _, a := range aud {
			if s, ok := a.(string); ok && s == accepted {
				return true
			}
		}
	case []string:
		for _, s := range aud {
			if s == accepted {
				return true
			}
		}
	}
	return false
}

// googleCertCache fetches Google's JWK set once per TTL, keyed by kid.
type googleCertCache struct {
	client *http.Client
	ttl    time.Duration

	mu      sync.Mutex
	keys    map[string]*rsa.PublicKey
	expires time.Time
}

// NewGoogleCertCache returns a GoogleCertSource with TTL caching.
func NewGoogleCertCache(client *http.Client, ttl time.Duration) GoogleCertSource {
	if client == nil {
		client = http.DefaultClient
	}
	if ttl <= 0 {
		ttl = time.Hour
	}
	c := &googleCertCache{client: client, ttl: ttl}
	return c.get
}

func (c *googleCertCache) get(ctx context.Context) (map[string]*rsa.PublicKey, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if time.Now().Before(c.expires) && c.keys != nil {
		return c.keys, nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, googleCertURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch google certs: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch google certs: status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("read google certs: %w", err)
	}
	var set struct {
		Keys []struct {
			Kid string `json:"kid"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}
	if err := json.NewDecoder(bytes.NewReader(body)).Decode(&set); err != nil {
		return nil, fmt.Errorf("parse google certs: %w", err)
	}
	keys := make(map[string]*rsa.PublicKey, len(set.Keys))
	for _, k := range set.Keys {
		nBytes, err := base64.RawURLEncoding.DecodeString(k.N)
		if err != nil {
			continue
		}
		eBytes, err := base64.RawURLEncoding.DecodeString(k.E)
		if err != nil {
			continue
		}
		e := 0
		for _, b := range eBytes {
			e = e<<8 + int(b)
		}
		if k.Kid == "" || len(nBytes) == 0 || e == 0 {
			continue
		}
		keys[k.Kid] = &rsa.PublicKey{N: new(big.Int).SetBytes(nBytes), E: e}
	}
	if len(keys) == 0 {
		return nil, fmt.Errorf("fetch google certs: no usable keys")
	}
	c.keys = keys
	c.expires = time.Now().Add(c.ttl)
	return keys, nil
}

// GoogleLoginService exchanges a verified Google identity for an app token,
// creating the account on first use through the shared OAuth upsert (same
// username rules, quota, and admin-seeding as the browser flow).
type GoogleLoginService struct {
	oauth    *OAuthLoginService
	verifier *GoogleIDTokenVerifier
}

func NewGoogleLoginService(oauth *OAuthLoginService, verifier *GoogleIDTokenVerifier) *GoogleLoginService {
	return &GoogleLoginService{oauth: oauth, verifier: verifier}
}

// GoogleLoginResult is the app session plus whether the account was created.
type GoogleLoginResult struct {
	Pair  *TokenPair
	User  *models.User
	IsNew bool
}

// Execute verifies the ID token and returns an app session for the linked
// (or newly created) account.
func (s *GoogleLoginService) Execute(ctx context.Context, idToken string) (*GoogleLoginResult, error) {
	identity, err := s.verifier.Verify(ctx, idToken)
	if err != nil {
		return nil, err
	}
	profile := &auth.OAuthProfile{
		Username: googleUsername(identity),
		Email:    identity.Email,
		Subject:  identity.Subject,
	}
	user, isNew, err := s.oauth.UpsertProfile("google", profile)
	if err != nil {
		return nil, err
	}
	pair, err := s.oauth.tokens.IssueFor(user)
	if err != nil {
		return nil, err
	}
	return &GoogleLoginResult{Pair: pair, User: user, IsNew: isNew}, nil
}

// googleUsername prefers the email local part (stable, unique-ish), then the
// display name; the upsert sanitizes and disambiguates from there.
func googleUsername(identity *GoogleIDToken) string {
	if at := strings.Index(identity.Email, "@"); at > 0 {
		return identity.Email[:at]
	}
	if identity.Name != "" {
		return identity.Name
	}
	return ""
}
