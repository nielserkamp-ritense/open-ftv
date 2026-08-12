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

var traceParentRX = regexp.MustCompile(`^([0-9a-f]{2})-([0-9a-f]{32})-([0-9a-f]{16})-([0-9a-f]{2})$`)

var invalidTraceID = strings.Repeat("0", TraceIDLength)

// TraceContext holds parsed W3C trace-context identifiers.
type TraceContext struct {
	TraceID string
	SpanID  string
}

// ParseTraceParent extracts trace and span IDs from a W3C traceparent value.
func ParseTraceParent(s string) (TraceContext, bool) {
	s = strings.ToLower(strings.TrimSpace(s))
	list := traceParentRX.FindStringSubmatch(s)
	if len(list) != 5 || list[1] != "00" || list[2] == invalidTraceID {
		return TraceContext{}, false
	}
	return TraceContext{TraceID: list[2], SpanID: list[3]}, true
}

// NewSpanID generates a new W3C span-id (16 lowercase hex characters).
func NewSpanID() string {
	return randomHex(SpanIDLength / 2)
}

// NewTraceParent generates a W3C traceparent header value (version 00, sampled).
func NewTraceParent() (string, TraceContext) {
	traceID := randomHex(TraceIDLength / 2)
	spanID := randomHex(SpanIDLength / 2)
	parent := fmt.Sprintf("00-%s-%s-01", traceID, spanID)
	return parent, TraceContext{TraceID: traceID, SpanID: spanID}
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
