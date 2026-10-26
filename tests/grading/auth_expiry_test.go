// Package servly_test contains the authoritative, hidden grading suite for
// API-315 — the auth middleware must reject expired bearer tokens with 401,
// not just verify that the token is recognized. Restored from the base
// branch in CI so a candidate PR can't edit it.
package servly_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	servly "github.com/WorkSim/servly-template"
)

// fixedClock is a deterministic Clock — grading tests never read the real
// wall clock.
type fixedClock struct{ now time.Time }

func (c fixedClock) Now() time.Time { return c.now }

func newAuthedHandler(store *servly.TokenStore, clock servly.Clock) http.Handler {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	return servly.AuthMiddleware(store, clock)(inner)
}

func authedRequest(bearer string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/records", nil)
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	return req
}

func TestExpiredTokenIsRejectedWith401(t *testing.T) {
	issuedAt := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	clock := fixedClock{now: issuedAt.Add(2 * time.Hour)}

	store := servly.NewTokenStore()
	store.Issue("stale-token", servly.Token{Key: "client-a", ExpiresAt: issuedAt.Add(time.Hour)})

	handler := newAuthedHandler(store, clock)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, authedRequest("stale-token"))

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for expired token, got %d", rr.Code)
	}
}

func TestValidUnexpiredTokenStillAuthenticates(t *testing.T) {
	issuedAt := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	clock := fixedClock{now: issuedAt.Add(30 * time.Minute)}

	store := servly.NewTokenStore()
	store.Issue("fresh-token", servly.Token{Key: "client-a", ExpiresAt: issuedAt.Add(time.Hour)})

	handler := newAuthedHandler(store, clock)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, authedRequest("fresh-token"))

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 for valid unexpired token, got %d", rr.Code)
	}
}

func TestTokenExpiringExactlyAtNowIsRejected(t *testing.T) {
	expiresAt := time.Date(2026, 1, 1, 1, 0, 0, 0, time.UTC)
	clock := fixedClock{now: expiresAt}

	store := servly.NewTokenStore()
	store.Issue("boundary-token", servly.Token{Key: "client-a", ExpiresAt: expiresAt})

	handler := newAuthedHandler(store, clock)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, authedRequest("boundary-token"))

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for a token expiring exactly now, got %d", rr.Code)
	}
}

func TestMissingAuthorizationHeaderIsRejected(t *testing.T) {
	clock := fixedClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	store := servly.NewTokenStore()

	handler := newAuthedHandler(store, clock)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, authedRequest(""))

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for missing Authorization header, got %d", rr.Code)
	}
}

func TestUnknownTokenIsRejected(t *testing.T) {
	clock := fixedClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	store := servly.NewTokenStore()

	handler := newAuthedHandler(store, clock)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, authedRequest("never-issued"))

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for an unrecognized token, got %d", rr.Code)
	}
}
