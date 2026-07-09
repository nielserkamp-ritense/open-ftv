package adl

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
)

// traceContext holds the W3C Trace Context identifiers derived for a decision.
type traceContext struct {
	traceID      string // 32 lowercase-hex.
	spanID       string // 16 lowercase-hex (freshly generated for this decision's span).
	parentSpanID string // 16 lowercase-hex, or "" when this decision is the trace root.
}

// deriveTraceContext computes the trace identifiers for a decision from an incoming
// W3C traceparent header value.
//
// Rules (per the ADL standard and W3C Trace Context):
//   - A trace_id received from another organisation MUST be preserved unchanged; a new
//     trace_id MUST NOT be allocated mid-flow. When the incoming traceparent parses, its
//     trace_id is kept and its parent-id becomes this record's parent_span_id.
//   - When no valid trace context is present, a new trace is started with a root span
//     (parent_span_id omitted) and a CSPRNG-generated trace_id.
//   - A fresh span_id is always generated (CSPRNG) for this decision's span.
//   - The sampled bit is never inspected or modified here; logging is independent of it.
func deriveTraceContext(traceparent string) traceContext {
	tc := traceContext{spanID: newSpanID()}

	if tid, pid, ok := parseTraceParent(traceparent); ok {
		tc.traceID = tid
		tc.parentSpanID = pid
		return tc
	}

	tc.traceID = newTraceID()
	return tc
}

// parseTraceParent parses a W3C traceparent header value of the form
// "<version>-<trace-id>-<parent-id>-<flags>" and returns the trace-id and parent-id
// (both validated as lowercase hex of the correct length). ok is false for any
// malformed value or all-zero identifiers.
func parseTraceParent(tp string) (traceID, parentID string, ok bool) {
	parts := strings.Split(strings.TrimSpace(tp), "-")
	if len(parts) != 4 {
		return "", "", false
	}

	traceID = strings.ToLower(parts[1])
	parentID = strings.ToLower(parts[2])

	if !isHex(traceID, 32) || isZero(traceID) {
		return "", "", false
	}
	if !isHex(parentID, 16) || isZero(parentID) {
		return "", "", false
	}

	return traceID, parentID, true
}

// newTraceID returns a 16-byte trace id encoded as 32 lowercase hex characters,
// generated with a cryptographically secure RNG.
func newTraceID() string {
	return randomHex(16)
}

// newSpanID returns an 8-byte span id encoded as 16 lowercase hex characters,
// generated with a cryptographically secure RNG.
func newSpanID() string {
	return randomHex(8)
}

func randomHex(n int) string {
	b := make([]byte, n)
	// crypto/rand.Read never returns an error on the platforms we target; a failure
	// here is unrecoverable, so panicking is preferable to emitting a predictable id.
	if _, err := rand.Read(b); err != nil {
		panic("adl: failed to read from CSPRNG: " + err.Error())
	}
	return hex.EncodeToString(b)
}

func isHex(s string, n int) bool {
	if len(s) != n {
		return false
	}
	_, err := hex.DecodeString(s)
	return err == nil
}

func isZero(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] != '0' {
			return false
		}
	}
	return true
}
