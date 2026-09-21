package auth

import (
	"strings"
	"testing"
)

func TestPBKDF2HasherRoundTrip(t *testing.T) {
	h := NewPBKDF2Hasher()

	hash, err := h.Hash("correct horse battery staple")
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	if !strings.HasPrefix(hash, "pbkdf2-sha256$") {
		t.Errorf("hash = %q, want pbkdf2-sha256 prefix", hash)
	}
	if !h.Verify("correct horse battery staple", hash) {
		t.Error("Verify returned false for the correct password")
	}
	if h.Verify("wrong password", hash) {
		t.Error("Verify returned true for a wrong password")
	}
}

func TestPBKDF2HasherUsesRandomSalt(t *testing.T) {
	h := NewPBKDF2Hasher()

	a, err := h.Hash("same-password")
	if err != nil {
		t.Fatal(err)
	}
	b, err := h.Hash("same-password")
	if err != nil {
		t.Fatal(err)
	}
	if a == b {
		t.Error("two hashes of the same password are identical; salt is not random")
	}
}

func TestPBKDF2HasherRejectsMalformedHash(t *testing.T) {
	h := NewPBKDF2Hasher()

	for _, bad := range []string{"", "not-a-hash", "pbkdf2-sha256$abc$zz$zz", "pbkdf2-sha256$0$00$00"} {
		if h.Verify("whatever", bad) {
			t.Errorf("Verify accepted malformed hash %q", bad)
		}
	}
}
