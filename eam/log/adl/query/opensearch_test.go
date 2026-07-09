package query_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/adl"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/adl/query"
)

// openSearchFixture builds an OpenSearch _search response with each record verbatim as
// _source, exactly as the ADL OpenSearch sink stores it.
func openSearchFixture(t *testing.T, recs []adl.Record) []byte {
	t.Helper()
	type hit struct {
		Source json.RawMessage `json:"_source"`
	}
	var hits []hit
	for i := range recs {
		raw, err := json.Marshal(recs[i])
		if err != nil {
			t.Fatal(err)
		}
		hits = append(hits, hit{Source: raw})
	}
	resp := map[string]any{"hits": map[string]any{"hits": hits}}
	buf, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}
	return buf
}

func TestOpenSearchSourceFilter(t *testing.T) {
	recs := synthetic()
	body := openSearchFixture(t, recs)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/adl/_search") {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		// The event_name pushdown must appear in the query DSL.
		raw, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(raw), "event_name") {
			t.Errorf("query DSL missing event_name pushdown: %s", raw)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	src := query.NewOpenSearchSource(query.OpenSearchConfig{URL: srv.URL, Index: "adl"})

	deny := false
	got, err := src.Query(context.Background(), query.Filter{
		EventName: adl.EventAccessEvaluation,
		Doel:      "woz",
		Decision:  &deny,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 4 { // 2 afnemers x 2 denies for woz.
		t.Fatalf("doel+deny filter: got %d, want 4", len(got))
	}
}

func TestOpenSearchSourceError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "boom", http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	src := query.NewOpenSearchSource(query.OpenSearchConfig{URL: srv.URL, Index: "adl"})
	if _, err := src.Query(context.Background(), query.Filter{}); err == nil {
		t.Fatal("expected error on 503 response")
	}
}
