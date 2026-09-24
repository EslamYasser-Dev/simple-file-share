package events

import (
	"context"
	"testing"
	"time"
)

func TestPublishDeliversToSubscriber(t *testing.T) {
	bus := NewBus()
	ch, stop := bus.Subscribe(context.Background())
	defer stop()

	bus.Publish(Event{Type: TypeUpload, Path: "a.txt", User: "alice"})

	select {
	case e := <-ch:
		if e.Type != TypeUpload || e.Path != "a.txt" || e.User != "alice" {
			t.Fatalf("unexpected event: %+v", e)
		}
		if e.At.IsZero() {
			t.Fatal("expected At to be stamped")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for event")
	}
}

func TestPublishNilBusIsNoop(t *testing.T) {
	var bus *Bus
	bus.Publish(Event{Type: TypeDelete})
	ch, stop := bus.Subscribe(context.Background())
	stop()
	if _, ok := <-ch; ok {
		t.Fatal("expected closed channel from nil bus")
	}
}

func TestSubscribeStopClosesChannel(t *testing.T) {
	bus := NewBus()
	ch, stop := bus.Subscribe(context.Background())
	stop()
	if _, ok := <-ch; ok {
		t.Fatal("expected channel closed after stop")
	}
	bus.Publish(Event{Type: TypeUpdate})
}

func TestContextCancelUnsubscribes(t *testing.T) {
	bus := NewBus()
	ctx, cancel := context.WithCancel(context.Background())
	ch, stop := bus.Subscribe(ctx)
	defer stop()
	cancel()

	select {
	case _, ok := <-ch:
		if ok {
			t.Fatal("expected closed channel after cancel")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for unsubscribe")
	}
}

func TestSlowSubscriberDropsInsteadOfBlocking(t *testing.T) {
	bus := NewBus()
	_, stop := bus.Subscribe(context.Background())
	defer stop()

	done := make(chan struct{})
	go func() {
		for i := 0; i < subscriberBuffer*3; i++ {
			bus.Publish(Event{Type: TypeUpload})
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Publish blocked on a full subscriber buffer")
	}
}

func TestMultipleSubscribers(t *testing.T) {
	bus := NewBus()
	ch1, stop1 := bus.Subscribe(context.Background())
	defer stop1()
	ch2, stop2 := bus.Subscribe(context.Background())
	defer stop2()

	bus.Publish(Event{Type: TypeMkdir, Path: "dir"})

	for _, ch := range []<-chan Event{ch1, ch2} {
		select {
		case e := <-ch:
			if e.Type != TypeMkdir {
				t.Fatalf("got %q", e.Type)
			}
		case <-time.After(time.Second):
			t.Fatal("timed out")
		}
	}
}
