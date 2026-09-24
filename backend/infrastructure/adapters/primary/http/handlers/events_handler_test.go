package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/EslamYasser-Dev/simple-file-share/application/events"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/primary/authctx"
)

// lockedRecorder wraps httptest.ResponseRecorder so the test can safely read
// the body while the handler goroutine is still writing the SSE stream.
type lockedRecorder struct {
	mu  sync.Mutex
	rec *httptest.ResponseRecorder
}

func newLockedRecorder() *lockedRecorder {
	return &lockedRecorder{rec: httptest.NewRecorder()}
}

func (l *lockedRecorder) Header() http.Header { return l.rec.Header() }

func (l *lockedRecorder) Write(b []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.rec.Write(b)
}

func (l *lockedRecorder) WriteHeader(code int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.rec.WriteHeader(code)
}

func (l *lockedRecorder) Flush() {}

func (l *lockedRecorder) body() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.rec.Body.String()
}

func (l *lockedRecorder) headerValue(k string) string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.rec.Header().Get(k)
}

func waitFor(t *testing.T, timeout time.Duration, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("condition not met before timeout")
}

func TestEventsHandlerRejectsNonGET(t *testing.T) {
	h := NewEventsHandler(events.NewBus())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/events", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", rec.Code)
	}
}

func TestEventsHandlerUnavailableWithoutBus(t *testing.T) {
	h := NewEventsHandler(nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/events", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", rec.Code)
	}
}

func TestEventsHandlerStreamsPublishedEvent(t *testing.T) {
	bus := events.NewBus()
	h := NewEventsHandler(bus)
	rec := newLockedRecorder()

	ctx, cancel := context.WithCancel(context.Background())
	req := httptest.NewRequest(http.MethodGet, "/api/events", nil).WithContext(ctx)

	done := make(chan struct{})
	go func() {
		defer close(done)
		h.ServeHTTP(rec, req)
	}()

	waitFor(t, time.Second, func() bool {
		return strings.Contains(rec.body(), ": connected")
	})
	bus.Publish(events.Event{Type: events.TypeUpload, Path: "a.txt", User: "alice"})
	waitFor(t, time.Second, func() bool {
		return strings.Contains(rec.body(), `"type":"upload"`)
	})
	cancel()
	<-done

	if rec.headerValue("Content-Type") != "text/event-stream" {
		t.Fatalf("content-type = %q", rec.headerValue("Content-Type"))
	}
	body := rec.body()
	if !strings.Contains(body, ": connected") {
		t.Fatalf("missing connected comment: %q", body)
	}
	if !strings.Contains(body, `"type":"upload"`) {
		t.Fatalf("missing upload event: %q", body)
	}
}

func TestEventsHandlerFiltersOtherUsers(t *testing.T) {
	bus := events.NewBus()
	h := NewEventsHandler(bus)
	rec := newLockedRecorder()

	ctx, cancel := context.WithCancel(context.Background())
	user := &models.User{Username: "bob"}
	req := httptest.NewRequest(http.MethodGet, "/api/events", nil).
		WithContext(authctx.WithUser(ctx, user))

	done := make(chan struct{})
	go func() {
		defer close(done)
		h.ServeHTTP(rec, req)
	}()

	waitFor(t, time.Second, func() bool {
		return strings.Contains(rec.body(), ": connected")
	})
	bus.Publish(events.Event{Type: events.TypeUpload, Path: "secret.txt", User: "alice"})
	bus.Publish(events.Event{Type: events.TypeMkdir, Path: "shared-dir", User: "bob"})
	waitFor(t, time.Second, func() bool {
		return strings.Contains(rec.body(), "shared-dir")
	})
	cancel()
	<-done

	if rec.headerValue("Content-Type") != "text/event-stream" {
		t.Fatalf("content-type = %q", rec.headerValue("Content-Type"))
	}
	body := rec.body()
	if strings.Contains(body, "secret.txt") {
		t.Fatalf("delivered another user's event: %q", body)
	}
	if !strings.Contains(body, "shared-dir") {
		t.Fatalf("missing own event: %q", body)
	}
}

func TestShouldDeliver(t *testing.T) {
	own := events.Event{Type: events.TypeDelete, User: "bob"}
	other := events.Event{Type: events.TypeDelete, User: "alice"}
	broadcast := events.Event{Type: events.TypeQuota, User: ""}

	if !shouldDeliver(broadcast, &models.User{Username: "bob"}) {
		t.Fatal("broadcast should always deliver")
	}
	if !shouldDeliver(own, &models.User{Username: "bob"}) {
		t.Fatal("own event should deliver")
	}
	if shouldDeliver(other, &models.User{Username: "bob"}) {
		t.Fatal("other user's event should not deliver")
	}
	if !shouldDeliver(other, nil) {
		t.Fatal("system view should receive all events")
	}
}

func TestEventsEventJSONShape(t *testing.T) {
	e := events.Event{Type: events.TypeShare, Path: "a.txt", User: "alice", At: time.Unix(0, 0).UTC()}
	raw, err := json.Marshal(e)
	if err != nil {
		t.Fatal(err)
	}
	var round events.Event
	if err := json.Unmarshal(raw, &round); err != nil {
		t.Fatal(err)
	}
	if round.Type != events.TypeShare || round.Path != "a.txt" || round.User != "alice" {
		t.Fatalf("round-trip mismatch: %+v", round)
	}
}
