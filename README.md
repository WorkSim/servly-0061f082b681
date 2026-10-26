# Servly

Servly is a minimal Go `net/http` REST API service used as a **debugging sandbox** within the WorkSim platform. It models a public collection endpoint, bearer-token auth, and per-API-key rate limiting as pure, deterministic Go code, with intentional, documented bugs planted for candidate learning and assessment.

## Overview

**What it is:** A dependency-free (stdlib-only) Go module with:
- A paginated collection endpoint (`GET /records`, honoring `limit`/`offset`)
- Bearer-token auth middleware with expiry enforcement (via an injected `Clock`, never the real wall clock)
- A per-API-key rate limiter (independent counter + window per key)
- Intentional seeded bugs (via patches) that candidates must fix

**What it's for:** WorkSim uses Servly to test backend problem-solving skills — pagination correctness, auth/expiry handling, and per-tenant rate limiting — the kind of bugs that show up in real production APIs.

**Key constraint:** This is a standalone template repo. `main` here is the correct, unmodified baseline implementation. At provision time, the platform plants every bug in the sprint into a student's own copy of the repo's root commit — so while this template's `main` stays clean, a student's `main` carries all of the sprint's known defects from the start. The assigned ticket names the one in scope; unrelated failures are other tickets' bugs, not something to fix. There is no database, no real clock read, and no network call anywhere in the graded code paths — every function is a pure or in-memory-store-backed transformation of its inputs, driven by an injected `Clock`, so `go test` runs are fully deterministic.

## Quick Start

### Prerequisites
- Go 1.25+

### Installation & Development

```bash
# Download dependencies (none beyond the stdlib — this is effectively instant)
go mod download

# Run all tests (visible + grading)
go test ./...

# Run only the visible tests
go test ./tests/visible/...

# Run only the grading tests (authoritative test suite)
go test ./tests/grading/...

# Vet
go vet ./...

# Run the service locally
go run ./cmd/servly
```

## Project Structure

```
servly/
├── go.mod
├── clock.go             # Clock interface + RealClock — the only place time.Now() is called
├── records.go            # Record, Store, ListHandler — the paginated collection endpoint
├── auth.go               # Token, TokenStore, AuthMiddleware — bearer-token auth + expiry
├── ratelimit.go           # RateLimiter, RateLimitMiddleware — per-API-key rate limiting
├── server.go              # NewServer — wires the middleware chain + routes
├── cmd/servly/main.go      # Entry point — the only place a real network listener is opened
├── tests/
│   ├── visible/             # VISIBLE tests (develop against these)
│   │   ├── pagination_test.go
│   │   ├── auth_test.go
│   │   └── ratelimit_test.go
│   └── grading/              # GRADING tests (restored from baseline in CI)
│       ├── pagination_test.go
│       ├── auth_expiry_test.go
│       └── rate_limit_test.go
├── .github/workflows/
│   └── grade.yml                # CI grading workflow
└── .gitignore
```

## Testing Strategy

Servly uses a **two-tier test split** for learning and assessment:

### Visible Tests (`tests/visible/`)
- **Purpose:** For candidates to develop and validate their fixes
- **Run:** `go test ./tests/visible/...`
- **Edit:** Safe to edit and experiment with

### Grading Tests (`tests/grading/`)
- **Purpose:** Authoritative test suite for final assessment
- **Run:** `go test ./tests/grading/...`
- **Edit:** Do not edit — these are restored from the baseline in CI
- **Package:** All grading tests live in an external test package (`package servly_test`) that imports only Servly's exported API — they never reach into unexported internals, so they can't collide with student edits to package-internal files
- **Coverage:** One file per seeded bug
  - `pagination_test.go` — API-301 (limit/offset pagination)
  - `auth_expiry_test.go` — API-315 (token expiry enforcement)
  - `rate_limit_test.go` — API-322 (per-API-key rate limiting)

**Important:** This template repo's own `main` is the correct baseline. A student's provisioned repo ships with the sprint's known defects already committed to `main`; the assigned ticket names the one in scope. Running `go test ./...` in a student's repo may show unrelated failures — those are other tickets' bugs, not something to fix.

## CI Grading Workflow

The `.github/workflows/grade.yml` workflow runs on every pull request targeting `main`:

1. Checks out the PR's merge ref (candidate code + grading baseline merged)
2. Restores `tests/grading`, `go.mod`, and `.github/grading-map.json` from the base branch — so a PR cannot fake a pass by editing those files
3. **Scopes the run to the ticket the PR is for**, then installs with `go mod download` and runs `go test ./tests/grading/...`

### Why scoping matters

Every one of the sprint's bugs is planted in your repo's root commit from day one. An
unscoped run therefore fails on the tickets you have not been assigned yet, so a
perfectly correct day-1 PR would show a red check. That is noise, not signal.

The workflow finds the ticket id in your **branch name**, then the **PR title**, then the
**PR body** — first match wins — and removes the other tickets' grading files before
running. Name your branch after the ticket and the check grades only your work:

```
git checkout -b api-301-short-description
```

If no ticket id is found anywhere, the workflow runs everything and says so in the job
summary rather than pretending the result is meaningful.

## Development Tips

### Understanding the Modules

- `records.go` — `Store.List(limit, offset)` returns the requested page (capped at `MaxLimit`, defaulting to `DefaultLimit` when `limit` is absent/invalid) and never mutates the store. `ListHandler` wires this to `GET /records` via `ParseLimitOffset`.
- `auth.go` — `ValidateToken(store, bearer, now)` looks up the bearer string (standing in for signature verification) and rejects it as expired when `now` is not strictly before `Token.ExpiresAt`. `AuthMiddleware` wires this into the request chain, sourcing `now` from an injected `Clock`.
- `ratelimit.go` — `RateLimiter.Allow(key)` enforces a fixed-window limit per key, each with its own independent counter and reset time. `RateLimitMiddleware` wires this in, keyed by the `X-API-Key` header by default.
- `clock.go` — `Clock` is the only source of "now" anywhere in the graded code. Production wiring (`cmd/servly/main.go`) uses `RealClock`; every test injects a fixed or steppable clock.

### Running Tests Locally

```bash
# All tests
go test ./...

# Verbose
go test ./... -v

# One package
go test ./tests/visible/... -v
```

## Troubleshooting

**Q: Tests fail on `main`?**
A: Expected. Your repo ships with the sprint's known defects; your ticket names the one in scope. Unrelated failures are other tickets' bugs, not something to revert.

**Q: Grading tests are passing but I expected them to fail?**
A: All the sprint's known defects are already committed on `main` when your repo is provisioned; this template's own `main` is the correct baseline. Your ticket names the one in scope for grading.

**Q: Can I edit the grading tests?**
A: Not in the template. CI always restores them from the baseline. Edits in your fork are local only; the authoritative grading happens against the restored tests.

## Environment Variables

Servly uses no external APIs, database, or environment variables. All state lives in in-memory stores (see `tests/` for example usage), keeping the service fully deterministic under test.

---

**Maintained By:** WorkSim Team
