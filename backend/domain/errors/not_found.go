package errors

import "errors"

// ErrNotFound is the sentinel returned by repositories when a path does not
// exist, so the application layer never has to inspect OS-level errors.
var ErrNotFound = errors.New("not found")

type NotFoundError struct {
	Path string
}

func (e *NotFoundError) Error() string {
	return "path not found: " + e.Path
}
