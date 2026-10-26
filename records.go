package servly

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"sync"
)

// DefaultLimit is applied to the collection endpoint when the caller omits
// the `limit` query param.
const DefaultLimit = 20

// MaxLimit caps `limit` regardless of what the caller requests, so a client
// can't force an unbounded scan by passing an absurdly large value.
const MaxLimit = 100

// Record is a single row in the in-memory collection the list endpoint
// serves.
type Record struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Store is an in-memory, concurrency-safe collection of Records — a
// stand-in for a real database table.
type Store struct {
	mu      sync.RWMutex
	records []Record
}

// NewStore creates a Store pre-populated with records, preserving order.
// The provided slice is copied, so later mutation by the caller never
// affects the Store.
func NewStore(records []Record) *Store {
	cp := make([]Record, len(records))
	copy(cp, records)
	return &Store{records: cp}
}

// List returns the page of records starting at offset, at most limit long.
// It never mutates the store, never returns more than MaxLimit records, and
// never panics on an out-of-range offset — it returns an empty slice
// instead.
func (s *Store) List(limit, offset int) []Record {
	s.mu.RLock()
	defer s.mu.RUnlock()

	page := make([]Record, len(s.records))
	copy(page, s.records)
	return page
}

// ParseLimitOffset reads `limit`/`offset` from query params. `limit` falls
// back to DefaultLimit when absent, zero, negative, or unparsable; `offset`
// falls back to 0 under the same conditions.
func ParseLimitOffset(q url.Values) (limit, offset int) {
	limit = DefaultLimit
	if raw := q.Get("limit"); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v > 0 {
			limit = v
		}
	}

	offset = 0
	if raw := q.Get("offset"); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v >= 0 {
			offset = v
		}
	}

	return limit, offset
}

// ListHandler serves the collection endpoint, honoring `limit`/`offset`
// query params and encoding the resulting page as a JSON array.
func ListHandler(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit, offset := ParseLimitOffset(r.URL.Query())
		page := store.List(limit, offset)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(page)
	}
}
