package xhttp

import (
	"net"
	"net/http"
	"sync"
	"time"
)

// shareLimitRate is the default rate limit for share links (tokens per second).
const shareLimitRate = 10.0 // tokens refilled per second
const shareLimitBurst = 30

// ipBucket is a token bucket for a single client address.
type ipBucket struct {
	tokens float64
	last   time.Time
}

// IPLimiter is a small in-memory token bucket limiter keyed by client address.
// It is used to blunt brute-force probing of public share links and is not a
// substitute for a real edge proxy rate limiter.
type IPLimiter struct {
	mu      sync.Mutex
	buckets map[string]*ipBucket
	rate    float64 // tokens refilled per second
	burst   int
	now     func() time.Time
	keyFunc func(r *http.Request) string
}

func NewIPLimiter(rate float64, burst int) *IPLimiter {
	return &IPLimiter{
		buckets: make(map[string]*ipBucket),
		rate:    rate,
		burst:   burst,
		now:     time.Now,
		keyFunc: ClientIP,
	}
}

// Allow reports whether a request should proceed.
func (l *IPLimiter) Allow(r *http.Request) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	key := l.keyFunc(r)
	b, ok := l.buckets[key]
	if !ok {
		b = &ipBucket{tokens: float64(l.burst), last: now}
		l.buckets[key] = b
	} else {
		b.tokens += now.Sub(b.last).Seconds() * l.rate
		if b.tokens > float64(l.burst) {
			b.tokens = float64(l.burst)
		}
		b.last = now
	}

	if b.tokens < 1 {
		// Opportunistically drop long-idle buckets so the map stays bounded.
		if len(l.buckets) > 10000 {
			for k, candidate := range l.buckets {
				if now.Sub(candidate.last) > 10*time.Minute {
					delete(l.buckets, k)
				}
			}
		}
		return false
	}
	b.tokens--
	return true
}

// ClientIP returns the remote address without its port, so a client that opens
// many short-lived connections still counts as a single rate-limit bucket.
func ClientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// RateLimitMiddleware rejects requests that exceed the limiter's budget with
// 429 Too Many Requests. A nil limiter passes everything through.
func RateLimitMiddleware(limiter *IPLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if limiter == nil {
			return next
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !limiter.Allow(r) {
				w.Header().Set("Retry-After", "1")
				http.Error(w, "too many requests", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
