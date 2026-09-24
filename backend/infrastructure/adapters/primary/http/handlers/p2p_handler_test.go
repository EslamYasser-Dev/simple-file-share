package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/EslamYasser-Dev/simple-file-share/application/services"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/primary/authctx"
)

func TestP2PHandlerRejectsUnknownRoute(t *testing.T) {
	h := NewP2PHandler(services.NewP2PService())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/p2p/nope", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", rec.Code)
	}
}

func TestP2PHandlerRejectsBadMethod(t *testing.T) {
	h := NewP2PHandler(services.NewP2PService())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/p2p/peers", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", rec.Code)
	}
}

func TestP2PHandlerUnavailable(t *testing.T) {
	h := NewP2PHandler(nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/p2p/peers", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", rec.Code)
	}
}

func TestP2PHandlerStreamHello(t *testing.T) {
	svc := services.NewP2PService()
	defer svc.Stop()
	h := NewP2PHandler(svc)
	rec := newLockedRecorder()

	ctx, cancel := context.WithCancel(context.Background())
	req := httptest.NewRequest(http.MethodGet, "/api/p2p/stream", nil).WithContext(ctx)
	done := make(chan struct{})
	go func() {
		defer close(done)
		h.ServeHTTP(rec, req)
	}()

	waitFor(t, time.Second, func() bool {
		return strings.Contains(rec.body(), `"type":"hello"`) &&
			strings.Contains(rec.body(), `"peerId"`)
	})
	cancel()
	<-done

	if rec.headerValue("Content-Type") != "text/event-stream" {
		t.Fatalf("content-type = %q", rec.headerValue("Content-Type"))
	}
	body := rec.body()
	if !strings.Contains(body, "event: hello") {
		t.Fatalf("missing hello event: %q", body)
	}
	if !strings.Contains(body, "peerId") {
		t.Fatalf("missing peerId: %q", body)
	}
}

func TestP2PHandlerSignalRoundTrip(t *testing.T) {
	svc := services.NewP2PService()
	defer svc.Stop()
	h := NewP2PHandler(svc)

	// Register two peer sessions via Subscribe (as the stream handler does).
	idA, chA, stopA := svc.Subscribe(context.Background(), "alice")
	defer stopA()
	idB, chB, stopB := svc.Subscribe(context.Background(), "bob")
	defer stopB()

	// Drain hellos.
	drainP2PHello(t, chA)
	drainP2PHello(t, chB)

	body := strings.NewReader(`{"to":"` + idB + `","kind":"offer","payload":{"sdp":"v=0"}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/p2p/signal", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Peer-Id", idA)
	req = req.WithContext(authctx.WithUser(req.Context(), &models.User{Username: "alice"}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("signal status = %d body=%s", rec.Code, rec.Body.String())
	}

	deadline := time.After(time.Second)
	for {
		select {
		case ev, ok := <-chB:
			if !ok {
				t.Fatal("channel closed")
			}
			if ev.Type == services.P2PEventSignal && ev.Signal != nil {
				if ev.Signal.From != idA || ev.Signal.Kind != services.P2PSignalOffer {
					t.Fatalf("signal = %+v", ev.Signal)
				}
				return
			}
		case <-deadline:
			t.Fatal("timeout waiting for signal")
		}
	}
}

func TestP2PHandlerSignalRequiresPeerHeader(t *testing.T) {
	svc := services.NewP2PService()
	defer svc.Stop()
	h := NewP2PHandler(svc)
	req := httptest.NewRequest(http.MethodPost, "/api/p2p/signal", strings.NewReader(`{"to":"x","kind":"offer"}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestP2PHandlerSignalForbiddenUnknownSession(t *testing.T) {
	svc := services.NewP2PService()
	defer svc.Stop()
	h := NewP2PHandler(svc)
	req := httptest.NewRequest(http.MethodPost, "/api/p2p/signal", strings.NewReader(`{"to":"x","kind":"offer"}`))
	req.Header.Set("X-Peer-Id", "deadbeef")
	req = req.WithContext(authctx.WithUser(req.Context(), &models.User{Username: "alice"}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
}

func TestP2PHandlerSignalOfflinePeer(t *testing.T) {
	svc := services.NewP2PService()
	defer svc.Stop()
	h := NewP2PHandler(svc)
	idA, chA, stopA := svc.Subscribe(context.Background(), "alice")
	defer stopA()
	drainP2PHello(t, chA)

	req := httptest.NewRequest(http.MethodPost, "/api/p2p/signal",
		strings.NewReader(`{"to":"00000000000000000000000000000000","kind":"offer"}`))
	req.Header.Set("X-Peer-Id", idA)
	req = req.WithContext(authctx.WithUser(req.Context(), &models.User{Username: "alice"}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestP2PHandlerPeersList(t *testing.T) {
	svc := services.NewP2PService()
	defer svc.Stop()
	h := NewP2PHandler(svc)
	_, _, stop := svc.Subscribe(context.Background(), "alice")
	defer stop()

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/p2p/peers", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var resp struct {
		Peers []struct {
			ID   string `json:"id"`
			User string `json:"user"`
		} `json:"peers"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.Peers) != 1 || resp.Peers[0].User != "alice" {
		t.Fatalf("peers = %+v", resp.Peers)
	}
}

func drainP2PHello(t *testing.T, ch <-chan services.P2PEvent) {
	t.Helper()
	deadline := time.After(time.Second)
	for {
		select {
		case ev, ok := <-ch:
			if !ok {
				t.Fatal("closed")
			}
			if ev.Type == services.P2PEventHello {
				return
			}
		case <-deadline:
			t.Fatal("no hello")
		}
	}
}
