package services

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/EslamYasser-Dev/simple-file-share/domain/policy"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/auth"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/fs"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/memory"
)

type googleTestKeys struct {
	priv *rsa.PrivateKey
	pub  *rsa.PublicKey
}

func newGoogleTestKeys(t *testing.T) *googleTestKeys {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	return &googleTestKeys{priv: priv, pub: &priv.PublicKey}
}

func (k *googleTestKeys) source() GoogleCertSource {
	return func(_ context.Context) (map[string]*rsa.PublicKey, error) {
		return map[string]*rsa.PublicKey{"test-kid": k.pub}, nil
	}
}

// mintGoogleToken signs claims with the test key, stamping the kid header
// the way Google's certs do.
func (k *googleTestKeys) mint(t *testing.T, claims jwt.MapClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "test-kid"
	signed, err := token.SignedString(k.priv)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	return signed
}

func googleClaims() jwt.MapClaims {
	now := time.Now()
	return jwt.MapClaims{
		"iss":   "accounts.google.com",
		"aud":   "test-client-id",
		"sub":   "google-sub-123",
		"email": "bob@gmail.com",
		"name":  "Bob",
		"exp":   now.Add(time.Hour).Unix(),
		"iat":   now.Unix(),
	}
}

// TestGoogleVerifyAcceptsValidToken covers the happy path plus every
// rejection: audience, expiry, issuer, signature, key, algorithm, and the
// unconfigured server.
func TestGoogleVerifyAcceptsValidToken(t *testing.T) {
	keys := newGoogleTestKeys(t)
	verifier := NewGoogleIDTokenVerifier("test-client-id", keys.source())

	identity, err := verifier.Verify(context.Background(), keys.mint(t, googleClaims()))
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if identity.Subject != "google-sub-123" || identity.Email != "bob@gmail.com" || identity.Name != "Bob" {
		t.Fatalf("identity = %+v", identity)
	}
}

func TestGoogleVerifyRejects(t *testing.T) {
	keys := newGoogleTestKeys(t)
	other := newGoogleTestKeys(t)
	verifier := NewGoogleIDTokenVerifier("test-client-id", keys.source())

	cases := map[string]func(jwt.MapClaims) (string, *GoogleIDTokenVerifier){
		"wrong audience": func(c jwt.MapClaims) (string, *GoogleIDTokenVerifier) {
			c["aud"] = "other-client"
			return keys.mint(t, c), verifier
		},
		"expired": func(c jwt.MapClaims) (string, *GoogleIDTokenVerifier) {
			c["exp"] = time.Now().Add(-time.Hour).Unix()
			return keys.mint(t, c), verifier
		},
		"wrong issuer": func(c jwt.MapClaims) (string, *GoogleIDTokenVerifier) {
			c["iss"] = "https://evil.example.com"
			return keys.mint(t, c), verifier
		},
		"bad signature": func(c jwt.MapClaims) (string, *GoogleIDTokenVerifier) {
			return other.mint(t, c), verifier
		},
		"wrong algorithm": func(c jwt.MapClaims) (string, *GoogleIDTokenVerifier) {
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
			token.Header["kid"] = "test-kid"
			signed, err := token.SignedString([]byte("secret"))
			if err != nil {
				t.Fatalf("sign: %v", err)
			}
			return signed, verifier
		},
		"missing subject": func(c jwt.MapClaims) (string, *GoogleIDTokenVerifier) {
			delete(c, "sub")
			return keys.mint(t, c), verifier
		},
		"unconfigured server": func(c jwt.MapClaims) (string, *GoogleIDTokenVerifier) {
			return keys.mint(t, c), NewGoogleIDTokenVerifier("", keys.source())
		},
	}
	for name, mutate := range cases {
		token, v := mutate(googleClaims())
		if _, err := v.Verify(context.Background(), token); err == nil {
			t.Errorf("%s: expected rejection", name)
		}
	}
	if _, err := verifier.Verify(context.Background(), "not-a-jwt"); err == nil {
		t.Error("malformed: expected rejection")
	}
}

func newGoogleLoginFixture(t *testing.T, keys *googleTestKeys) *GoogleLoginService {
	t.Helper()
	dir := t.TempDir()
	userRepo := fs.NewUserFileRepository(dir)
	fileRepo := fs.NewIndexedFileRepository(fs.NewLocalFileRepository(dir), memory.NewFileIndexRepository(), nil)
	scoper := policy.NewPathScoper()
	hasher := auth.NewPBKDF2Hasher()
	tokens := NewTokenService(
		NewAuthenticateService(auth.NewUserAuthProvider(userRepo, hasher)),
		auth.NewJWTManager("test-secret", time.Hour),
		userRepo,
	)
	oauthSvc := NewOAuthLoginService(
		map[string]*auth.OAuthProvider{},
		userRepo, fileRepo, scoper, tokens, true, 0,
	)
	verifier := NewGoogleIDTokenVerifier("test-client-id", keys.source())
	return NewGoogleLoginService(oauthSvc, verifier)
}

// TestGoogleLoginCreatesThenReturnsAccount verifies the exchange mints an
// app session, creates the account once (IsNew), and returns to it after.
func TestGoogleLoginCreatesThenReturnsAccount(t *testing.T) {
	keys := newGoogleTestKeys(t)
	svc := newGoogleLoginFixture(t, keys)
	ctx := context.Background()

	first, err := svc.Execute(ctx, keys.mint(t, googleClaims()))
	if err != nil {
		t.Fatalf("first login: %v", err)
	}
	if first.Pair == nil || first.Pair.AccessToken == "" {
		t.Fatal("missing app session")
	}
	if !first.IsNew {
		t.Fatal("first login must report a new account")
	}
	if first.User.OAuthProvider != "google" || first.User.OAuthSubject != "google-sub-123" {
		t.Fatalf("binding = %+v", first.User)
	}
	if first.User.PasswordHash != "" {
		t.Fatal("oauth account must have no password")
	}

	second, err := svc.Execute(ctx, keys.mint(t, googleClaims()))
	if err != nil {
		t.Fatalf("second login: %v", err)
	}
	if second.IsNew {
		t.Fatal("second login must not report a new account")
	}
	if second.User.Username != first.User.Username {
		t.Fatalf("usernames diverged: %q vs %q", second.User.Username, first.User.Username)
	}

	// The minted session authenticates.
	stored, err := svc.oauth.tokens.Authenticate(first.Pair.AccessToken)
	if err != nil {
		t.Fatalf("authenticate: %v", err)
	}
	if stored.Username != first.User.Username {
		t.Fatalf("session user = %q", stored.Username)
	}
}
