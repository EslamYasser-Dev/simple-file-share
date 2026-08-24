package valueobjects

import (
	"path/filepath"
	"strings"

	"github.com/EslamYasser-Dev/simple-file-share/domain/errors"
)

// FilePath is a validated relative path within the file-share root.
type FilePath struct {
	value string
}

// NewFilePath validates and normalizes a user-supplied path.
func NewFilePath(path string) (FilePath, error) {
	if path == "" {
		return FilePath{}, errors.NewValidationError("path", path, "path cannot be empty")
	}

	normalized := filepath.ToSlash(strings.TrimPrefix(path, "/"))

	if normalized == "" {
		return FilePath{value: "/"}, nil
	}

	// Reject any exact ".." path segment (traversal). This still allows
	// legitimate names that merely contain dots, e.g. "backup..2024.tar".
	for _, segment := range strings.Split(normalized, "/") {
		if segment == ".." {
			return FilePath{}, errors.NewValidationError("path", path, "path traversal detected")
		}
	}

	dangerous := []string{"<", ">", ":", "\"", "|", "?", "*"}
	for _, ch := range dangerous {
		if strings.Contains(normalized, ch) {
			return FilePath{}, errors.NewValidationError("path", path, "contains invalid character")
		}
	}

	return FilePath{value: normalized}, nil
}

// String returns the normalized path (leading slash, forward slashes).
func (p FilePath) String() string {
	if p.value == "" || p.value == "/" {
		return "/"
	}
	return "/" + strings.TrimPrefix(p.value, "/")
}

// Relative returns the path without leading slash, for filesystem joins.
func (p FilePath) Relative() string {
	return strings.TrimPrefix(p.value, "/")
}

// Join appends a child segment to this path.
func (p FilePath) Join(name string) (FilePath, error) {
	if name == "" {
		return FilePath{}, errors.NewValidationError("path", name, "name cannot be empty")
	}
	base := strings.TrimPrefix(p.value, "/")
	combined := base
	if combined == "" {
		combined = name
	} else {
		combined = combined + "/" + name
	}
	return NewFilePath(combined)
}
