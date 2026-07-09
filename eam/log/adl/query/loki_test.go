package query_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/adl"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/adl/query"
)

// lokiFixture builds a Loki query_range JSON response from records, grouping them into
// streams by (event_name, decision) and encoding each record as one log line = json(rec),
// exactly as the collector's Loki exporter would store them.
func lokiFixture(t *testing.T, recs []adl.Record) []byte {
	t.Helper()
	type stream struct {
		Stream map[string]string `json:"stream"`
		Values [][]string        `json:"values"`
	}
	byKey := map[string]*stream{}
	for i := range recs {
		r := &recs[i]
		dec := adl.DecisionLabel(r)
		key := r.EventName + "|" + dec
		s := byKey[key]
		if s == nil {
			s = &stream{Stream: map[string]string{
				adl.LabelJob:       adl.JobValue,
				adl.LabelEventName: r.EventName,
				adl.LabelDecision:  dec,
			}}
			byKey[key] = s
		}
		line, err := json.Marshal(r)
		if err != nil {
			t.Fatal(err)
		}
		ts := strconv.FormatInt(int64(r.Timestamp)*1_000_000, 10)
		s.Values = append(s.Values, []string{ts, string(line)})
	}

	var streams []stream
	for _, s := range byKey {
		streams = append(streams, *s)
	}
	resp := map[string]any{
		"status": "success",
		"data":   map[string]any{"resultType": "streams", "result": streams},
	}
	buf, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}
	return buf
}

func newLokiServer(t *testing.T, recs []adl.Record, wantSelector []string) *httptest.Server {
	t.Helper()
	body := lokiFixture(t, recs)
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/loki/api/v1/query_range") {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		sel := r.URL.Query().Get("query")
		for _, want := range wantSelector {
			if !strings.Contains(sel, want) {
				t.Errorf("selector %q missing %q", sel, want)
			}
		}
		if r.URL.Query().Get("start") == "" || r.URL.Query().Get("end") == "" {
			t.Errorf("missing start/end: %q", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
}

func TestLokiSourceFilter(t *testing.T) {
	recs := synthetic()
	srv := newLokiServer(t, recs, []string{`job="adl"`})
	defer srv.Close()

	src := query.NewLokiSource(query.LokiConfig{URL: srv.URL})
	ctx := context.Background()

	// Afnemer filter (client-side).
	got, err := src.Query(ctx, query.Filter{Afnemer: "gemeente-amsterdam"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 24 {
		t.Fatalf("afnemer filter: got %d, want 24", len(got))
	}

	// Doel + deny.
	deny := false
	got, err = src.Query(ctx, query.Filter{Doel: "woz", Decision: &deny})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 4 {
		t.Fatalf("doel+deny filter: got %d, want 4", len(got))
	}
}

// TestLokiSourcePushdown asserts the event_name and decision labels reach the selector.
func TestLokiSourcePushdown(t *testing.T) {
	recs := synthetic()
	permit := true
	srv := newLokiServer(t, recs, []string{
		`job="adl"`,
		`event_name="adl.access_evaluation"`,
		`decision="permit"`,
	})
	defer srv.Close()

	src := query.NewLokiSource(query.LokiConfig{URL: srv.URL})
	got, err := src.Query(context.Background(), query.Filter{
		EventName: adl.EventAccessEvaluation,
		Decision:  &permit,
	})
	if err != nil {
		t.Fatal(err)
	}
	// 2 afnemers x 3 doelen x 6 permits = 36.
	if len(got) != 36 {
		t.Fatalf("permit filter: got %d, want 36", len(got))
	}
}

func TestLokiSourceError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	src := query.NewLokiSource(query.LokiConfig{URL: srv.URL})
	if _, err := src.Query(context.Background(), query.Filter{}); err == nil {
		t.Fatal("expected error on 500 response")
	}
}
