package ports

import "time"

// AuditEntry is one security-relevant event in the append-only audit trail.
type AuditEntry struct {
	At     time.Time `json:"at"`
	Action string    `json:"action"`
	Actor  string    `json:"actor"`
	Target string    `json:"target,omitempty"`
	Remote string    `json:"remote,omitempty"`
	Detail string    `json:"detail,omitempty"`
}

// AuditQuery filters the audit trail. Entries come back newest first.
type AuditQuery struct {
	// Action matches an entry's action exactly; empty matches all.
	Action string
	// Actor matches an entry's actor exactly; empty matches all.
	Actor string
	// Limit is the page size; values outside 1..500 are clamped.
	Limit int
	// Cursor is the opaque next cursor from a previous page; empty starts over.
	Cursor string
}

// AuditLog appends and queries security audit events. Implementations must be
// safe for concurrent use. Record must never panic: a full disk or I/O error
// is returned so callers can decide (application services swallow it — an
// audit failure must not fail the business operation).
type AuditLog interface {
	// Record appends one entry to the durable trail.
	Record(entry AuditEntry) error
	// Query returns a page of matching entries newest first plus an opaque
	// next cursor ("" when the page is the last one).
	Query(query AuditQuery) (entries []AuditEntry, nextCursor string, err error)
}
