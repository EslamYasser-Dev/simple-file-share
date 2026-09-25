package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/EslamYasser-Dev/simple-file-share/application/events"
)

const sseHeartbeat = 15 * time.Second

// EventsHandler streams application events as Server-Sent Events. Clients
// authenticate through the shared API middleware (Bearer or HttpOnly cookie);
// EventSource cannot set headers, so the cookie path is the browser one.
type EventsHandler struct {
	bus *events.Bus
}

func NewEventsHandler(bus *events.Bus) *EventsHandler {
	return &EventsHandler{bus: bus}
}

func (h *EventsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if h.bus == nil {
		http.Error(w, "events unavailable", http.StatusServiceUnavailable)
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	subscriber := currentUser(r)
	ch, stop := h.bus.Subscribe(r.Context())
	defer stop()

	heartbeat := time.NewTicker(sseHeartbeat)
	defer heartbeat.Stop()

	// Initial comment keeps proxies and the browser from treating silence as a
	// stalled connection before the first real event.
	_, _ = fmt.Fprint(w, ": connected\n\n")
	flusher.Flush()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-heartbeat.C:
			if _, err := fmt.Fprint(w, ": ping\n\n"); err != nil {
				return
			}
			flusher.Flush()
		case e, open := <-ch:
			if !open {
				return
			}
			if !events.ShouldDeliver(e, subscriber) {
				continue
			}
			payload, err := json.Marshal(e)
			if err != nil {
				continue
			}
			if _, err := fmt.Fprintf(w, "data: %s\n\n", payload); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}
