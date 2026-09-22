package errors

// ForbiddenError indicates the authenticated user is not allowed to perform
// the requested action on the given path.
type ForbiddenError struct {
	Action string
	Path   string
}

func (e *ForbiddenError) Error() string {
	if e.Path != "" {
		return "forbidden: cannot " + e.Action + " " + e.Path
	}
	return "forbidden: cannot " + e.Action
}
