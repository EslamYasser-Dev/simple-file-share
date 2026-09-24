package events

import (
	"context"
	"sync"
	"time"
)

// Event types emitted by application services for live UI refresh.
const (
	TypeUpload      = "upload"
	TypeUpdate      = "update"
	TypeMkdir       = "mkdir"
	TypeDelete      = "delete"
	TypeShare       = "share"
	TypeShareRevoke = "share_revoke"
	TypeRestore     = "restore"
	TypeQuota       = "quota"
	TypeDownload    = "download"
	TypeLogin       = "login"
)

// Event is one filesystem or account change notification.
type Event struct {
	Type  string    `json:"type"`
	Path  string    `json:"path,omitempty"`
	User  string    `json:"user,omitempty"`
	Bytes int64     `json:"bytes,omitempty"`
	At    time.Time `json:"at"`
}

const subscriberBuffer = 32

// Bus fans events out to live subscribers (SSE clients). Publish never blocks:
// a slow subscriber drops events rather than stalling writers.
type Bus struct {
	mu   sync.Mutex
	subs map[uint64]chan Event
	next uint64
}

func NewBus() *Bus {
	return &Bus{subs: make(map[uint64]chan Event)}
}

// Publish delivers e to every current subscriber. A nil bus is a no-op.
func (b *Bus) Publish(e Event) {
	if b == nil {
		return
	}
	if e.At.IsZero() {
		e.At = time.Now().UTC()
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, ch := range b.subs {
		select {
		case ch <- e:
		default:
		}
	}
}

// Subscribe registers a new subscriber until ctx is cancelled or the returned
// stop function is called. A nil bus yields a closed channel and a no-op stop.
func (b *Bus) Subscribe(ctx context.Context) (<-chan Event, func()) {
	if b == nil {
		ch := make(chan Event)
		close(ch)
		return ch, func() {}
	}

	ch := make(chan Event, subscriberBuffer)
	b.mu.Lock()
	b.next++
	id := b.next
	b.subs[id] = ch
	b.mu.Unlock()

	var once sync.Once
	stop := func() {
		once.Do(func() {
			b.mu.Lock()
			if _, ok := b.subs[id]; ok {
				delete(b.subs, id)
				close(ch)
			}
			b.mu.Unlock()
		})
	}

	if ctx != nil && ctx.Done() != nil {
		go func() {
			<-ctx.Done()
			stop()
		}()
	}
	return ch, stop
}
