package models

import (
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

	_, ok = ParseTraceParent("00-00000000000000000000000000000000-00f067aa0ba902b7-01")
	assert.False(t, ok)

	_, ok = ParseTraceParent("00-4bf92f3577b34da6a3ce929d0e0e4736-0000000000000000-01")
	assert.False(t, ok)

	_, ok = ParseTraceParent("invalid")
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

func TestNewTraceParent(t *testing.T) {
	t.Parallel()

	parent, tc := NewTraceParent()
	_, ok := ParseTraceParent(parent)
	require.True(t, ok)
	assert.Len(t, tc.TraceID, TraceIDLength)
	assert.Len(t, tc.SpanID, SpanIDLength)
}

func TestHeaderHelpers(t *testing.T) {
	t.Parallel()

	headers := map[string][]string{"TraceParent": {"00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"}}
	assert.Equal(t, "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01", FirstHeader(headers, HeaderTraceParent))

	SetHeader(headers, HeaderTraceParent, "00-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa-bbbbbbbbbbbbbbbb-01")
	assert.Equal(t, "00-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa-bbbbbbbbbbbbbbbb-01", FirstHeader(headers, "traceparent"))
}
