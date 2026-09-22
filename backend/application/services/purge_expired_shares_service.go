package services

import (
	"time"

	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

// PurgeExpiredSharesService removes public links whose validity window has
// passed. It is invoked opportunistically (e.g. at startup) to keep the share
// store from accumulating dead links.
type PurgeExpiredSharesService struct {
	shareRepo ports.ShareRepository
	now       func() time.Time
}

func NewPurgeExpiredSharesService(shareRepo ports.ShareRepository) *PurgeExpiredSharesService {
	return &PurgeExpiredSharesService{shareRepo: shareRepo, now: time.Now}
}

// Execute returns the number of expired shares removed.
func (s *PurgeExpiredSharesService) Execute() (int, error) {
	return s.shareRepo.PurgeExpired(s.now())
}
