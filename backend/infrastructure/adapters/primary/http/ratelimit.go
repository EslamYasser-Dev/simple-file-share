package xhttp

import (
	"net"
	"net/http"
	"sync"
	"time"
)

// ipBucket is a token bucket for a single client address.
type ipBucket struct {
	tokens float64
	last   time.Time
}

// Share rate-limit budget guarding public share links. Naming the values here
// keeps the route table self-documenting and lets tests shape the budget.
const (
	shareLimitRate  = 10.0 // tokens refilled per second
	shareLimitBurst = 30
)

// IPLimiter is a small in-memory token bucket limiter keyed by client address.
// It is used to blunt brute-force probing of public share links and is not a
// substitute for a real edge proxy rate limiter.
type IPLimiter struct {
	mu      sync.Mutex
	buckets map[string]*ipBucket
	rate    float64 // tokens refilled per second
	burst   int
	now     func() time.Time
}

func NewIPLimiter(rate float64, burst int) *IPLimiter {
	return &IPLimiter{
		buckets: make(map[string]*ipBucket),
		rate:    rate,
		burst:   burst,
		now:     time.Now,
	}
}

// Allow reports whether a request from the given client should proceed.
func (l *IPLimiter) Allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	b, ok := l.buckets[ip]
	if !ok {
		b = &ipBucket{tokens: float64(l.burst), last: now}
		l.buckets[ip] = b
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
			if !limiter.Allow(ClientIP(r)) {
				w.Header().Set("Retry-After", "1")
				http.Error(w, "too many requests", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
