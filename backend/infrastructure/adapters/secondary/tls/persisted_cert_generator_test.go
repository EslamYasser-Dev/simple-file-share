package tls

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// countingGenerator returns a distinct fixed key pair per call and counts
// invocations, letting tests assert memoization and persistence precisely.
type countingGenerator struct {
	calls int
}

func (g *countingGenerator) GenerateCert() ([]byte, []byte, error) {
	g.calls++
	return newTestCert(time.Now().Add(365 * 24 * time.Hour))
}

type failingGenerator struct{}

func (failingGenerator) GenerateCert() ([]byte, []byte, error) {
	return nil, nil, os.ErrPermission
}

func newTestCert(notAfter time.Time) ([]byte, []byte, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "test"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     notAfter,
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		return nil, nil, err
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	return certPEM, keyPEM, nil
}

func TestPersistedGeneratorMemoizes(t *testing.T) {
	base := &countingGenerator{}
	gen := NewPersistedTLSCertGenerator(base, t.TempDir())

	first, _, err := gen.GenerateCert()
	if err != nil {
		t.Fatal(err)
	}
	second, _, err := gen.GenerateCert()
	if err != nil {
		t.Fatal(err)
	}
	if base.calls != 1 {
		t.Fatalf("base called %d times, want 1", base.calls)
	}
	if string(first) != string(second) {
		t.Fatal("second call returned a different certificate")
	}
}

func TestPersistedGeneratorLoadsFromDisk(t *testing.T) {
	dir := t.TempDir()
	first, _, err := NewPersistedTLSCertGenerator(&countingGenerator{}, dir).GenerateCert()
	if err != nil {
		t.Fatal(err)
	}

	second, _, err := NewPersistedTLSCertGenerator(failingGenerator{}, dir).GenerateCert()
	if err != nil {
		t.Fatalf("expected load from disk, got %v", err)
	}
	if string(first) != string(second) {
		t.Fatal("restarted instance served a different certificate")
	}
}

func TestPersistedGeneratorRegeneratesCorruptFiles(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, certFileName), []byte("junk"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, keyFileName), []byte("junk"), 0o600); err != nil {
		t.Fatal(err)
	}

	base := &countingGenerator{}
	certPEM, keyPEM, err := NewPersistedTLSCertGenerator(base, dir).GenerateCert()
	if err != nil {
		t.Fatal(err)
	}
	if base.calls != 1 {
		t.Fatalf("expected regeneration, base called %d times", base.calls)
	}
	if err := validateKeyPair(certPEM, keyPEM); err != nil {
		t.Fatalf("regenerated pair invalid: %v", err)
	}
}

func TestPersistedGeneratorRenewsExpiringCert(t *testing.T) {
	dir := t.TempDir()
	expiring, expiringKey, err := newTestCert(time.Now().Add(24 * time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, certFileName), expiring, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, keyFileName), expiringKey, 0o600); err != nil {
		t.Fatal(err)
	}

	base := &countingGenerator{}
	certPEM, _, err := NewPersistedTLSCertGenerator(base, dir).GenerateCert()
	if err != nil {
		t.Fatal(err)
	}
	if base.calls != 1 {
		t.Fatalf("expected renewal, base called %d times", base.calls)
	}
	if string(certPEM) == string(expiring) {
		t.Fatal("expiring certificate was reused")
	}
}

func TestPersistedGeneratorCertPEM(t *testing.T) {
	gen := NewPersistedTLSCertGenerator(&countingGenerator{}, t.TempDir())
	if _, ok := gen.CertPEM(); ok {
		t.Fatal("CertPEM reported present before any generation")
	}
	certPEM, _, err := gen.GenerateCert()
	if err != nil {
		t.Fatal(err)
	}
	got, ok := gen.CertPEM()
	if !ok || string(got) != string(certPEM) {
		t.Fatal("CertPEM did not return the generated certificate")
	}
}

func TestPersistedGeneratorPersistFailure(t *testing.T) {
	// A file where the directory should be makes MkdirAll fail.
	blocker := filepath.Join(t.TempDir(), "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	gen := NewPersistedTLSCertGenerator(&countingGenerator{}, filepath.Join(blocker, "tls"))

	if _, _, err := gen.GenerateCert(); err == nil {
		t.Fatal("expected persistence failure")
	}
	if _, ok := gen.CertPEM(); ok {
		t.Fatal("memo must stay empty after a persistence failure")
	}
}

func TestPersistedGeneratorWithoutDir(t *testing.T) {
	base := &countingGenerator{}
	gen := NewPersistedTLSCertGenerator(base, "")
	if _, _, err := gen.GenerateCert(); err != nil {
		t.Fatal(err)
	}
	if _, ok := gen.CertPEM(); !ok {
		t.Fatal("CertPEM missing for in-memory generator")
	}
	if _, _, err := gen.GenerateCert(); err != nil {
		t.Fatal(err)
	}
	if base.calls != 1 {
		t.Fatalf("base called %d times, want 1", base.calls)
	}
}
