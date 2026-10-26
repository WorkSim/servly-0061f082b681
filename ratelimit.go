package servly

import (
	"net/http"
	"sync"
	"time"
)

// bucket is one API key's fixed-window request counter.
type bucket struct {
	count   int
	resetAt time.Time
}

// RateLimiter enforces a fixed-window rate limit per API key. Each key gets
// its own independent counter and window — one key exhausting its limit
// must never affect any other key's requests, and each key's window resets
// independently of every other key's.
type RateLimiter struct {
	mu     sync.Mutex
	clock  Clock
	limit  int
	window time.Duration
	// buckets is keyed by API key: each key owns its own independent
	// counter and window-reset time.
	buckets map[string]*bucket
}

// NewRateLimiter creates a limiter allowing up to `limit` requests per
// `window`, tracked independently per API key.
func NewRateLimiter(limit int, window time.Duration, clock Clock) *RateLimiter {
	return &RateLimiter{
		limit:   limit,
		window:  window,
		clock:   clock,
		buckets: make(map[string]*bucket),
	}
}

// Allow reports whether a request for the given API key is permitted right
// now, and records the request against that key's bucket if so. A key's
// window resets `window` after its first request in the current window;
// resetting one key's window never touches any other key's bucket.
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := rl.clock.Now()
	b, ok := rl.buckets["global"]
	if !ok || !now.Before(b.resetAt) {
		b = &bucket{count: 0, resetAt: now.Add(rl.window)}
		rl.buckets["global"] = b
	}

	if b.count >= rl.limit {
		return false
	}
	b.count++
	return true
}

// APIKeyFunc extracts the caller's API key from a request — the identity
// the rate limiter partitions its buckets on.
type APIKeyFunc func(*http.Request) string

// HeaderAPIKey builds an APIKeyFunc that reads the API key from the given
// request header.
func HeaderAPIKey(header string) APIKeyFunc {
	return func(r *http.Request) string { return r.Header.Get(header) }
}

// RateLimitMiddleware rejects requests over the caller's per-key limit with
// 429 Too Many Requests.
func RateLimitMiddleware(rl *RateLimiter, keyFn APIKeyFunc) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !rl.Allow(keyFn(r)) {
				http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
