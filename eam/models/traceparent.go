package models

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
)

// TraceIDLength is the number of hex characters in a W3C trace-id (16 bytes).
const TraceIDLength = 32

// SpanIDLength is the number of hex characters in a W3C span-id (8 bytes).
const SpanIDLength = 16

// hexCharsPerByte is the number of hex characters needed to represent one byte.
const hexCharsPerByte = 2

// The trailing (-.*)? tolerates fields added by traceparent versions higher than 00, per trace-context §3.2.4:
// trace-id/parent-id/flags always sit at these fixed offsets, and anything after them is parsed but ignored.
var traceParentRX = regexp.MustCompile(`^([0-9a-f]{2})-([0-9a-f]{32})-([0-9a-f]{16})-([0-9a-f]{2})(-.*)?$`)

var invalidTraceID = strings.Repeat("0", TraceIDLength)

var invalidSpanID = strings.Repeat("0", SpanIDLength)

// invalidVersion is the one version value trace-context §3.2.2.1 forbids outright.
const invalidVersion = "ff"

// traceParentSubmatchCount is FindStringSubmatch's result length for traceParentRX: the full match plus
// its 5 capture groups (version, trace-id, span-id, flags, optional trailing fields).
const traceParentSubmatchCount = 6

// TraceContext holds parsed W3C trace-context identifiers.
type TraceContext struct {
	TraceID string
	SpanID  string
	Flags   string // 2 lowercase hex chars, e.g. "01" (sampled) or "00" (not sampled).
}

// ParseTraceParent extracts trace and span IDs from a W3C traceparent value.
func ParseTraceParent(s string) (TraceContext, bool) {
	s = strings.ToLower(strings.TrimSpace(s))

	list := traceParentRX.FindStringSubmatch(s)
	if len(list) != traceParentSubmatchCount {
		return TraceContext{}, false
	}

	version, traceID, spanID, flags, rest := list[1], list[2], list[3], list[4], list[5]
	if version == invalidVersion || traceID == invalidTraceID || spanID == invalidSpanID {
		return TraceContext{}, false
	}

	// Version 00's format is exactly these four fields; trailing fields are only valid for higher versions.
	if version == "00" && rest != "" {
		return TraceContext{}, false
	}

	return TraceContext{TraceID: traceID, SpanID: spanID, Flags: flags}, true
}

// NewSpanID generates a new W3C span-id (16 lowercase hex characters).
func NewSpanID() string {
	return randomHex(SpanIDLength / hexCharsPerByte)
}

// NewTraceParent generates a W3C traceparent header value (version 00, sampled).
func NewTraceParent() (string, TraceContext) {
	traceID := randomHex(TraceIDLength / hexCharsPerByte)
	spanID := randomHex(SpanIDLength / hexCharsPerByte)
	parent := fmt.Sprintf("00-%s-%s-01", traceID, spanID)

	return parent, TraceContext{TraceID: traceID, SpanID: spanID, Flags: "01"}
}

// ResolveTraceParent returns a valid traceparent, preserving the input when valid.
func ResolveTraceParent(existing string) string {
	existing = strings.TrimSpace(existing)
	if existing != "" {
		if _, ok := ParseTraceParent(existing); ok {
			return strings.ToLower(existing)
		}
	}

	parent, _ := NewTraceParent()

	return parent
}

// ResolveOutgoingTraceParent returns a traceparent for an outgoing hop: a new child span_id, same trace_id and sampled flag.
func ResolveOutgoingTraceParent(incoming string) string {
	tc, ok := ParseTraceParent(strings.TrimSpace(incoming))
	if !ok {
		parent, _ := NewTraceParent()
		return parent
	}

	return fmt.Sprintf("00-%s-%s-%s", tc.TraceID, NewSpanID(), tc.Flags)
}

// ResolveTraceState returns tracestate to propagate, dropped when ResolveTraceParent had to start a new trace.
func ResolveTraceState(incomingTraceParent, tracestate string) string {
	if _, ok := ParseTraceParent(strings.TrimSpace(incomingTraceParent)); !ok {
		return ""
	}

	return strings.TrimSpace(tracestate)
}

// FirstHeader returns the first value for a case-insensitive header key.
func FirstHeader(headers map[string][]string, key string) string {
	for k, list := range headers {
		if strings.EqualFold(k, key) && len(list) > 0 && list[0] != "" {
			return list[0]
		}
	}

	return ""
}

// SetHeader sets a header value in a multi-map, replacing an existing key with any casing.
func SetHeader(headers map[string][]string, key, value string) {
	for k := range headers {
		if strings.EqualFold(k, key) {
			headers[k] = []string{value}
			return
		}
	}

	headers[key] = []string{value}
}

func randomHex(byteLen int) string {
	b := make([]byte, byteLen)
	_, _ = rand.Read(b)

	return hex.EncodeToString(b)
}
