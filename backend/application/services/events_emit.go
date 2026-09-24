package services

import (
	"time"

	"github.com/EslamYasser-Dev/simple-file-share/application/events"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
)

// publishEvent sends a live-update notification when a bus is attached.
// path is the virtual path the client already knows; user is the actor.
func publishEvent(bus *events.Bus, eventType, path string, user *models.User) {
	publishEventBytes(bus, eventType, path, user, 0)
}

// publishEventBytes is publishEvent with an optional payload size for analytics.
func publishEventBytes(bus *events.Bus, eventType, path string, user *models.User, bytes int64) {
	if bus == nil {
		return
	}
	username := ""
	if user != nil {
		username = user.Username
	}
	bus.Publish(events.Event{
		Type:  eventType,
		Path:  path,
		User:  username,
		Bytes: bytes,
		At:    time.Now().UTC(),
	})
}
