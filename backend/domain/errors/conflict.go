package errors

// IsDirectoryError signals that a single-file download was requested for a
// directory. Handlers map it to 409 Conflict.
type IsDirectoryError struct {
	Path string
}

func (e *IsDirectoryError) Error() string {
	return "path is a directory: " + e.Path
}

// NotDirectoryError signals that a directory (zip) download was requested for a
// regular file. Handlers map it to 400 Bad Request.
type NotDirectoryError struct {
	Path string
}

func (e *NotDirectoryError) Error() string {
	return "path is not a directory: " + e.Path
}
