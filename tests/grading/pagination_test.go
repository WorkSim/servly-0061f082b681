// Package servly_test contains the authoritative, hidden grading suite for
// API-301 — the collection endpoint must honor `limit`/`offset`, not return
// every row regardless of what the caller asks for. Restored from the base
// branch in CI so a candidate PR can't edit it (see the repo README).
package servly_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	servly "github.com/WorkSim/servly-template"
)

func seedRecords(n int) []servly.Record {
	records := make([]servly.Record, n)
	for i := range records {
		records[i] = servly.Record{ID: i + 1, Name: fmt.Sprintf("record-%d", i+1)}
	}
	return records
}

func decodeRecords(t *testing.T, rr *httptest.ResponseRecorder) []servly.Record {
	t.Helper()
	var got []servly.Record
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("decoding response body: %v", err)
	}
	return got
}

func TestListHonorsExplicitLimit(t *testing.T) {
	store := servly.NewStore(seedRecords(50))
	handler := servly.ListHandler(store)

	req := httptest.NewRequest(http.MethodGet, "/records?limit=20", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	got := decodeRecords(t, rr)
	if len(got) != 20 {
		t.Fatalf("expected 20 records with ?limit=20, got %d", len(got))
	}
}

func TestListAppliesDefaultLimitWhenAbsent(t *testing.T) {
	store := servly.NewStore(seedRecords(50))
	handler := servly.ListHandler(store)

	req := httptest.NewRequest(http.MethodGet, "/records", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	got := decodeRecords(t, rr)
	if len(got) != servly.DefaultLimit {
		t.Fatalf("expected default limit (%d) records, got %d", servly.DefaultLimit, len(got))
	}
}

func TestListNeverIgnoresLimitAndReturnsWholeTable(t *testing.T) {
	total := 50
	store := servly.NewStore(seedRecords(total))
	handler := servly.ListHandler(store)

	req := httptest.NewRequest(http.MethodGet, "/records?limit=5", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	got := decodeRecords(t, rr)
	if len(got) == total {
		t.Fatalf("list endpoint ignored limit and returned all %d rows", total)
	}
	if len(got) != 5 {
		t.Fatalf("expected 5 records with ?limit=5, got %d", len(got))
	}
}

func TestOffsetPaginatesToNextSlice(t *testing.T) {
	store := servly.NewStore(seedRecords(10))
	handler := servly.ListHandler(store)

	firstReq := httptest.NewRequest(http.MethodGet, "/records?limit=4&offset=0", nil)
	firstRR := httptest.NewRecorder()
	handler.ServeHTTP(firstRR, firstReq)
	first := decodeRecords(t, firstRR)

	secondReq := httptest.NewRequest(http.MethodGet, "/records?limit=4&offset=4", nil)
	secondRR := httptest.NewRecorder()
	handler.ServeHTTP(secondRR, secondReq)
	second := decodeRecords(t, secondRR)

	if len(first) != 4 || len(second) != 4 {
		t.Fatalf("expected 4+4 records across two pages, got %d+%d", len(first), len(second))
	}
	if first[0].ID != 1 || first[3].ID != 4 {
		t.Fatalf("unexpected first page IDs: %+v", first)
	}
	if second[0].ID != 5 || second[3].ID != 8 {
		t.Fatalf("unexpected second page IDs: %+v", second)
	}
}

func TestOffsetBeyondTableReturnsEmpty(t *testing.T) {
	store := servly.NewStore(seedRecords(5))
	handler := servly.ListHandler(store)

	req := httptest.NewRequest(http.MethodGet, "/records?limit=10&offset=100", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	got := decodeRecords(t, rr)
	if len(got) != 0 {
		t.Fatalf("expected empty page for out-of-range offset, got %d records", len(got))
	}
}
