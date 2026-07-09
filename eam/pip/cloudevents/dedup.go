package cloudevents

import (
	"strings"
	"sync"
)

// Dedup provides idempotent ingest on the (source, id) tuple of CloudEvents.
//
// It keeps a bounded in-memory set of recently seen keys; once the limit is
// reached the oldest half is discarded (coarse-grained, allocation-free aging).
// A Dedup is safe for concurrent use.
type Dedup struct {
	limit int
	mutex sync.Mutex
	old   map[string]struct{}
	cur   map[string]struct{}
}

// NewDedup creates a Dedup that remembers at least limit/2 and at most limit keys.
func NewDedup(limit int) *Dedup {
	if limit <= 0 {
		limit = 4096
	}
	return &Dedup{limit: limit, cur: make(map[string]struct{}, 64)}
}

// Seen records the event's (source, id) and reports whether it was already ingested.
func (d *Dedup) Seen(e *Event) bool {
	key := e.DedupKey()

	d.mutex.Lock()
	defer d.mutex.Unlock()

	if _, ok := d.cur[key]; ok {
		return true
	}
	if _, ok := d.old[key]; ok {
		return true
	}

	if len(d.cur) >= d.limit/2 {
		d.old = d.cur
		d.cur = make(map[string]struct{}, 64)
	}
	d.cur[key] = struct{}{}
	return false
}

// ParseTraceParent extracts the trace-id and parent span-id from a W3C trace-context
// "traceparent" header value ("00-<32 hex>-<16 hex>-<2 hex>"). ok is false when the
// value does not follow that layout.
func ParseTraceParent(value string) (traceID, parentID string, ok bool) {
	parts := strings.Split(strings.TrimSpace(value), "-")
	if len(parts) < 4 || len(parts[1]) != 32 || len(parts[2]) != 16 {
		return "", "", false
	}
	return strings.ToLower(parts[1]), strings.ToLower(parts[2]), true
}
