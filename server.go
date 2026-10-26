package servly

import (
	"net/http"
	"time"
)

// NewServer assembles the HTTP routes and middleware chain for the public
// API: rate limiting (keyed on the "X-API-Key" header) wraps auth (bearer
// token, checked against clock), which wraps the collection handler.
func NewServer(store *Store, tokens *TokenStore, clock Clock) http.Handler {
	limiter := NewRateLimiter(100, time.Minute, clock)

	var handler http.Handler = ListHandler(store)
	handler = AuthMiddleware(tokens, clock)(handler)
	handler = RateLimitMiddleware(limiter, HeaderAPIKey("X-API-Key"))(handler)

	mux := http.NewServeMux()
	mux.Handle("/records", handler)
	return mux
}
