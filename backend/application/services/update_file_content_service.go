package services

import (
	"io"
	"strings"

	"github.com/EslamYasser-Dev/simple-file-share/application/events"
	"github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
	"github.com/EslamYasser-Dev/simple-file-share/domain/valueobjects"
)

// UpdateFileContentService overwrites a file's contents with the supplied text
// (used by the markdown editor). Like uploads, it keeps the search index in
// sync through the port and enforces the same size cap and storage quota.
type UpdateFileContentService struct {
	fileRepo ports.FileRepository
	scoper   ports.PathScoper
	users    ports.UserRepository
	index    ports.FileIndexRepository
	maxBytes int64 // <= 0 means unlimited
	bus      *events.Bus
}

func NewUpdateFileContentService(fileRepo ports.FileRepository, scoper ports.PathScoper) *UpdateFileContentService {
	return &UpdateFileContentService{fileRepo: fileRepo, scoper: scoper}
}

// SetEventBus attaches a live-update bus (nil disables publishing).
func (s *UpdateFileContentService) SetEventBus(bus *events.Bus) { s.bus = bus }

// SetLimits wires the per-request size cap and the quota accounting ports.
// Leaving users/index nil disables quota checks (tests, system views).
func (s *UpdateFileContentService) SetLimits(users ports.UserRepository, index ports.FileIndexRepository, maxBytes int64) {
	s.users = users
	s.index = index
	s.maxBytes = maxBytes
}

func (s *UpdateFileContentService) Execute(user *models.User, path, content string) (int64, error) {
	if s.maxBytes > 0 && int64(len(content)) > s.maxBytes {
		return 0, errors.NewValidationError("content", len(content), "upload exceeds size limit")
	}

	fp, err := valueobjects.NewFilePath(path)
	if err != nil {
		return 0, err
	}

	physical, err := s.scoper.WritePath(user, fp.Relative())
	if err != nil {
		return 0, err
	}

	exists, err := s.fileRepo.FileExists(physical)
	if err != nil {
		return 0, err
	}
	if !exists {
		return 0, &errors.NotFoundError{Path: path}
	}

	isDir, err := s.fileRepo.IsDirectory(physical)
	if err != nil {
		return 0, err
	}
	if isDir {
		return 0, errors.NewValidationError("path", path, "cannot edit a directory")
	}

	if err := s.ensureQuota(user, physical, int64(len(content))); err != nil {
		return 0, err
	}

	reader := io.NopCloser(strings.NewReader(content))
	written, err := s.fileRepo.WriteFile(physical, reader)
	if err == nil {
		publishEvent(s.bus, events.TypeUpdate, path, user)
	}
	return written, err
}

// ensureQuota rejects an edit that would push the account past its storage
// quota, counting only the growth of this file (an overwrite that shrinks or
// stays level is always fine). A nil users/index pair disables the check.
func (s *UpdateFileContentService) ensureQuota(user *models.User, physical string, newSize int64) error {
	if s.users == nil || s.index == nil || user == nil {
		return nil
	}
	quota, err := s.users.GetQuotaBytes(user.Username)
	if err != nil || quota <= 0 {
		// A vanished account record (or unlimited quota) has nothing to enforce.
		return nil
	}
	info, err := s.fileRepo.GetFileInfo(physical)
	if err != nil {
		return err
	}
	_, used, err := s.index.PrefixStats(s.scoper.PrivatePrefix(user.Username))
	if err != nil {
		return err
	}
	growth := newSize - info.Size
	if growth > 0 && used+growth > quota {
		return &errors.QuotaExceededError{Action: "edit", Path: physical}
	}
	return nil
}
