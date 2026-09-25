package tls

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

const (
	certFileName = "cert.pem"
	keyFileName  = "key.pem"
	// renewBefore forces a new key pair once the stored certificate is within
	// this window of NotAfter, so a healthy server never starts serving an
	// already-expired certificate.
	renewBefore = 7 * 24 * time.Hour
)

// PersistedTLSCertGenerator memoizes a single key pair for the whole process
// and persists it under dir. Every listener (HTTP and gRPC) therefore serves
// the same certificate, and clients that pinned it keep trusting it across
// restarts instead of breaking whenever the process re-instantiates.
type PersistedTLSCertGenerator struct {
	base ports.TLSCertGenerator
	dir  string

	mu      sync.Mutex
	certPEM []byte
	keyPEM  []byte
}

// NewPersistedTLSCertGenerator wraps base; dir == "" disables persistence
// (memoization only).
func NewPersistedTLSCertGenerator(base ports.TLSCertGenerator, dir string) *PersistedTLSCertGenerator {
	return &PersistedTLSCertGenerator{base: base, dir: dir}
}

// GenerateCert returns the process-wide key pair, loading it from disk when
// possible and generating + persisting it otherwise.
func (g *PersistedTLSCertGenerator) GenerateCert() ([]byte, []byte, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.certPEM != nil {
		return g.certPEM, g.keyPEM, nil
	}
	if g.dir != "" {
		if certPEM, keyPEM, err := g.load(); err == nil {
			g.certPEM, g.keyPEM = certPEM, keyPEM
			return certPEM, keyPEM, nil
		}
	}

	certPEM, keyPEM, err := g.base.GenerateCert()
	if err != nil {
		return nil, nil, err
	}
	if g.dir != "" {
		if err := g.persist(certPEM, keyPEM); err != nil {
			return nil, nil, err
		}
	}
	g.certPEM, g.keyPEM = certPEM, keyPEM
	return certPEM, keyPEM, nil
}

// CertPEM exposes the served certificate for the public pin endpoint. ok is
// false until a key pair exists (TLS disabled everywhere).
func (g *PersistedTLSCertGenerator) CertPEM() ([]byte, bool) {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.certPEM, g.certPEM != nil
}

func (g *PersistedTLSCertGenerator) load() ([]byte, []byte, error) {
	certPEM, err := os.ReadFile(filepath.Join(g.dir, certFileName))
	if err != nil {
		return nil, nil, err
	}
	keyPEM, err := os.ReadFile(filepath.Join(g.dir, keyFileName))
	if err != nil {
		return nil, nil, err
	}
	if err := validateKeyPair(certPEM, keyPEM); err != nil {
		return nil, nil, err
	}
	return certPEM, keyPEM, nil
}

func (g *PersistedTLSCertGenerator) persist(certPEM, keyPEM []byte) error {
	if err := os.MkdirAll(g.dir, 0o700); err != nil {
		return fmt.Errorf("create tls dir: %w", err)
	}
	// Key first, cert last: validateKeyPair rejects any partial write, so a
	// crash between the two writes regenerates instead of serving a mismatch.
	if err := os.WriteFile(filepath.Join(g.dir, keyFileName), keyPEM, 0o600); err != nil {
		return fmt.Errorf("write tls key: %w", err)
	}
	if err := os.WriteFile(filepath.Join(g.dir, certFileName), certPEM, 0o600); err != nil {
		return fmt.Errorf("write tls cert: %w", err)
	}
	return nil
}

// validateKeyPair parses both PEM blocks, checks the pair matches, and checks
// the certificate is currently usable (and not inside the renewal window).
func validateKeyPair(certPEM, keyPEM []byte) error {
	if _, err := tls.X509KeyPair(certPEM, keyPEM); err != nil {
		return fmt.Errorf("key pair mismatch: %w", err)
	}
	block, _ := pem.Decode(certPEM)
	if block == nil || block.Type != "CERTIFICATE" {
		return errors.New("no certificate PEM block")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return fmt.Errorf("parse certificate: %w", err)
	}
	now := time.Now()
	if now.Before(cert.NotBefore) {
		return errors.New("certificate not yet valid")
	}
	if now.Add(renewBefore).After(cert.NotAfter) {
		return errors.New("certificate expired or expiring soon")
	}
	return nil
}

var _ ports.TLSCertGenerator = (*PersistedTLSCertGenerator)(nil)
