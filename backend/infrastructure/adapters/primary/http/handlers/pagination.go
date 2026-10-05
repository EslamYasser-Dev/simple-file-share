package handlers

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/primary/http/dto"
)

// Collection endpoints share one envelope — {"items":[...],"nextCursor":"…"} —
// with an opaque offset cursor. limit defaults to 50 and clamps to [1, 500],
// matching the audit endpoint's bounds.
const (
	defaultPageLimit = 50
	maxPageLimit     = 500
	// maxPageOffset bounds how deep a cursor may reach, so a tampered cursor
	// cannot demand an unbounded scan from the backing index.
	maxPageOffset = 1_000_000
)

type paging struct {
	limit  int
	offset int
}

// parsePaging reads ?limit= and ?cursor=. A malformed limit falls back to the
// default (lenient, like the old behavior); a malformed cursor is an error
// because it is an opaque token clients must round-trip verbatim.
func parsePaging(r *http.Request) (paging, error) {
	p := paging{limit: defaultPageLimit}
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			p.limit = n
		}
	}
	if p.limit > maxPageLimit {
		p.limit = maxPageLimit
	}

	raw := r.URL.Query().Get("cursor")
	if raw == "" {
		return p, nil
	}
	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return p, errors.New("malformed cursor")
	}
	var payload struct {
		Offset int `json:"o"`
	}
	if err := json.Unmarshal(decoded, &payload); err != nil || payload.Offset < 0 {
		return p, errors.New("malformed cursor")
	}
	if payload.Offset > maxPageOffset {
		payload.Offset = maxPageOffset
	}
	p.offset = payload.Offset
	return p, nil
}

// fetchSize is how many leading items a backing service must materialize so
// the window [offset, offset+limit) can be served — plus one lookahead item
// so the envelope can tell "more exist" from an exactly-full last page.
func (p paging) fetchSize() int {
	return p.offset + p.limit + 1
}

// window slices the caller's window out of items and builds the envelope.
// Services hand back the full (or at-least-fetchSize) ordered list; the page
// boundary is decided here so every endpoint stays consistent.
func window[T any](p paging, items []T) dto.Page[T] {
	if p.offset >= len(items) {
		return dto.Page[T]{Items: []T{}}
	}
	end := p.offset + p.limit
	if end > len(items) {
		end = len(items)
	}
	page := dto.Page[T]{Items: items[p.offset:end]}
	if end < len(items) {
		next, _ := json.Marshal(struct {
			Offset int `json:"o"`
		}{Offset: end})
		page.NextCursor = base64.RawURLEncoding.EncodeToString(next)
	}
	return page
}
