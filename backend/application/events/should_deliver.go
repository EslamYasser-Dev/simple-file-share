package events

import "github.com/EslamYasser-Dev/simple-file-share/domain/models"

// ShouldDeliver filters events for a connected subscriber. Nil subscribers
// (auth disabled / system view) receive everything; otherwise a user sees
// their own events plus broadcast events (empty actor).
func ShouldDeliver(e Event, user *models.User) bool {
	if user == nil {
		return true
	}
	if e.User == "" {
		return true
	}
	return e.User == user.Username
}
