package services

import (
	"sort"
	"strings"

	domainerrors "github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

// maxDirectoryResults bounds directory search output: this is a call-sheet
// lookup, not an export.
const maxDirectoryResults = 20

// DirectoryService searches accounts by username for calling and sharing.
// It returns usernames only — never roles, quotas, or contact details — and
// requires authentication (wired via apiChain, like the rest of the API).
type DirectoryService struct {
	users ports.UserRepository
}

func NewDirectoryService(users ports.UserRepository) *DirectoryService {
	return &DirectoryService{users: users}
}

// Search returns up to maxDirectoryResults usernames containing query
// (case-insensitive), excluding the requester. Empty queries match nothing
// rather than listing the user base.
func (s *DirectoryService) Search(requester *models.User, query string) ([]string, error) {
	query = strings.TrimSpace(query)
	if query == "" || len(query) > 64 {
		return nil, domainerrors.NewValidationError("q", query, "search query must be 1-64 characters")
	}
	self := ""
	if requester != nil {
		self = strings.ToLower(requester.Username)
	}

	list, err := s.users.ListUsers()
	if err != nil {
		return nil, err
	}
	needle := strings.ToLower(query)
	out := make([]string, 0, maxDirectoryResults)
	for _, u := range list {
		if len(out) >= maxDirectoryResults {
			break
		}
		if self != "" && strings.ToLower(u.Username) == self {
			continue
		}
		if strings.Contains(strings.ToLower(u.Username), needle) {
			out = append(out, u.Username)
		}
	}
	sort.Strings(out)
	return out, nil
}
