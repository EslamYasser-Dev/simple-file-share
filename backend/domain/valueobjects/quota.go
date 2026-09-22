package valueobjects

import (
	"math"
	"strconv"
	"strings"
)

// StorageQuota is an account's storage allowance in bytes. A value of 0 means
// unlimited. The type is an int64 so a quota can travel as a plain JSON number.
type StorageQuota int64

// Bytes returns the quota as a plain byte count (0 = unlimited).
func (q StorageQuota) Bytes() int64 { return int64(q) }

// IsUnlimited reports whether the quota allows unbounded storage.
func (q StorageQuota) IsUnlimited() bool { return q <= 0 }

// ParseStorageQuota parses a quota from a byte count or a human size such as
// "2GB", "512MB", "10TB". "0" and "unlimited" mean no limit. Negative values
// and anything that overflows int64 are rejected.
func ParseStorageQuota(raw string) (StorageQuota, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return 0, &quotaParseError{value: raw, reason: "quota must not be empty"}
	}

	lower := strings.ToLower(s)
	if lower == "unlimited" || lower == "0" || lower == "0b" {
		return 0, nil
	}

	if n, err := strconv.ParseInt(s, 10, 64); err == nil {
		if n <= 0 {
			return 0, &quotaParseError{value: raw, reason: "quota must be a positive number of bytes or \"unlimited\""}
		}
		return StorageQuota(n), nil
	}

	units := []struct {
		suffix string
		mul    int64
	}{
		{"TB", 1 << 40}, {"GB", 1 << 30}, {"MB", 1 << 20},
		{"KB", 1 << 10}, {"KIB", 1 << 10}, {"B", 1},
	}
	upper := strings.ToUpper(s)
	for _, u := range units {
		if !strings.HasSuffix(upper, u.suffix) {
			continue
		}
		num, err := strconv.ParseInt(strings.TrimSpace(strings.TrimSuffix(upper, u.suffix)), 10, 64)
		if err != nil || num <= 0 {
			return 0, &quotaParseError{value: raw, reason: "invalid size"}
		}
		if num > math.MaxInt64/u.mul {
			return 0, &quotaParseError{value: raw, reason: "size overflows the supported range"}
		}
		return StorageQuota(num * u.mul), nil
	}

	return 0, &quotaParseError{value: raw, reason: "expected bytes or a size like \"500MB\", \"2GB\""}
}

type quotaParseError struct {
	value  string
	reason string
}

func (e *quotaParseError) Error() string {
	return "invalid quota " + strconv.Quote(e.value) + ": " + e.reason
}
