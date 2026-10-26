// Package servly_test contains the authoritative, hidden grading suite for
// API-322 — the rate limiter must be enforced per API key, not as a single
// shared counter for the whole service. Restored from the base branch in CI
// so a candidate PR can't edit it.
package servly_test

import (
	"testing"
	"time"

	servly "github.com/WorkSim/servly-template"
)

// mutableClock is a deterministic Clock whose `now` can be advanced
// explicitly between assertions — grading tests never read the real wall
// clock. Every clock this file needs is defined HERE, deliberately: grading
// runs one ticket's files alone (the grader empties the grading directory and
// overlays only the ticket under test), so a helper borrowed from another
// ticket's file is undefined at compile time and fails every submission.
type mutableClock struct{ now time.Time }

func (c *mutableClock) Now() time.Time { return c.now }

func TestRateLimiterIsolatesKeysIndependently(t *testing.T) {
	clock := &mutableClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	rl := servly.NewRateLimiter(2, time.Minute, clock)

	if !rl.Allow("key-a") || !rl.Allow("key-a") {
		t.Fatalf("expected key-a's first two requests to be allowed")
	}
	if rl.Allow("key-a") {
		t.Fatalf("expected key-a's third request within the window to be denied")
	}

	// key-b must be completely unaffected by key-a exhausting its limit —
	// a global counter would deny these too.
	if !rl.Allow("key-b") {
		t.Fatalf("expected key-b's first request to be allowed independently of key-a")
	}
	if !rl.Allow("key-b") {
		t.Fatalf("expected key-b's second request to be allowed independently of key-a")
	}
}

func TestRateLimiterOneKeyExhaustingLimitDoesNotThrottleOthers(t *testing.T) {
	clock := &mutableClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	rl := servly.NewRateLimiter(1, time.Minute, clock)

	if !rl.Allow("noisy-client") {
		t.Fatalf("expected noisy-client's first request to be allowed")
	}
	// Exhaust noisy-client's limit several times over.
	for i := 0; i < 5; i++ {
		rl.Allow("noisy-client")
	}

	if !rl.Allow("quiet-client") {
		t.Fatalf("expected quiet-client's request to be allowed despite noisy-client being throttled")
	}
}

func TestRateLimiterResetsPerKeyWindowIndependently(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	clock := &mutableClock{now: now}
	rl := servly.NewRateLimiter(1, time.Minute, clock)

	if !rl.Allow("key-a") {
		t.Fatalf("expected key-a's first request to be allowed")
	}
	if rl.Allow("key-a") {
		t.Fatalf("expected key-a's second request within the window to be denied")
	}

	// Advance past key-a's window.
	clock.now = now.Add(time.Minute + time.Second)
	if !rl.Allow("key-a") {
		t.Fatalf("expected key-a's request after its window reset to be allowed")
	}

	// key-b, never having made a request, must get its own fresh window and
	// be unaffected by key-a's window state.
	if !rl.Allow("key-b") {
		t.Fatalf("expected key-b's first request to be allowed")
	}
	if rl.Allow("key-b") {
		t.Fatalf("expected key-b's second request within its own window to be denied")
	}
}
