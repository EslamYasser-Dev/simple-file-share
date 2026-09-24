package services

import (
	"errors"

	"github.com/EslamYasser-Dev/simple-file-share/application/events"
	domainerrors "github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/policy"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
	"github.com/EslamYasser-Dev/simple-file-share/domain/valueobjects"
)

type ListVersionsService struct {
	fileRepo    ports.FileRepository
	versionRepo ports.VersionRepository
	scoper      ports.PathScoper
}

func NewListVersionsService(fileRepo ports.FileRepository, versionRepo ports.VersionRepository, scoper ports.PathScoper) *ListVersionsService {
	return &ListVersionsService{fileRepo: fileRepo, versionRepo: versionRepo, scoper: scoper}
}

func (s *ListVersionsService) Execute(user *models.User, path string) ([]*models.FileInfo, error) {
	fp, err := valueobjects.NewFilePath(requestPath(path))
	if err != nil {
		return nil, err
	}

	physical, err := s.scoper.ReadPath(user, fp.Relative())
	if err != nil {
		return nil, err
	}

	exists, err := s.fileRepo.FileExists(physical)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, &domainerrors.NotFoundError{Path: path}
	}

	isDir, err := s.fileRepo.IsDirectory(physical)
	if err != nil {
		return nil, err
	}
	if isDir {
		return nil, nil
	}

	versions, err := s.versionRepo.ListVersions(physical)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return nil, &domainerrors.NotFoundError{Path: path}
		}
		return nil, err
	}
	for _, v := range versions {
		v.Path = s.scoper.PhysicalToVirtual(user, v.Path)
	}
	return versions, nil
}

type DownloadVersionService struct {
	fileRepo    ports.FileRepository
	versionRepo ports.VersionRepository
	scoper      ports.PathScoper
	policy      policy.ContentDispositionPolicy
}

func NewDownloadVersionService(fileRepo ports.FileRepository, versionRepo ports.VersionRepository, scoper ports.PathScoper) *DownloadVersionService {
	return &DownloadVersionService{fileRepo: fileRepo, versionRepo: versionRepo, scoper: scoper}
}

func (s *DownloadVersionService) Execute(user *models.User, path string, n int) (*models.Download, error) {
	if n < 1 {
		return nil, domainerrors.NewValidationError("version", n, "version must be positive")
	}

	fp, err := valueobjects.NewFilePath(requestPath(path))
	if err != nil {
		return nil, err
	}

	physical, err := s.scoper.ReadPath(user, fp.Relative())
	if err != nil {
		return nil, err
	}

	exists, err := s.fileRepo.FileExists(physical)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, &domainerrors.NotFoundError{Path: path}
	}

	isDir, err := s.fileRepo.IsDirectory(physical)
	if err != nil {
		return nil, err
	}
	if isDir {
		return nil, &domainerrors.IsDirectoryError{Path: path}
	}

	stream, filename, err := s.versionRepo.ServeVersion(physical, n)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return nil, &domainerrors.NotFoundError{Path: path}
		}
		return nil, err
	}
	contentType := s.policy.ContentTypeFor(filename)
	return &models.Download{
		Stream:      stream,
		Filename:    filename,
		ContentType: contentType,
		Inline:      s.policy.IsInlineContentType(contentType),
	}, nil
}

type RestoreVersionService struct {
	fileRepo    ports.FileRepository
	versionRepo ports.VersionRepository
	scoper      ports.PathScoper
	bus         *events.Bus
}

func NewRestoreVersionService(fileRepo ports.FileRepository, versionRepo ports.VersionRepository, scoper ports.PathScoper) *RestoreVersionService {
	return &RestoreVersionService{fileRepo: fileRepo, versionRepo: versionRepo, scoper: scoper}
}

// SetEventBus attaches a live-update bus (nil disables publishing).
func (s *RestoreVersionService) SetEventBus(bus *events.Bus) { s.bus = bus }

func (s *RestoreVersionService) Execute(user *models.User, path string, n int) error {
	if n < 1 {
		return domainerrors.NewValidationError("version", n, "version must be positive")
	}

	fp, err := valueobjects.NewFilePath(requestPath(path))
	if err != nil {
		return err
	}

	physical, err := s.scoper.WritePath(user, fp.Relative())
	if err != nil {
		return err
	}

	exists, err := s.fileRepo.FileExists(physical)
	if err != nil {
		return err
	}
	if !exists {
		return &domainerrors.NotFoundError{Path: path}
	}

	isDir, err := s.fileRepo.IsDirectory(physical)
	if err != nil {
		return err
	}
	if isDir {
		return &domainerrors.IsDirectoryError{Path: path}
	}

	if err := s.versionRepo.RestoreVersion(physical, n); err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return &domainerrors.NotFoundError{Path: path}
		}
		return err
	}
	// Restore bypasses FileRepository.WriteFile; refresh metadata + full-text
	// so search does not keep serving the pre-restore content.
	if syncer, ok := s.fileRepo.(interface{ SyncPath(string) }); ok {
		syncer.SyncPath(physical)
	}
	publishEvent(s.bus, events.TypeRestore, path, user)
	return nil
}
