package policy

import (
	"io"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

// MaxIndexTextBytes caps how much of a single file is read for full-text
// indexing. Larger files are truncated at a UTF-8 boundary so memory stays
// bounded per document.
const MaxIndexTextBytes = 256 << 10 // 256 KiB

// indexableExtensions lists file types whose contents are plain text and safe
// to tokenize. Binaries and office formats are skipped.
var indexableExtensions = map[string]bool{
	".txt": true, ".text": true, ".log": true,
	".md": true, ".markdown": true,
	".csv": true, ".tsv": true,
	".json": true, ".jsonl": true,
	".yml": true, ".yaml": true, ".toml": true, ".ini": true, ".cfg": true, ".conf": true,
	".xml": true, ".html": true, ".htm": true, ".css": true,
	".go": true, ".mod": true, ".sum": true,
	".js": true, ".jsx": true, ".ts": true, ".tsx": true, ".mjs": true, ".cjs": true,
	".py": true, ".rb": true, ".rs": true, ".java": true, ".kt": true, ".kts": true,
	".c": true, ".h": true, ".cpp": true, ".hpp": true, ".cc": true, ".cs": true,
	".sh": true, ".bash": true, ".zsh": true, ".ps1": true,
	".sql": true, ".graphql": true, ".gql": true,
	".env": true,
}

// indexableBaseNames covers extensionless text files worth searching.
var indexableBaseNames = map[string]bool{
	"makefile":   true,
	"dockerfile": true,
	"license":    true,
	"readme":     true,
	"procfile":   true,
}

// IsIndexableTextPath reports whether a file path's contents should be fed to
// the full-text index (extension allowlist; walker already excludes .file-share).
func IsIndexableTextPath(path string) bool {
	base := filepath.Base(path)
	ext := strings.ToLower(filepath.Ext(base))
	if ext != "" {
		return indexableExtensions[ext]
	}
	return indexableBaseNames[strings.ToLower(base)]
}

// ReadIndexableText reads at most MaxIndexTextBytes from r and returns the
// extracted text. ok is false when the path is not indexable or the bytes are
// not valid UTF-8 text (binary sniff via NUL).
func ReadIndexableText(path string, r io.Reader) (text string, ok bool) {
	if !IsIndexableTextPath(path) {
		return "", false
	}
	buf := make([]byte, MaxIndexTextBytes+1)
	n, err := io.ReadFull(r, buf)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return "", false
	}
	buf = buf[:n]
	if len(buf) > MaxIndexTextBytes {
		buf = buf[:MaxIndexTextBytes]
		for len(buf) > 0 && !utf8.Valid(buf) {
			buf = buf[:len(buf)-1]
		}
	}
	if looksBinary(buf) {
		return "", false
	}
	if !utf8.Valid(buf) {
		return "", false
	}
	return string(buf), true
}

func looksBinary(b []byte) bool {
	for _, c := range b {
		if c == 0 {
			return true
		}
	}
	return false
}
