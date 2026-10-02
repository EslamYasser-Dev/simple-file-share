// TOTP implements RFC 6238 TOTP (SHA-1, 6 digits, 30s step) with stdlib
// crypto only, plus the provisioning URI and backup-code helpers used by the
// two-factor enrollment flow.
package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"net/url"
	"strings"
	"time"
)

const (
	totpPeriodSeconds = 30
	totpDigits        = 6
	totpSecretBytes   = 20 // RFC 4226 recommended 160-bit secret
)

var base32NoPad = base32.StdEncoding.WithPadding(base32.NoPadding)

// GenerateTOTPSecret returns a new base32 (no padding) TOTP secret.
func GenerateTOTPSecret() (string, error) {
	buf := make([]byte, totpSecretBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate totp secret: %w", err)
	}
	return base32NoPad.EncodeToString(buf), nil
}

// TOTPCode computes the current 6-digit code for secret at time t.
func TOTPCode(secret string, t time.Time) (string, error) {
	key, err := base32NoPad.DecodeString(strings.ToUpper(strings.TrimSpace(secret)))
	if err != nil {
		return "", fmt.Errorf("decode totp secret: %w", err)
	}
	counter := uint64(t.Unix() / totpPeriodSeconds)
	var msg [8]byte
	binary.BigEndian.PutUint64(msg[:], counter)
	mac := hmac.New(sha1.New, key)
	mac.Write(msg[:])
	sum := mac.Sum(nil)
	// RFC 4226 dynamic truncation.
	offset := sum[len(sum)-1] & 0x0f
	code := (binary.BigEndian.Uint32(sum[offset:offset+4]) & 0x7fffffff) % 1000000
	return fmt.Sprintf("%0*d", totpDigits, code), nil
}

// ValidateTOTP reports whether code matches the secret within ±window steps
// of now (window 1 tolerates ±30s of clock skew).
func ValidateTOTP(secret, code string, window int, now time.Time) bool {
	code = strings.TrimSpace(code)
	if len(code) != totpDigits {
		return false
	}
	if window < 0 {
		window = 0
	}
	match := 0
	for i := -window; i <= window; i++ {
		at := now.Add(time.Duration(i) * totpPeriodSeconds * time.Second)
		want, err := TOTPCode(secret, at)
		if err != nil {
			return false
		}
		// Constant-time compare; accumulate so the loop always runs.
		if subtle.ConstantTimeCompare([]byte(code), []byte(want)) == 1 {
			match = 1
		}
	}
	return match == 1
}

// TOTPProvisioningURI builds the otpauth:// link accepted by authenticator
// apps (Google Authenticator, 1Password, Aegis, ...).
func TOTPProvisioningURI(account, issuer, secret string) string {
	v := url.Values{}
	v.Set("secret", strings.TrimSpace(secret))
	v.Set("issuer", issuer)
	v.Set("algorithm", "SHA1")
	v.Set("digits", fmt.Sprintf("%d", totpDigits))
	v.Set("period", fmt.Sprintf("%d", totpPeriodSeconds))
	return fmt.Sprintf("otpauth://totp/%s:%s?%s",
		url.PathEscape(issuer), url.PathEscape(account), v.Encode())
}

// GenerateBackupCodes returns n single-use recovery codes of the form
// "xxxx-xxxx" (8 hex chars). They are stored hashed and only shown once.
func GenerateBackupCodes(n int) ([]string, error) {
	if n <= 0 {
		return nil, nil
	}
	codes := make([]string, 0, n)
	for i := 0; i < n; i++ {
		buf := make([]byte, 4)
		if _, err := rand.Read(buf); err != nil {
			return nil, fmt.Errorf("generate backup code: %w", err)
		}
		hexed := fmt.Sprintf("%x", buf)
		codes = append(codes, hexed[:4]+"-"+hexed[4:])
	}
	return codes, nil
}

// LooksLikeBackupCode reports whether an entered second factor has the
// backup-code shape (so login can skip TOTP validation for it).
func LooksLikeBackupCode(code string) bool {
	code = strings.TrimSpace(code)
	if len(code) != 9 || code[4] != '-' {
		return false
	}
	for i, r := range code {
		if i == 4 {
			continue
		}
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') && (r < 'A' || r > 'F') {
			return false
		}
	}
	return true
}
