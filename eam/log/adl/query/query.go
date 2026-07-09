// Package query provides a read-side over the Authorization Decision Log (ADL).
//
// It reads adl.Record values from a durable source - by default the JSONL
// write-ahead log written by the adl.Logger - and lets callers filter them by
// period, event name, consumer (afnemer/subject), purpose (doel), decision and
// trace id. On top of the filtered records it computes privacy-preserving
// aggregate statistics (see aggregate.go): permit/deny counts grouped per
// (afnemer, doel, period-bucket), with a configurable k-anonymity threshold so
// that small buckets are suppressed and no personal data leaves the aggregate.
//
// The Source interface decouples the query layer from the physical storage, so
// the WAL-backed implementation here can later be replaced by, for example, an
// OpenSearch-backed source without touching the callers.
package query

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/adl"
)

// Source is a read-side over ADL records. Implementations return every record
// that matches the given Filter. The WAL-backed implementation is WALSource;
// an OpenSearch-backed implementation can be added without changing callers.
type Source interface {
	Query(ctx context.Context, f Filter) ([]adl.Record, error)
}

// Filter selects a subset of ADL records. A zero-valued field is not applied,
// so the zero Filter matches every record.
type Filter struct {
	// From and To bound the record timestamp (inclusive From, exclusive To).
	// A zero time means the bound is not applied.
	From time.Time
	To   time.Time

	// EventName, when set, restricts to records with this adl event_name.
	EventName string

	// Afnemer, when set, restricts to records whose subject id equals it.
	Afnemer string

	// Doel, when set, restricts to records whose purpose equals it.
	Doel string

	// Decision, when non-nil, restricts to permit (true) or deny (false) records.
	Decision *bool

	// TraceID, when set, restricts to a single trace.
	TraceID string

	// TraceIDs, when non-empty, restricts to records whose trace_id is in the set.
	TraceIDs []string
}

// Match reports whether the record satisfies every set filter field.
func (f Filter) Match(r *adl.Record) bool {
	if !f.From.IsZero() && recordTime(r).Before(f.From) {
		return false
	}
	if !f.To.IsZero() && !recordTime(r).Before(f.To) {
		return false
	}
	if f.EventName != "" && r.EventName != f.EventName {
		return false
	}
	if f.Afnemer != "" && Afnemer(r) != f.Afnemer {
		return false
	}
	if f.Doel != "" && Doel(r) != f.Doel {
		return false
	}
	if f.Decision != nil {
		d, ok := Decision(r)
		if !ok || d != *f.Decision {
			return false
		}
	}
	if f.TraceID != "" && r.TraceID != f.TraceID {
		return false
	}
	if len(f.TraceIDs) > 0 && !contains(f.TraceIDs, r.TraceID) {
		return false
	}
	return true
}

// WALSource reads ADL records from the JSONL write-ahead log at Path.
type WALSource struct {
	// Path is the file path of the ADL write-ahead log (JSONL, ADL_PATH).
	Path string
	// MaxLine bounds the scanner buffer; defaults to 4 MiB when zero.
	MaxLine int
}

// NewWALSource returns a WALSource reading the write-ahead log at path.
func NewWALSource(path string) *WALSource { return &WALSource{Path: path} }

// Query implements Source. It streams the JSONL write-ahead log, decoding one
// record per line and returning every record that matches the filter. A missing
// file yields no records and no error (the log may simply be empty).
func (s *WALSource) Query(ctx context.Context, f Filter) ([]adl.Record, error) {
	file, err := os.Open(s.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("query: open write-ahead log %q: %w", s.Path, err)
	}
	defer func() { _ = file.Close() }()

	maxLine := s.MaxLine
	if maxLine <= 0 {
		maxLine = 4 * 1024 * 1024
	}

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), maxLine)

	var out []adl.Record
	for scanner.Scan() {
		if err := ctx.Err(); err != nil {
			return out, err
		}
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var rec adl.Record
		if err := json.Unmarshal(line, &rec); err != nil {
			// Skip malformed lines rather than abort the whole query.
			continue
		}
		if f.Match(&rec) {
			out = append(out, rec)
		}
	}
	return out, scanner.Err()
}

// SliceSource is an in-memory Source, primarily useful for tests.
type SliceSource []adl.Record

// Query implements Source.
func (s SliceSource) Query(_ context.Context, f Filter) ([]adl.Record, error) {
	var out []adl.Record
	for i := range s {
		if f.Match(&s[i]) {
			out = append(out, s[i])
		}
	}
	return out, nil
}

// compile-time assertions.
var (
	_ Source = (*WALSource)(nil)
	_ Source = (SliceSource)(nil)
)

// --- record field extractors -------------------------------------------------

// recordTime returns the record timestamp as a time.Time (from Unix milliseconds).
func recordTime(r *adl.Record) time.Time {
	return time.UnixMilli(int64(r.Timestamp)).UTC()
}

// Afnemer returns the consumer identity of the record: the subject id of the
// AuthZEN request. In the FSC/BRP simulation the subject identifies the
// requesting (consumer) organisation, which is how the ADL standard (§5a)
// attributes usage to an afnemer.
func Afnemer(r *adl.Record) string {
	req := asMap(bodyGet(r, adl.KeyRequest))
	return asString(asMap(req["subject"])["id"])
}

// Doel returns the purpose (doelbinding) of the record. Purpose is not a
// first-class ADL field; per the standard it is modelled through the request
// context or action. We look, in order, at request.context.purpose,
// request.context.processing_activity, request.purpose and finally action.name.
func Doel(r *adl.Record) string {
	req := asMap(bodyGet(r, adl.KeyRequest))
	if ctx := asMap(req["context"]); len(ctx) > 0 {
		if p := asString(ctx["purpose"]); p != "" {
			return p
		}
		if p := asString(ctx["processing_activity"]); p != "" {
			return p
		}
	}
	if p := asString(req["purpose"]); p != "" {
		return p
	}
	return asString(asMap(req["action"])["name"])
}

// Decision returns the AuthZEN decision (true=permit, false=deny) and whether a
// decision was present (absent on records with status Error).
func Decision(r *adl.Record) (permit bool, ok bool) {
	resp := asMap(bodyGet(r, adl.KeyResponse))
	if resp == nil {
		return false, false
	}
	d, ok := resp["decision"].(bool)
	return d, ok
}

// TransactionID returns the adl.fsc.transaction_id attribute, if present. It is
// one of the correlation keys (alongside trace_id) for the Inzicht API.
func TransactionID(r *adl.Record) string {
	if r.Attributes == nil {
		return ""
	}
	return asString(r.Attributes[adl.KeyTransactionID])
}

// bodyGet returns r.Body[key] or nil.
func bodyGet(r *adl.Record, key string) any {
	if r.Body == nil {
		return nil
	}
	return r.Body[key]
}

func asMap(v any) map[string]any {
	if m, ok := v.(map[string]any); ok {
		return m
	}
	return nil
}

func asString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func contains(ss []string, s string) bool {
	for _, x := range ss {
		if x == s {
			return true
		}
	}
	return false
}
