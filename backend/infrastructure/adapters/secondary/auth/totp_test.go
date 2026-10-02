package auth

import (
	"strings"
	"testing"
	"time"
)

// RFC 6238 Appendix B test vector secret (ASCII "12345678901234567890").
const rfcSecret = "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"

// Expected 6-digit codes — the RFC's SHA-1 vectors truncated to 6 digits
// (cross-checked against an independent HMAC-SHA1 implementation).
func TestTOTPRFC6238Vectors(t *testing.T) {
	cases := []struct {
		unix int64
		want string
	}{
		{59, "287082"},
		{1111111109, "081804"},
		{1111111111, "050471"},
		{1234567890, "005924"},
		{2000000000, "279037"},
		{20000000000, "353130"},
	}
	for _, c := range cases {
		got, err := TOTPCode(rfcSecret, time.Unix(c.unix, 0).UTC())
		if err != nil {
			t.Fatalf("TOTPCode(%d): %v", c.unix, err)
		}
		if got != c.want {
			t.Errorf("TOTPCode(%d) = %s, want %s", c.unix, got, c.want)
		}
		if !ValidateTOTP(rfcSecret, c.want, 1, time.Unix(c.unix, 0).UTC()) {
			t.Errorf("ValidateTOTP rejected vector at %d", c.unix)
		}
	}
}

func TestGenerateValidateRoundtrip(t *testing.T) {
	secret, err := GenerateTOTPSecret()
	if err != nil {
		t.Fatal(err)
	}
	if len(secret) != 32 {
		t.Fatalf("secret len = %d, want 32 base32 chars", len(secret))
	}
	now := time.Now()
	code, err := TOTPCode(secret, now)
	if err != nil {
		t.Fatal(err)
	}
	if len(code) != 6 {
		t.Fatalf("code = %q, want 6 digits", code)
	}
	if !ValidateTOTP(secret, code, 1, now) {
		t.Fatal("fresh code must validate")
	}
}

func TestValidateWindowAndRejections(t *testing.T) {
	secret, err := GenerateTOTPSecret()
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC().Truncate(time.Second)
	prev, _ := TOTPCode(secret, now.Add(-30*time.Second))
	if !ValidateTOTP(secret, prev, 1, now) {
		t.Fatal("one step back must be inside window=1")
	}
	stale, _ := TOTPCode(secret, now.Add(-90*time.Second))
	if ValidateTOTP(secret, stale, 1, now) {
		t.Fatal("three steps back must be outside window=1")
	}
	for _, bad := range []string{"", "12345", "1234567", "abcdef", "  " + prev[:5] + "0"} {
		if ValidateTOTP(secret, bad, 1, now) {
			t.Errorf("ValidateTOTP accepted %q", bad)
		}
	}
	if ValidateTOTP("not-base32!!", "123456", 1, now) {
		t.Fatal("invalid secret must not validate")
	}
}

func TestProvisioningURI(t *testing.T) {
	uri := TOTPProvisioningURI("alice@example.com", "Simple File Share", "SECRET123")
	for _, part := range []string{
		"otpauth://totp/",
		// PathEscape leaves '@' intact (valid pchar); spaces do escape.
		"Simple%20File%20Share:alice@example.com",
		"secret=SECRET123",
		"algorithm=SHA1",
		"digits=6",
		"period=30",
	} {
		if !strings.Contains(uri, part) {
			t.Errorf("URI %q missing %q", uri, part)
		}
	}
}

func TestBackupCodes(t *testing.T) {
	codes, err := GenerateBackupCodes(10)
	if err != nil {
		t.Fatal(err)
	}
	if len(codes) != 10 {
		t.Fatalf("got %d codes, want 10", len(codes))
	}
	seen := map[string]bool{}
	for _, c := range codes {
		if !LooksLikeBackupCode(c) {
			t.Errorf("code %q does not match backup format", c)
		}
		if seen[c] {
			t.Errorf("duplicate code %q", c)
		}
		seen[c] = true
	}
	for _, bad := range []string{"", "1234-123", "1234-12345", "gggg-gggg", "12345678", "1234_1234"} {
		if LooksLikeBackupCode(bad) {
			t.Errorf("LooksLikeBackupCode(%q) = true", bad)
		}
	}
	if codes, err := GenerateBackupCodes(0); err != nil || codes != nil {
		t.Fatalf("zero codes: %v %v", codes, err)
	}
}
