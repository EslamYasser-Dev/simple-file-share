package valueobjects

import "testing"

func TestParseStorageQuota(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    int64
		wantErr bool
	}{
		{"empty", "", 0, true},
		{"spaces only", "   ", 0, true},
		{"unlimited word", "unlimited", 0, false},
		{"unlimited upper", "UNLIMITED", 0, false},
		{"zero", "0", 0, false},
		{"plain bytes", "1048576", 1048576, false},
		{"kb", "500KB", 500 << 10, false},
		{"kib lowercase", "2kib", 2 << 10, false},
		{"mb", "512MB", 512 << 20, false},
		{"gb", "2GB", 2 << 30, false},
		{"tb", "1TB", 1 << 40, false},
		{"mb with spaces", " 5 mb ", 5 << 20, false},
		{"negative", "-1", 0, true},
		{"negative size", "-5GB", 0, true},
		{"garbage", "lots", 0, true},
		{"zero suffix", "0GB", 0, true},
		{"overflow", "999999999999999999999999GB", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseStorageQuota(tt.raw)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseStorageQuota(%q) = %d, want error", tt.raw, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseStorageQuota(%q) unexpected error: %v", tt.raw, err)
			}
			if got.Bytes() != tt.want {
				t.Fatalf("ParseStorageQuota(%q) = %d, want %d", tt.raw, got, tt.want)
			}
			if got.IsUnlimited() != (tt.want == 0) {
				t.Fatalf("IsUnlimited() = %v, want %v for %d", got.IsUnlimited(), tt.want == 0, tt.want)
			}
		})
	}
}
