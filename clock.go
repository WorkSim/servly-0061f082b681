// Package servly is a minimal REST API service (net/http, no framework)
// used as a debugging sandbox within the WorkSim platform. The core logic —
// pagination, token expiry, and per-key rate limiting — is pure and
// deterministic: no real network calls and no direct time.Now() reads
// anywhere in the graded code paths. Time always flows through the Clock
// interface below, and state lives in in-memory stores, so grading tests
// never depend on wall-clock time or external services.
package servly

import "time"

// Clock abstracts "now" so time-dependent logic (token expiry, rate-limit
// window resets) never calls time.Now() directly. Production wiring uses
// RealClock; tests inject a fixed or stepped implementation for fully
// deterministic assertions.
type Clock interface {
	Now() time.Time
}

// RealClock is the production Clock, backed by the real wall clock in UTC.
type RealClock struct{}

// Now returns the current time in UTC.
func (RealClock) Now() time.Time { return time.Now().UTC() }
