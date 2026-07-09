package query_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/adl"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/adl/query"
)

// mkRecord builds a synthetic ADL record for an afnemer/doel/decision at a time.
func mkRecord(traceID, afnemer, doel string, permit bool, ts time.Time) adl.Record {
	return adl.Record{
		TraceID:   traceID,
		SpanID:    "0000000000000001",
		EventName: adl.EventAccessEvaluation,
		Timestamp: uint64(ts.UnixMilli()),
		Status:    adl.StatusOk,
		Attributes: map[string]any{
			adl.KeyTransactionID: "tx-" + traceID,
		},
		Body: map[string]any{
			adl.KeyRequest: map[string]any{
				"subject": map[string]any{"type": "organisation", "id": afnemer},
				"action":  map[string]any{"name": "read"},
				"context": map[string]any{"purpose": doel},
			},
			adl.KeyResponse: map[string]any{"decision": permit},
		},
	}
}

// synthetic builds records for 2 afnemers x 3 doelen on a fixed day.
func synthetic() []adl.Record {
	day := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
	afnemers := []string{"gemeente-amsterdam", "gemeente-utrecht"}
	doelen := []string{"opsporing", "huisvestingswet", "woz"}

	var recs []adl.Record
	n := 0
	for _, a := range afnemers {
		for _, d := range doelen {
			// 6 permits + 2 denies per (afnemer, doel): total 8 >= default k=5.
			for i := 0; i < 6; i++ {
				n++
				recs = append(recs, mkRecord(traceID(n), a, d, true, day))
			}
			for i := 0; i < 2; i++ {
				n++
				recs = append(recs, mkRecord(traceID(n), a, d, false, day))
			}
		}
	}
	return recs
}

func traceID(n int) string {
	const hex = "0123456789abcdef"
	b := make([]byte, 32)
	for i := range b {
		b[i] = '0'
	}
	b[31] = hex[n%16]
	b[30] = hex[(n/16)%16]
	return string(b)
}

func writeWAL(t *testing.T, recs []adl.Record) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "adl.jsonl")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	enc := json.NewEncoder(f)
	for i := range recs {
		if err := enc.Encode(recs[i]); err != nil {
			t.Fatal(err)
		}
	}
	return path
}

func TestWALSourceFilter(t *testing.T) {
	recs := synthetic()
	src := query.NewWALSource(writeWAL(t, recs))
	ctx := context.Background()

	// Filter by afnemer.
	got, err := src.Query(ctx, query.Filter{Afnemer: "gemeente-amsterdam"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 24 { // 3 doelen x 8 records.
		t.Fatalf("afnemer filter: got %d, want 24", len(got))
	}

	// Filter by doel + decision=deny.
	deny := false
	got, err = src.Query(ctx, query.Filter{Doel: "woz", Decision: &deny})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 4 { // 2 afnemers x 2 denies.
		t.Fatalf("doel+deny filter: got %d, want 4", len(got))
	}

	// Filter by event name that matches nothing.
	got, err = src.Query(ctx, query.Filter{EventName: adl.EventSearchSubject})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("event filter: got %d, want 0", len(got))
	}
}

func TestWALSourcePeriodFilter(t *testing.T) {
	recs := synthetic()
	// Add a record from a different month.
	recs = append(recs, mkRecord(traceID(999), "gemeente-utrecht", "woz", true,
		time.Date(2026, 8, 15, 9, 0, 0, 0, time.UTC)))
	src := query.NewWALSource(writeWAL(t, recs))

	got, err := src.Query(context.Background(), query.Filter{
		From: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("period filter: got %d, want 1", len(got))
	}
}

func TestAggregateCountsAndKThreshold(t *testing.T) {
	recs := synthetic()
	agg := query.Aggregate(recs, query.AggregateOptions{Bucket: query.BucketDay, K: 5})

	// 2 afnemers x 3 doelen = 6 buckets, all with total 8 >= k.
	if len(agg) != 6 {
		t.Fatalf("aggregate: got %d buckets, want 6", len(agg))
	}
	for _, c := range agg {
		if c.Permit != 6 || c.Deny != 2 || c.Total != 8 {
			t.Fatalf("bucket %+v: unexpected counts", c)
		}
		if c.Periode != "2026-07-01" {
			t.Fatalf("bucket period = %q, want 2026-07-01", c.Periode)
		}
	}

	// Deterministic ordering: first bucket is amsterdam/huisvestingswet.
	if agg[0].Afnemer != "gemeente-amsterdam" || agg[0].Doel != "huisvestingswet" {
		t.Fatalf("ordering: first bucket = %+v", agg[0].Key)
	}
}

func TestAggregateKSuppression(t *testing.T) {
	day := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
	var recs []adl.Record
	// One large bucket (10 records) and one tiny bucket (3 records < k=5).
	for i := 0; i < 10; i++ {
		recs = append(recs, mkRecord(traceID(i), "afnemer-a", "doel-big", true, day))
	}
	for i := 0; i < 3; i++ {
		recs = append(recs, mkRecord(traceID(100+i), "afnemer-b", "doel-small", true, day))
	}

	agg := query.Aggregate(recs, query.AggregateOptions{K: 5})
	if len(agg) != 1 {
		t.Fatalf("k-suppression: got %d buckets, want 1", len(agg))
	}
	if agg[0].Doel != "doel-big" {
		t.Fatalf("k-suppression kept wrong bucket: %+v", agg[0].Key)
	}

	// With k=1 nothing is suppressed.
	agg = query.Aggregate(recs, query.AggregateOptions{K: 1})
	if len(agg) != 2 {
		t.Fatalf("k=1: got %d buckets, want 2", len(agg))
	}
}

func TestAggregateBucketing(t *testing.T) {
	var recs []adl.Record
	// 5 records across the same ISO week but different days.
	base := time.Date(2026, 7, 6, 0, 0, 0, 0, time.UTC) // Monday, ISO week 28.
	for i := 0; i < 5; i++ {
		recs = append(recs, mkRecord(traceID(i), "afnemer-a", "doel-x", true, base.AddDate(0, 0, i)))
	}

	byDay := query.Aggregate(recs, query.AggregateOptions{Bucket: query.BucketDay, K: 1})
	if len(byDay) != 5 {
		t.Fatalf("day bucket: got %d, want 5", len(byDay))
	}

	byWeek := query.Aggregate(recs, query.AggregateOptions{Bucket: query.BucketWeek, K: 1})
	if len(byWeek) != 1 {
		t.Fatalf("week bucket: got %d, want 1", len(byWeek))
	}
	if byWeek[0].Periode != "2026-W28" {
		t.Fatalf("week label = %q, want 2026-W28", byWeek[0].Periode)
	}

	byMonth := query.Aggregate(recs, query.AggregateOptions{Bucket: query.BucketMonth, K: 1})
	if len(byMonth) != 1 || byMonth[0].Periode != "2026-07" {
		t.Fatalf("month bucket: got %+v", byMonth)
	}
}

func TestExtractors(t *testing.T) {
	r := mkRecord(traceID(1), "afnemer-a", "opsporing", false,
		time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC))
	if query.Afnemer(&r) != "afnemer-a" {
		t.Errorf("Afnemer = %q", query.Afnemer(&r))
	}
	if query.Doel(&r) != "opsporing" {
		t.Errorf("Doel = %q", query.Doel(&r))
	}
	if d, ok := query.Decision(&r); !ok || d {
		t.Errorf("Decision = %v,%v want false,true", d, ok)
	}
	if query.TransactionID(&r) != "tx-"+traceID(1) {
		t.Errorf("TransactionID = %q", query.TransactionID(&r))
	}
}
