package errors

type NotFoundError struct {
	Path string
}

func (e *NotFoundError) Error() string {
	return "path not found: " + e.Path
}
