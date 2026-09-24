package services

import (
	"time"

	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

// AnalyticsService exposes admin usage reports backed by AnalyticsStore.
// Authorization reuses users.read so the same gate as the account list applies.
type AnalyticsService struct {
	store ports.AnalyticsStore
	roles *RoleCatalog
}

func NewAnalyticsService(store ports.AnalyticsStore, roles *RoleCatalog) *AnalyticsService {
	return &AnalyticsService{store: store, roles: roles}
}

func (s *AnalyticsService) authorize(actor *models.User) error {
	if actor == nil {
		return nil // auth disabled — system view
	}
	if actor.IsSystemView() || actor.HasPermission(models.PermUsersRead, s.roles) {
		return nil
	}
	return forbiddenErr("view analytics")
}

// window resolves days (1–365, default 30) into [since, now).
func window(days int) (time.Time, time.Time) {
	if days <= 0 {
		days = 30
	}
	if days > 365 {
		days = 365
	}
	now := time.Now().UTC()
	return now.AddDate(0, 0, -days), now
}

func (s *AnalyticsService) Overview(actor *models.User, days int) (*ports.AnalyticsOverview, error) {
	if err := s.authorize(actor); err != nil {
		return nil, err
	}
	since, until := window(days)
	return s.store.Overview(since, until)
}

func (s *AnalyticsService) Timeline(actor *models.User, days int) ([]ports.AnalyticsBucket, error) {
	if err := s.authorize(actor); err != nil {
		return nil, err
	}
	since, until := window(days)
	return s.store.Timeline(since, until)
}

func (s *AnalyticsService) TopFiles(actor *models.User, days, limit int) ([]ports.AnalyticsTopFile, error) {
	if err := s.authorize(actor); err != nil {
		return nil, err
	}
	since, until := window(days)
	return s.store.TopFiles(since, until, limit)
}
