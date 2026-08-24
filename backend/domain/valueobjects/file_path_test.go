package valueobjects

import "testing"

func TestNewFilePath(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"root", "/", false},
		{"simple", "docs/readme.txt", false},
		{"traversal", "../../etc/passwd", true},
		{"traversal middle", "docs/../../etc/passwd", true},
		{"traversal suffix", "docs/..", true},
		{"dots in name allowed", "backup..2024.tar", false},
		{"dot segment", "docs/./file.txt", false},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewFilePath(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewFilePath(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestFilePathString(t *testing.T) {
	fp, err := NewFilePath("docs/file.txt")
	if err != nil {
		t.Fatal(err)
	}
	if got := fp.String(); got != "/docs/file.txt" {
		t.Fatalf("String() = %q, want /docs/file.txt", got)
	}
}
