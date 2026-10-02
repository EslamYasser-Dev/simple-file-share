package webhook

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/EslamYasser-Dev/simple-file-share/application/events"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/logging"
)

type received struct {
	header http.Header
	body   []byte
}

// captureServer records every POST it receives.
func captureServer(t *testing.T, status int) (*httptest.Server, chan received) {
	t.Helper()
	got := make(chan received, 64)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		// Non-blocking: retries in tests can outpace what a test consumes,
		// and a blocked handler would stall dispatcher shutdown.
		select {
		case got <- received{header: r.Header.Clone(), body: body}:
		default:
		}
		w.WriteHeader(status)
	}))
	t.Cleanup(srv.Close)
	return srv, got
}

// startDispatcher launches d.Run against a fresh bus and registers shutdown.
// It returns the bus (for publishing) and the dispatcher's stop function.
func startDispatcher(t *testing.T, d *Dispatcher) (*events.Bus, context.CancelFunc, <-chan struct{}) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	bus := events.NewBus()
	done := make(chan struct{})
	go func() {
		d.Run(ctx, bus)
		close(done)
	}()
	t.Cleanup(func() {
		cancel()
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Error("dispatcher did not stop")
		}
	})
	return bus, cancel, done
}

// publishUntilDelivered re-publishes ev until the endpoint observes it or the
// deadline passes. Run subscribes asynchronously, so a single Publish can
// race ahead of the subscription and be dropped (the bus has no replay).
func publishUntilDelivered(t *testing.T, bus *events.Bus, ev events.Event, got <-chan received) received {
	t.Helper()
	deadline := time.After(2 * time.Second)
	for {
		bus.Publish(ev)
		select {
		case msg := <-got:
			return msg
		case <-deadline:
			t.Fatal("no webhook delivered")
			return received{}
		case <-time.After(10 * time.Millisecond):
		}
	}
}

func TestDispatcherDeliversSignedEvent(t *testing.T) {
	srv, got := captureServer(t, http.StatusNoContent)
	d := NewDispatcher([]string{srv.URL}, "topsecret", logging.NewStdLogger())
	bus, _, _ := startDispatcher(t, d)

	msg := publishUntilDelivered(t, bus, events.Event{Type: events.TypeUpload, Path: "a.txt", User: "alice", Bytes: 3}, got)
	{
		if ev := msg.header.Get("X-Webhook-Event"); ev != "upload" {
			t.Errorf("X-Webhook-Event = %q, want upload", ev)
		}
		sig := msg.header.Get("X-Webhook-Signature")
		if !Verify("topsecret", msg.body, sig) {
			t.Errorf("signature %q does not verify", sig)
		}
		if Verify("wrong-secret", msg.body, sig) {
			t.Error("signature verified under the wrong secret")
		}
		var e events.Event
		if err := json.Unmarshal(msg.body, &e); err != nil {
			t.Fatalf("body %q: %v", msg.body, err)
		}
		if e.Type != "upload" || e.Path != "a.txt" || e.User != "alice" || e.Bytes != 3 {
			t.Errorf("payload = %+v", e)
		}
		if e.At.IsZero() {
			t.Error("event timestamp missing")
		}
	}
}

func TestDispatcherOmitsSignatureWithoutSecret(t *testing.T) {
	srv, got := captureServer(t, http.StatusOK)
	d := NewDispatcher([]string{srv.URL}, "", logging.NewStdLogger())
	bus, _, _ := startDispatcher(t, d)

	msg := publishUntilDelivered(t, bus, events.Event{Type: events.TypeDelete, Path: "b.txt"}, got)
	if sig := msg.header.Get("X-Webhook-Signature"); sig != "" {
		t.Errorf("unsigned dispatcher sent signature %q", sig)
	}
}

func TestDispatcherContinuesPastFailingEndpoint(t *testing.T) {
	bad, badGot := captureServer(t, http.StatusInternalServerError)
	good, goodGot := captureServer(t, http.StatusNoContent)
	d := NewDispatcher([]string{bad.URL, good.URL}, "", logging.NewStdLogger())
	bus, _, _ := startDispatcher(t, d)

	// The healthy endpoint must see the event even though its sibling is a 500.
	publishUntilDelivered(t, bus, events.Event{Type: events.TypeShare, Path: "c.txt"}, goodGot)
	// The failing endpoint was attempted for the same deliveries.
	select {
	case <-badGot:
	case <-time.After(2 * time.Second):
		t.Fatal("failing endpoint was never attempted")
	}
}

func TestDispatcherWithoutURLsReturnsImmediately(t *testing.T) {
	d := NewDispatcher(nil, "", nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // even an already-cancelled context must not matter: no urls = no loop
	done := make(chan struct{})
	go func() {
		d.Run(ctx, events.NewBus())
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Run blocked despite having no endpoints")
	}
}

func TestSignVerifyRoundTrip(t *testing.T) {
	body := []byte(`{"type":"login"}`)
	sig := Sign("k", body)
	if !Verify("k", body, sig) {
		t.Fatal("round trip failed")
	}
	if Verify("k", []byte(`{"type":"logout"}`), sig) {
		t.Fatal("tampered body verified")
	}
}

func TestDispatcherStopIsIdempotentUnderLoad(t *testing.T) {
	srv, _ := captureServer(t, http.StatusOK)
	d := NewDispatcher([]string{srv.URL}, "", logging.NewStdLogger())
	bus, cancel, done := startDispatcher(t, d)

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			bus.Publish(events.Event{Type: events.TypeUpdate, Path: "hot.txt"})
		}()
	}
	wg.Wait()
	cancel()
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("dispatcher did not shut down after publish burst")
	}
}
