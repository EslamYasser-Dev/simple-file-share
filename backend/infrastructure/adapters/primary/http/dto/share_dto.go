package dto

import (
	"path/filepath"
	"time"

	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
)

// CreateShareRequest is the body of POST /api/shares.
type CreateShareRequest struct {
	Path string `json:"path"`
	// ExpiresInSeconds is the link validity in seconds. 0 means no expiry.
	ExpiresInSeconds int64 `json:"expiresInSeconds"`
}

// ShareItem is a share link as seen by the API. ExpiresAt is empty when the
// link never expires.
type ShareItem struct {
	Token     string `json:"token"`
	Path      string `json:"path"`
	Name      string `json:"name"`
	Owner     string `json:"owner"`
	CreatedAt string `json:"createdAt"`
	ExpiresAt string `json:"expiresAt"`
}

func FromShare(s *models.Share) ShareItem {
	created := ""
	if !s.CreatedAt.IsZero() {
		created = s.CreatedAt.UTC().Format(time.RFC3339)
	}
	expires := ""
	if !s.ExpiresAt.IsZero() {
		expires = s.ExpiresAt.UTC().Format(time.RFC3339)
	}
	return ShareItem{
		Token:     s.Token,
		Path:      s.Path,
		Name:      filepath.Base(s.Path),
		Owner:     s.Owner,
		CreatedAt: created,
		ExpiresAt: expires,
	}
}

func FromShares(shares []*models.Share) []ShareItem {
	items := make([]ShareItem, 0, len(shares))
	for _, s := range shares {
		items = append(items, FromShare(s))
	}
	return items
}
