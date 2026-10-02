// Package webhook fans application events out to configured HTTP(S)
// endpoints. Deliveries are best-effort: a slow or failing receiver is logged
// and skipped, never propagated back to the caller that published the event.
package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/EslamYasser-Dev/simple-file-share/application/events"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/logging"
)

const (
	// workerCount bounds concurrent deliveries; the queue in front of them
	// absorbs bursts. When both are full the event is dropped — the same
	// backpressure philosophy the SSE bus already uses.
	workerCount = 4
	queueSize   = 64
	// deliveryTimeout bounds one HTTP POST so a hung receiver can only pin a
	// worker for a few seconds.
	deliveryTimeout = 5 * time.Second
)

// Dispatcher subscribes to the event bus and POSTs JSON copies of every event
// to each configured endpoint.
type Dispatcher struct {
	urls   []string
	secret string
	client *http.Client
	logger *logging.StdLogger
}

// NewDispatcher builds a dispatcher for the given endpoints. An empty URL
// list produces a no-op dispatcher (Run returns immediately).
func NewDispatcher(urls []string, secret string, logger *logging.StdLogger) *Dispatcher {
	return &Dispatcher{
		urls:   urls,
		secret: secret,
		client: &http.Client{Timeout: deliveryTimeout},
		logger: logger,
	}
}

// Run consumes bus events until ctx is cancelled, then drains in-flight
// deliveries before returning. It is meant to be launched as a goroutine.
func (d *Dispatcher) Run(ctx context.Context, bus *events.Bus) {
	if len(d.urls) == 0 {
		return
	}
	in, stop := bus.Subscribe(ctx)
	defer stop()

	queue := make(chan events.Event, queueSize)
	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for e := range queue {
				d.deliver(e)
			}
		}()
	}
	defer func() {
		close(queue)
		wg.Wait()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case e, ok := <-in:
			if !ok {
				return
			}
			select {
			case queue <- e:
			default:
				d.warn("webhook queue full, dropping event", "type", e.Type)
			}
		}
	}
}

// warn is a nil-safe logger shortcut (a dispatcher may run without one in
// tests or embedded use).
func (d *Dispatcher) warn(msg string, kv ...any) {
	if d.logger != nil {
		d.logger.Warn(msg, kv...)
	}
}

// deliver POSTs the event to every endpoint sequentially. Failures are
// logged per endpoint and never abort the remaining ones.
func (d *Dispatcher) deliver(e events.Event) {
	body, err := json.Marshal(e)
	if err != nil {
		d.warn("webhook marshal failed", "type", e.Type, "error", err)
		return
	}
	for _, endpoint := range d.urls {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, endpoint, bytes.NewReader(body))
		if err != nil {
			d.warn("webhook request build failed", "endpoint", endpoint, "error", err)
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "simple-file-share-webhook/1")
		req.Header.Set("X-Webhook-Event", e.Type)
		if d.secret != "" {
			req.Header.Set("X-Webhook-Signature", Sign(d.secret, body))
		}
		resp, err := d.client.Do(req)
		if err != nil {
			d.warn("webhook delivery failed", "endpoint", endpoint, "type", e.Type, "error", err)
			continue
		}
		// Drain so the connection can be reused, then close.
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			d.warn("webhook rejected", "endpoint", endpoint, "type", e.Type, "status", resp.StatusCode)
		}
	}
}

// Sign returns the X-Webhook-Signature value for a payload: HMAC-SHA256 of
// the raw body hex-encoded under "sha256=". Receivers recompute it over the
// exact bytes they read to detect tampering.
func Sign(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

// Verify reports whether signature matches the HMAC of body under secret.
func Verify(secret string, body []byte, signature string) bool {
	expected := Sign(secret, body)
	return hmac.Equal([]byte(expected), []byte(signature))
}
