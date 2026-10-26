// Package servly_test holds the VISIBLE tests — develop and iterate against
// these locally. The authoritative grading suite lives in tests/grading.
package servly_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	servly "github.com/WorkSim/servly-template"
)

// staticClock is a fixed Clock for local development tests.
type staticClock struct{ now time.Time }

func (c staticClock) Now() time.Time { return c.now }

func TestValidTokenAuthenticates(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	store := servly.NewTokenStore()
	store.Issue("good-token", servly.Token{Key: "client", ExpiresAt: now.Add(time.Hour)})

	handler := servly.AuthMiddleware(store, staticClock{now: now})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/records", nil)
	req.Header.Set("Authorization", "Bearer good-token")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestExpiredTokenIsRejected(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	store := servly.NewTokenStore()
	store.Issue("old-token", servly.Token{Key: "client", ExpiresAt: now.Add(-time.Minute)})

	handler := servly.AuthMiddleware(store, staticClock{now: now})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/records", nil)
	req.Header.Set("Authorization", "Bearer old-token")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestUnknownTokenIsRejected(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	store := servly.NewTokenStore()

	handler := servly.AuthMiddleware(store, staticClock{now: now})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/records", nil)
	req.Header.Set("Authorization", "Bearer nonexistent")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}
