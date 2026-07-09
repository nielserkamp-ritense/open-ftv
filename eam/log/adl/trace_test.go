package adl

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	hex32 = regexp.MustCompile(`^[0-9a-f]{32}$`)
	hex16 = regexp.MustCompile(`^[0-9a-f]{16}$`)
)

func TestDeriveTraceContext_IncomingTraceParent(t *testing.T) {
	t.Parallel()

	tc := deriveTraceContext("00-28dbeec32e77635cc19bc3204ec56c41-893e1b2ac52d712f-01")

	// An externally received trace_id MUST be preserved unchanged.
	assert.Equal(t, "28dbeec32e77635cc19bc3204ec56c41", tc.traceID)
	// The incoming parent-id becomes this record's parent_span_id.
	assert.Equal(t, "893e1b2ac52d712f", tc.parentSpanID)
	// A fresh span id is always generated.
	assert.Regexp(t, hex16, tc.spanID)
	assert.NotEqual(t, tc.parentSpanID, tc.spanID)
}

func TestDeriveTraceContext_NoTraceParent(t *testing.T) {
	t.Parallel()

	tc := deriveTraceContext("")

	assert.Regexp(t, hex32, tc.traceID)
	assert.Regexp(t, hex16, tc.spanID)
	// Root of a new trace: parent_span_id is omitted.
	assert.Empty(t, tc.parentSpanID)
}

func TestDeriveTraceContext_MalformedTraceParent(t *testing.T) {
	t.Parallel()

	for _, tp := range []string{
		"garbage",
		"00-zzzzeec32e77635cc19bc3204ec56c41-893e1b2ac52d712f-01", // non-hex trace id.
		"00-28dbeec32e77635cc19bc3204ec56c41-1234-01",             // short parent id.
		"00-00000000000000000000000000000000-893e1b2ac52d712f-01", // all-zero trace id.
		"00-28dbeec32e77635cc19bc3204ec56c41-0000000000000000-01", // all-zero parent id.
	} {
		tc := deriveTraceContext(tp)
		assert.Regexp(t, hex32, tc.traceID, "input %q", tp)
		assert.NotEqual(t, "28dbeec32e77635cc19bc3204ec56c41", tc.traceID, "input %q must start a new trace", tp)
		assert.Empty(t, tc.parentSpanID, "input %q", tp)
	}
}

func TestNewIDsAreRandom(t *testing.T) {
	t.Parallel()

	seen := make(map[string]struct{})
	for i := 0; i < 100; i++ {
		tid, sid := newTraceID(), newSpanID()
		assert.Regexp(t, hex32, tid)
		assert.Regexp(t, hex16, sid)
		seen[tid] = struct{}{}
		seen[sid] = struct{}{}
	}
	assert.Len(t, seen, 200, "generated ids must not repeat")
}
