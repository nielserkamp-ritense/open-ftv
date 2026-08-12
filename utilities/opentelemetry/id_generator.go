package opentelemetry

import (
	"context"
	"crypto/rand"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

type spanIDsKey struct{}

type spanIDs struct {
	traceID trace.TraceID
	spanID  trace.SpanID
}

// ContextWithSpanIDs returns a context that forces the next span started
// against it to use the given W3C trace/span IDs (lowercase hex, 32/16
// characters), instead of SDK-generated ones.
//
// This is what lets an emitted span's own identity match an externally
// determined trace/span ID — e.g. the Logius ADL trace_id/span_id fields,
// which per ADL §3.4 must correlate with the span carrying the record, not
// merely be attached to it as an attribute. Invalid or empty IDs leave ctx
// unchanged, so the tracer falls back to its normal random generation.
func ContextWithSpanIDs(ctx context.Context, traceIDHex, spanIDHex string) context.Context {
	tid, err := trace.TraceIDFromHex(traceIDHex)
	if err != nil || !tid.IsValid() {
		return ctx
	}

	sid, err := trace.SpanIDFromHex(spanIDHex)
	if err != nil || !sid.IsValid() {
		return ctx
	}

	return context.WithValue(ctx, spanIDsKey{}, spanIDs{traceID: tid, spanID: sid})
}

// idGenerator is a sdktrace.IDGenerator that uses the trace/span IDs stashed in ctx by ContextWithSpanIDs when present,
// and otherwise generates random ones.
type idGenerator struct{}

var _ sdktrace.IDGenerator = idGenerator{}

// NewIDs implements the sdktrace.IDGenerator interface.
func (idGenerator) NewIDs(ctx context.Context) (trace.TraceID, trace.SpanID) {
	if ids, ok := ctx.Value(spanIDsKey{}).(spanIDs); ok {
		return ids.traceID, ids.spanID
	}

	var (
		tid trace.TraceID
		sid trace.SpanID
	)

	_, _ = rand.Read(tid[:])
	_, _ = rand.Read(sid[:])

	return tid, sid
}

// NewSpanID implements the sdktrace.IDGenerator interface.
func (idGenerator) NewSpanID(ctx context.Context, _ trace.TraceID) trace.SpanID {
	if ids, ok := ctx.Value(spanIDsKey{}).(spanIDs); ok {
		return ids.spanID
	}

	var sid trace.SpanID

	_, _ = rand.Read(sid[:])

	return sid
}
