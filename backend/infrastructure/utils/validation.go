package utils

import (
	"path/filepath"
	"strings"

	"github.com/EslamYasser-Dev/simple-file-share/domain/errors"
)

// ValidatePath ensures the path is safe and doesn't contain path traversal
func ValidatePath(path string) error {
	if path == "" {
		return errors.NewValidationError("path", path, "path cannot be empty")
	}

	// Check for path traversal attempts
	if strings.Contains(path, "..") {
		return errors.NewValidationError("path", path, "path traversal detected")
	}

	// Check for absolute paths
	if filepath.IsAbs(path) {
		return errors.NewValidationError("path", path, "absolute paths not allowed")
	}

	// Check for dangerous characters
	dangerousChars := []string{"<", ">", ":", "\"", "|", "?", "*"}
	for _, char := range dangerousChars {
		if strings.Contains(path, char) {
			return errors.NewValidationError("path", path, "contains dangerous character: "+char)
		}
	}

	return nil
}

// SanitizeFilename removes dangerous characters from filename
func SanitizeFilename(filename string) string {
	// Remove path separators
	filename = strings.ReplaceAll(filename, "/", "_")
	filename = strings.ReplaceAll(filename, "\\", "_")
	
	// Remove other dangerous characters
	dangerousChars := []string{"<", ">", ":", "\"", "|", "?", "*"}
	for _, char := range dangerousChars {
		filename = strings.ReplaceAll(filename, char, "_")
	}
	
	// Limit length
	if len(filename) > 255 {
		filename = filename[:255]
	}
	
	return filename
}
