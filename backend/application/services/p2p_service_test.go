package services

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func drainUntil(t *testing.T, ch <-chan P2PEvent, timeout time.Duration, pred func(P2PEvent) bool) P2PEvent {
	t.Helper()
	deadline := time.After(timeout)
	for {
		select {
		case ev, ok := <-ch:
			if !ok {
				t.Fatal("channel closed before match")
			}
			if pred(ev) {
				return ev
			}
		case <-deadline:
			t.Fatal("timeout waiting for event")
		}
	}
}

func TestP2PSubscribeHelloAndPeers(t *testing.T) {
	svc := NewP2PService()
	defer svc.Stop()

	idA, chA, stopA := svc.Subscribe(context.Background(), "alice")
	defer stopA()
	if idA == "" {
		t.Fatal("empty peer id")
	}
	hello := drainUntil(t, chA, time.Second, func(e P2PEvent) bool { return e.Type == P2PEventHello })
	if hello.PeerID != idA {
		t.Fatalf("hello peerId = %q, want %q", hello.PeerID, idA)
	}

	idB, chB, stopB := svc.Subscribe(context.Background(), "bob")
	defer stopB()
	if idB == idA {
		t.Fatal("peer ids must be unique")
	}
	_ = chB

	peers := svc.ListPeers(idA)
	if len(peers) != 1 || peers[0].ID != idB {
		t.Fatalf("list peers = %+v", peers)
	}
}

func TestP2PSendSignalDelivery(t *testing.T) {
	svc := NewP2PService()
	defer svc.Stop()

	idA, chA, stopA := svc.Subscribe(context.Background(), "alice")
	defer stopA()
	idB, chB, stopB := svc.Subscribe(context.Background(), "bob")
	defer stopB()

	drainUntil(t, chA, time.Second, func(e P2PEvent) bool { return e.Type == P2PEventHello })
	drainUntil(t, chB, time.Second, func(e P2PEvent) bool { return e.Type == P2PEventHello })

	payload := json.RawMessage(`{"sdp":"v=0"}`)
	if err := svc.SendSignal(idA, idB, P2PSignalOffer, payload); err != nil {
		t.Fatalf("SendSignal: %v", err)
	}

	ev := drainUntil(t, chB, time.Second, func(e P2PEvent) bool { return e.Type == P2PEventSignal })
	if ev.Signal == nil {
		t.Fatal("nil signal")
	}
	if ev.Signal.From != idA || ev.Signal.To != idB || ev.Signal.Kind != P2PSignalOffer {
		t.Fatalf("signal = %+v", ev.Signal)
	}
	if string(ev.Signal.Payload) != `{"sdp":"v=0"}` {
		t.Fatalf("payload = %s", ev.Signal.Payload)
	}

	select {
	case e := <-chA:
		if e.Type == P2PEventSignal {
			t.Fatalf("sender received own signal: %+v", e)
		}
	case <-time.After(50 * time.Millisecond):
	}
}

func TestP2PSendSignalValidation(t *testing.T) {
	svc := NewP2PService()
	defer svc.Stop()
	idA, _, stopA := svc.Subscribe(context.Background(), "alice")
	defer stopA()

	if err := svc.SendSignal(idA, idA, P2PSignalOffer, nil); err != ErrInvalidSignal {
		t.Fatalf("self signal err = %v", err)
	}
	if err := svc.SendSignal(idA, "missing", P2PSignalOffer, nil); err != ErrPeerOffline {
		t.Fatalf("offline err = %v", err)
	}
	if err := svc.SendSignal(idA, "missing", "bogus", nil); err != ErrInvalidSignal {
		t.Fatalf("kind err = %v", err)
	}
	if err := svc.SendSignal(idA, "missing", P2PSignalOffer, json.RawMessage("{")); err != ErrInvalidSignal {
		t.Fatalf("json err = %v", err)
	}
}

func TestP2PPeerOfflineAfterStop(t *testing.T) {
	svc := NewP2PService()
	idA, _, stopA := svc.Subscribe(context.Background(), "alice")
	defer stopA()
	idB, _, stopB := svc.Subscribe(context.Background(), "bob")

	stopB()
	// Unregister is async with ctx cancel path; stopB runs synchronously.
	if err := svc.SendSignal(idA, idB, P2PSignalOffer, nil); err != ErrPeerOffline {
		t.Fatalf("after stop err = %v", err)
	}
}
