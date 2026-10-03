package services

import (
	pathpkg "path"

	domainerrors "github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
	"github.com/EslamYasser-Dev/simple-file-share/domain/valueobjects"

	"github.com/EslamYasser-Dev/simple-file-share/application/events"
)

// VisibilityService enforces per-file privacy. The repository is the source
// of truth, but every serving path must call CanView/CanStream — a missing
// record always means private, never public.
type VisibilityService struct {
	fileRepo  ports.FileRepository
	scoper    ports.PathScoper
	visRepo   ports.VisibilityRepository
	follows   ports.FollowRepository
	timeline  ports.TimelineRepository
	downloads *DownloadService
	users     ports.UserRepository
	bus       *events.Bus
}

func NewVisibilityService(
	fileRepo ports.FileRepository,
	scoper ports.PathScoper,
	visRepo ports.VisibilityRepository,
	follows ports.FollowRepository,
	timeline ports.TimelineRepository,
	downloads *DownloadService,
	users ports.UserRepository,
) *VisibilityService {
	return &VisibilityService{
		fileRepo:  fileRepo,
		scoper:    scoper,
		visRepo:   visRepo,
		follows:   follows,
		timeline:  timeline,
		downloads: downloads,
		users:     users,
	}
}

// SetEventBus attaches a live-update bus (nil disables publishing).
func (s *VisibilityService) SetEventBus(bus *events.Bus) { s.bus = bus }

// Set changes the audience of one file the caller owns. Only the owner (as
// resolved by the path scoper) may re-scope; everyone else gets Forbidden.
func (s *VisibilityService) Set(user *models.User, path string, level models.VisibilityLevel, allowStream bool) (*models.FileVisibility, error) {
	if user == nil {
		return nil, &domainerrors.ForbiddenError{Action: "set visibility on", Path: path}
	}
	if !level.Valid() {
		return nil, domainerrors.NewValidationError("level", string(level), "must be private, link, or public")
	}

	fp, err := valueobjects.NewFilePath(path)
	if err != nil {
		return nil, err
	}
	virtual := fp.Relative()
	physical, err := s.scoper.ReadPath(user, virtual)
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

	vis := &models.FileVisibility{
		Owner:       user.Username,
		Path:        virtual,
		Level:       level,
		AllowStream: allowStream,
		UpdatedAt:   nowUTC(),
	}
	if err := s.visRepo.Set(vis); err != nil {
		return nil, err
	}
	// Old feed entries must not outlive a downgrade.
	if s.timeline != nil {
		_ = s.timeline.UpdateVisibility(user.Username, virtual, level)
		_ = s.timeline.Append(&models.TimelineEvent{
			ID:         newTimelineID(),
			Owner:      user.Username,
			Kind:       models.TimelineVisibility,
			Path:       virtual,
			Name:       pathpkg.Base(virtual),
			Visibility: level,
			CreatedAt:  nowUTC(),
		})
	}
	publishEvent(s.bus, events.TypeVisibility, virtual, user)
	return vis, nil
}

// Get returns the stored setting for owner+path (private default when
// unset). Callers enforce their own authZ; the level label alone reveals
// nothing about content.
func (s *VisibilityService) Get(owner, path string) (*models.FileVisibility, error) {
	return s.visRepo.Get(owner, path)
}

// CanView reports whether viewer may see the file (feed entry, metadata).
// Owners always can; private never leaves the owner; link needs a follow
// edge; public needs an authenticated viewer.
func (s *VisibilityService) CanView(viewer *models.User, owner, path string) (bool, error) {
	if viewer != nil && viewer.Username == owner {
		return true, nil
	}
	vis, err := s.visRepo.Get(owner, path)
	if err != nil {
		return false, err
	}
	switch vis.Level {
	case models.VisibilityPublic:
		return viewer != nil, nil
	case models.VisibilityLink:
		if viewer == nil {
			return false, nil
		}
		return s.follows.IsFollowing(viewer.Username, owner)
	default:
		return false, nil
	}
}

// CanStream reports whether viewer may play/view the media bytes. It is
// CanView plus the owner's streaming toggle; owners always stream.
func (s *VisibilityService) CanStream(viewer *models.User, owner, path string) (bool, error) {
	if viewer != nil && viewer.Username == owner {
		return true, nil
	}
	ok, err := s.CanView(viewer, owner, path)
	if err != nil || !ok {
		return ok, err
	}
	vis, err := s.visRepo.Get(owner, path)
	if err != nil {
		return false, err
	}
	return vis.AllowStream, nil
}

// Serve resolves another owner's file into a streamable download after
// re-checking the streaming gate. It replays the owner's read scope (with
// their stored admin flag, mirroring share resolution) so the stored virtual
// path resolves to the same location. Deleted files surface as NotFound.
func (s *VisibilityService) Serve(viewer *models.User, owner, path string) (*models.Download, error) {
	if viewer == nil {
		return nil, &domainerrors.ForbiddenError{Action: "view shared file", Path: path}
	}
	ok, err := s.CanStream(viewer, owner, path)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, &domainerrors.ForbiddenError{Action: "view shared file", Path: path}
	}
	stored, err := s.users.FindByUsername(owner)
	if err != nil {
		return nil, &domainerrors.NotFoundError{Path: path}
	}
	ownerCtx := &models.User{Username: stored.Username, IsAdmin: stored.IsAdmin}
	return s.downloads.Execute(ownerCtx, path)
}
