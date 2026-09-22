package errors

// QuotaExceededError indicates a write was rejected because it would push the
// account over its configured storage quota.
type QuotaExceededError struct {
	Action string
	Path   string
}

func (e *QuotaExceededError) Error() string {
	if e.Path != "" {
		return "storage quota exceeded: cannot " + e.Action + " " + e.Path
	}
	return "storage quota exceeded: cannot " + e.Action
}
