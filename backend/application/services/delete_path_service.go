package services

import (
	"github.com/EslamYasser-Dev/simple-file-share/application/events"
	"github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
	"github.com/EslamYasser-Dev/simple-file-share/domain/valueobjects"
)

// DeletePathService removes a file or directory from the caller's namespace.
type DeletePathService struct {
	fileRepo ports.FileRepository
	scoper   ports.PathScoper
	bus      *events.Bus
}

func NewDeletePathService(fileRepo ports.FileRepository, scoper ports.PathScoper) *DeletePathService {
	return &DeletePathService{fileRepo: fileRepo, scoper: scoper}
}

// SetEventBus attaches a live-update bus (nil disables publishing).
func (s *DeletePathService) SetEventBus(bus *events.Bus) { s.bus = bus }

func (s *DeletePathService) Execute(user *models.User, path string) error {
	fp, err := valueobjects.NewFilePath(path)
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
		return &errors.NotFoundError{Path: path}
	}

	if err := s.fileRepo.DeletePath(physical); err != nil {
		return err
	}
	publishEvent(s.bus, events.TypeDelete, path, user)
	return nil
}
