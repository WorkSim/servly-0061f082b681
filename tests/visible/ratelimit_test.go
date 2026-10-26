// Package servly_test holds the VISIBLE tests — develop and iterate against
// these locally. The authoritative grading suite lives in tests/grading.
package servly_test

import (
	"testing"
	"time"

	servly "github.com/WorkSim/servly-template"
)

func TestRateLimiterAllowsUpToLimit(t *testing.T) {
	clock := staticClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	rl := servly.NewRateLimiter(3, time.Minute, clock)

	for i := 0; i < 3; i++ {
		if !rl.Allow("client-1") {
			t.Fatalf("expected request %d to be allowed", i+1)
		}
	}
	if rl.Allow("client-1") {
		t.Fatalf("expected 4th request to be denied")
	}
}

func TestRateLimiterTracksKeysSeparately(t *testing.T) {
	clock := staticClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	rl := servly.NewRateLimiter(1, time.Minute, clock)

	if !rl.Allow("client-a") {
		t.Fatalf("expected client-a's request to be allowed")
	}
	if !rl.Allow("client-b") {
		t.Fatalf("expected client-b's request to be allowed independently of client-a")
	}
}
