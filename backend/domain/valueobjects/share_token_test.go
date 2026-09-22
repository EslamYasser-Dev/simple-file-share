package valueobjects

import (
	"strings"
	"testing"
)

func TestNewShareTokenAcceptsURLSafeTokens(t *testing.T) {
	tok, err := NewShareToken("abc-ABC_123")
	if err != nil {
		t.Fatalf("valid token rejected: %v", err)
	}
	if tok.String() != "abc-ABC_123" {
		t.Errorf("token = %q, want abc-ABC_123", tok.String())
	}
}

func TestNewShareTokenRejectsInvalidTokens(t *testing.T) {
	cases := map[string]string{
		"empty":    "",
		"too long": strings.Repeat("a", 65),
		"padded":   "abc==",
		"spaces":   "abc def",
		"slash":    "ab/cd",
		"newline":  "ab\ncd",
		"unicode":  "ab€cd",
		"dot":      "ab.cd",
		"colon":    "ab:cd",
	}
	for name, token := range cases {
		if _, err := NewShareToken(token); err == nil {
			t.Errorf("%s: token %q accepted, want error", name, token)
		}
	}
}
