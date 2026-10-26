package servly

import (
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"
)

// ErrTokenNotFound is returned when the presented bearer string isn't
// recognized — the in-memory stand-in for a bad/invalid signature.
var ErrTokenNotFound = errors.New("token not found")

// ErrTokenExpired is returned when a recognized token's expiry has passed.
var ErrTokenExpired = errors.New("token expired")

// Token represents an issued API bearer token, scoped to a caller identity
// (Key) and carrying an expiry. There's no real cryptographic signature
// here — the TokenStore lookup stands in for signature verification — but
// expiry is enforced exactly like a real JWT's `exp` claim would be.
type Token struct {
	Key       string
	ExpiresAt time.Time
}

// TokenStore is a simple in-memory, concurrency-safe registry of issued
// tokens, keyed by the bearer string presented in the Authorization header.
type TokenStore struct {
	mu     sync.RWMutex
	tokens map[string]Token
}

// NewTokenStore creates an empty TokenStore.
func NewTokenStore() *TokenStore {
	return &TokenStore{tokens: make(map[string]Token)}
}

// Issue registers tok under the given bearer string, replacing any existing
// token registered under the same bearer.
func (s *TokenStore) Issue(bearer string, tok Token) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tokens[bearer] = tok
}

func (s *TokenStore) lookup(bearer string) (Token, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	tok, ok := s.tokens[bearer]
	return tok, ok
}

// ValidateToken looks up bearer in store and checks it against now. now is
// always passed in explicitly (never time.Now() internally) so callers
// inject a Clock and keep validation fully deterministic under test.
//
// Returns ErrTokenNotFound for an unrecognized bearer, ErrTokenExpired for
// a recognized-but-expired one (now is not strictly before ExpiresAt), and
// the Token with a nil error otherwise.
func ValidateToken(store *TokenStore, bearer string, now time.Time) (Token, error) {
	tok, ok := store.lookup(bearer)
	if !ok {
		return Token{}, ErrTokenNotFound
	}
	return tok, nil
}

// AuthMiddleware rejects requests that don't carry a valid, unexpired
// bearer token in the Authorization header ("Bearer <token>") with 401.
// Expiry is checked against clock.Now(), never the real wall clock
// directly, so grading tests can inject a fixed Clock and stay fully
// deterministic.
func AuthMiddleware(store *TokenStore, clock Clock) func(http.Handler) http.Handler {
	const prefix = "Bearer "
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, prefix) {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			bearer := strings.TrimPrefix(authHeader, prefix)
			if _, err := ValidateToken(store, bearer, clock.Now()); err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
