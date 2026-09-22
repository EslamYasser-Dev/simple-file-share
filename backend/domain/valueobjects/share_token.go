package valueobjects

import "fmt"

// ShareToken is a validated public-link token. Its structure is checked before
// any repository or filesystem access so crafted garbage never reaches storage.
type ShareToken struct {
	value string
}

// NewShareToken validates a token's shape (URL-safe base64 alphabet, bounded
// length) and fails for anything that cannot be a real token.
func NewShareToken(token string) (ShareToken, error) {
	if len(token) == 0 || len(token) > 64 {
		return ShareToken{}, fmt.Errorf("invalid share token")
	}
	for _, r := range token {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '-' || r == '_':
		default:
			return ShareToken{}, fmt.Errorf("invalid share token")
		}
	}
	return ShareToken{value: token}, nil
}

// String returns the raw token.
func (t ShareToken) String() string {
	if t.value == "" {
		return ""
	}
	return t.value
}
