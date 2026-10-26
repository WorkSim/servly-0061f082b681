// Package servly_test holds the VISIBLE tests — develop and iterate against
// these locally. The authoritative grading suite lives in tests/grading.
package servly_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	servly "github.com/WorkSim/servly-template"
)

func TestListDefaultsToDefaultLimit(t *testing.T) {
	records := make([]servly.Record, 30)
	for i := range records {
		records[i] = servly.Record{ID: i + 1}
	}
	store := servly.NewStore(records)

	req := httptest.NewRequest(http.MethodGet, "/records", nil)
	rr := httptest.NewRecorder()
	servly.ListHandler(store).ServeHTTP(rr, req)

	var got []servly.Record
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if len(got) != servly.DefaultLimit {
		t.Fatalf("expected %d records by default, got %d", servly.DefaultLimit, len(got))
	}
}

func TestListRespectsExplicitLimit(t *testing.T) {
	records := make([]servly.Record, 10)
	for i := range records {
		records[i] = servly.Record{ID: i + 1}
	}
	store := servly.NewStore(records)

	req := httptest.NewRequest(http.MethodGet, "/records?limit=3", nil)
	rr := httptest.NewRecorder()
	servly.ListHandler(store).ServeHTTP(rr, req)

	var got []servly.Record
	_ = json.Unmarshal(rr.Body.Bytes(), &got)
	if len(got) != 3 {
		t.Fatalf("expected 3 records, got %d", len(got))
	}
}

func TestListOffsetSkipsEarlierRows(t *testing.T) {
	store := servly.NewStore([]servly.Record{{ID: 1}, {ID: 2}, {ID: 3}, {ID: 4}})

	req := httptest.NewRequest(http.MethodGet, "/records?limit=2&offset=2", nil)
	rr := httptest.NewRecorder()
	servly.ListHandler(store).ServeHTTP(rr, req)

	var got []servly.Record
	_ = json.Unmarshal(rr.Body.Bytes(), &got)
	if len(got) != 2 || got[0].ID != 3 {
		t.Fatalf("expected page starting at id 3, got %+v", got)
	}
}
