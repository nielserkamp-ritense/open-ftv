package models

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTraceParent(t *testing.T) {
	t.Parallel()

	tc, ok := ParseTraceParent("00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")
	require.True(t, ok)
	assert.Equal(t, "4bf92f3577b34da6a3ce929d0e0e4736", tc.TraceID)
	assert.Equal(t, "00f067aa0ba902b7", tc.SpanID)
	assert.Equal(t, "01", tc.Flags)

	tc, ok = ParseTraceParent("00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-00")
	require.True(t, ok)
	assert.Equal(t, "00", tc.Flags)

	_, ok = ParseTraceParent("00-00000000000000000000000000000000-00f067aa0ba902b7-01")
	assert.False(t, ok)

	_, ok = ParseTraceParent("00-4bf92f3577b34da6a3ce929d0e0e4736-0000000000000000-01")
	assert.False(t, ok)

	_, ok = ParseTraceParent("invalid")
	assert.False(t, ok)

	// Higher versions add fields after trace-flags; those must be parsed for trace-id/span-id/flags and the
	// rest ignored, not rejected outright (trace-context §3.2.4).
	tc, ok = ParseTraceParent("01-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01-extra-fields")
	require.True(t, ok)
	assert.Equal(t, "4bf92f3577b34da6a3ce929d0e0e4736", tc.TraceID)
	assert.Equal(t, "00f067aa0ba902b7", tc.SpanID)
	assert.Equal(t, "01", tc.Flags)

	// Version 00's format is exactly four fields; trailing fields there are invalid, not forward-compatible.
	_, ok = ParseTraceParent("00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01-extra")
	assert.False(t, ok)

	// ff is explicitly forbidden as a version.
	_, ok = ParseTraceParent("ff-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")
	assert.False(t, ok)
}

func TestResolveTraceParent(t *testing.T) {
	t.Parallel()

	const valid = "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"
	assert.Equal(t, valid, ResolveTraceParent(valid))
	assert.Equal(t, valid, ResolveTraceParent("  "+valid+"  "))

	generated := ResolveTraceParent("")
	_, ok := ParseTraceParent(generated)
	assert.True(t, ok)

	generated = ResolveTraceParent("not-a-traceparent")
	_, ok = ParseTraceParent(generated)
	assert.True(t, ok)
}

func TestResolveOutgoingTraceParent(t *testing.T) {
	t.Parallel()

	const incoming = "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-00"

	outgoing := ResolveOutgoingTraceParent(incoming)

	tc, ok := ParseTraceParent(outgoing)
	require.True(t, ok)
	assert.Equal(t, "4bf92f3577b34da6a3ce929d0e0e4736", tc.TraceID)
	assert.Equal(t, "00", tc.Flags, "sampled flag MUST NOT be modified")
	assert.NotEqual(t, "00f067aa0ba902b7", tc.SpanID, "outgoing hop MUST start a new child span")

	for _, bad := range []string{"", "not-a-traceparent"} {
		outgoing = ResolveOutgoingTraceParent(bad)
		_, ok = ParseTraceParent(outgoing)
		assert.True(t, ok)
	}

	// A higher-version incoming traceparent MUST NOT force a new trace_id mid-flow; the outgoing header is
	// downgraded to version 00 but keeps the same trace_id (trace-context §3.4 "Downgrade the version").
	outgoing = ResolveOutgoingTraceParent("01-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01-extra-fields")
	tc, ok = ParseTraceParent(outgoing)
	require.True(t, ok)
	assert.Equal(t, "4bf92f3577b34da6a3ce929d0e0e4736", tc.TraceID)
	assert.True(t, strings.HasPrefix(outgoing, "00-"))
}

func TestResolveTraceState(t *testing.T) {
	t.Parallel()

	const valid = "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"
	assert.Equal(t, "vendorA=1,vendorB=2", ResolveTraceState(valid, "  vendorA=1,vendorB=2  "))
	assert.Equal(t, "", ResolveTraceState(valid, ""))

	assert.Equal(t, "", ResolveTraceState("", "vendorA=1"))
	assert.Equal(t, "", ResolveTraceState("not-a-traceparent", "vendorA=1"))
}

func TestNewTraceParent(t *testing.T) {
	t.Parallel()

	parent, tc := NewTraceParent()
	_, ok := ParseTraceParent(parent)
	require.True(t, ok)
	assert.Len(t, tc.TraceID, TraceIDLength)
	assert.Len(t, tc.SpanID, SpanIDLength)
	assert.Equal(t, "01", tc.Flags)
}

func TestHeaderHelpers(t *testing.T) {
	t.Parallel()

	headers := map[string][]string{"TraceParent": {"00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"}}
	assert.Equal(t, "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01", FirstHeader(headers, HeaderTraceParent))

	SetHeader(headers, HeaderTraceParent, "00-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa-bbbbbbbbbbbbbbbb-01")
	assert.Equal(t, "00-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa-bbbbbbbbbbbbbbbb-01", FirstHeader(headers, "traceparent"))
}
