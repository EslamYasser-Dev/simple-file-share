package services

import (
	"crypto/rand"
	"encoding/base64"
	"path"
	"strings"
	"time"

	domainerrors "github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"

	"github.com/EslamYasser-Dev/simple-file-share/application/events"
)

// Feed page bounds: small enough for phones, large enough to survive
// visibility filtering without extra round trips.
const (
	defaultFeedLimit = 20
	maxFeedLimit     = 50
)

func nowUTC() time.Time { return time.Now().UTC() }

// TimelineService records upload/share/visibility events and serves the
// follower feed. The repository only stores; every read re-applies the
// audience rules, so a visibility downgrade hides old entries immediately.
type TimelineService struct {
	timeline ports.TimelineRepository
	follows  ports.FollowRepository
	bus      *events.Bus
	now      func() time.Time
}

func NewTimelineService(timeline ports.TimelineRepository, follows ports.FollowRepository) *TimelineService {
	return &TimelineService{timeline: timeline, follows: follows, now: time.Now}
}

// SetEventBus attaches a live-update bus (nil disables publishing).
func (s *TimelineService) SetEventBus(bus *events.Bus) { s.bus = bus }

// Record appends one feed entry for the owner's file. New uploads default to
// private visibility; re-scoping happens through VisibilityService, which
// refreshes the snapshot on every entry of that file.
func (s *TimelineService) Record(kind string, user *models.User, virtualPath string, size int64, visibility models.VisibilityLevel) (*models.TimelineEvent, error) {
	if user == nil {
		return nil, &domainerrors.ForbiddenError{Action: "record timeline event for", Path: virtualPath}
	}
	switch kind {
	case models.TimelineUpload, models.TimelineShare, models.TimelineVisibility:
	default:
		return nil, domainerrors.NewValidationError("kind", kind, "must be upload, share, or visibility")
	}
	if !visibility.Valid() {
		visibility = models.VisibilityPrivate
	}
	event := &models.TimelineEvent{
		ID:         newTimelineID(),
		Owner:      user.Username,
		Kind:       kind,
		Path:       virtualPath,
		Name:       path.Base(strings.TrimSuffix(virtualPath, "/")),
		Size:       size,
		Visibility: visibility,
		CreatedAt:  s.now().UTC(),
	}
	if err := s.timeline.Append(event); err != nil {
		return nil, err
	}
	return event, nil
}

// Feed returns the viewer's timeline newest-first: their own entries, entries
// from followed owners scoped link/public, and public entries. The cursor is
// the last seen event ID (empty starts at newest).
func (s *TimelineService) Feed(viewer *models.User, after string, limit int) ([]*models.TimelineEvent, error) {
	if viewer == nil {
		return nil, &domainerrors.ForbiddenError{Action: "read timeline feed"}
	}
	if limit <= 0 {
		limit = defaultFeedLimit
	}
	if limit > maxFeedLimit {
		limit = maxFeedLimit
	}

	out := make([]*models.TimelineEvent, 0, limit)
	cursor := after
	for len(out) < limit {
		// Over-fetch: audience filtering can drop entries, and the store is
		// a small JSON document, so a bounded scan is cheap.
		chunk, err := s.timeline.List(cursor, maxFeedLimit)
		if err != nil {
			return nil, err
		}
		if len(chunk) == 0 {
			break
		}
		for _, e := range chunk {
			cursor = e.ID
			if !s.visibleTo(viewer, e) {
				continue
			}
			out = append(out, e)
			if len(out) >= limit {
				break
			}
		}
		if len(chunk) < maxFeedLimit {
			break
		}
	}
	return out, nil
}

func (s *TimelineService) visibleTo(viewer *models.User, e *models.TimelineEvent) bool {
	if e.Owner == viewer.Username {
		return true
	}
	switch e.Visibility {
	case models.VisibilityPublic:
		return true
	case models.VisibilityLink:
		ok, err := s.follows.IsFollowing(viewer.Username, e.Owner)
		return err == nil && ok
	default:
		return false
	}
}

func newTimelineID() string {
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return base64.RawURLEncoding.EncodeToString([]byte(time.Now().UTC().Format(time.RFC3339Nano)))
	}
	return base64.RawURLEncoding.EncodeToString(raw)
}
