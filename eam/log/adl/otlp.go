package adl

import (
	"context"
	"encoding/json"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/opentelemetry"
)

// otlpSink flushes ADL records to an OpenTelemetry pipeline (OTLP, or any target the
// utilities/opentelemetry logger supports: "stdout", "stderr", "slog", or an OTLP URL).
//
// The standard's "Span attributes" section recommends that, when ADL is carried on an
// OpenTelemetry pipeline, the span's name equals the record's event_name and selected
// adl.core.* attributes are mirrored onto the span. The incoming trace_id is preserved by
// starting the span under a remote parent carrying that trace_id; the record's exact
// span_id / parent_span_id are mirrored as attributes so the ADL identifiers survive.
type otlpSink struct {
	logger opentelemetry.Logger
}

// NewOTLPSink instantiates an ADL sink backed by the OpenTelemetry logger. url may be an
// OTLP endpoint or one of "stdout"/"stderr"/"slog"/"" (see utilities/opentelemetry.New).
func NewOTLPSink(cfg *opentelemetry.LoggerConfig) (Sink, error) {
	// Reject a cleartext http:// OTLP URL unless insecure transport is explicitly allowed.
	// The "stdout"/"stderr"/"slog"/"" targets and bare host:port endpoints are not http://
	// URLs and pass through. Note: the underlying utilities/opentelemetry gRPC exporter
	// dials WithInsecure() (plaintext) - hardening that shared exporter to TLS is out of
	// this component's ownership and remains a documented deployment-level requirement.
	if err := requireSecureEndpoint(cfg.URL); err != nil {
		return nil, err
	}
	logger, err := opentelemetry.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("adl: failed to create OTLP sink: %w", err)
	}
	return &otlpSink{logger: logger}, nil
}

// NewOTLPSinkFromLogger wraps an already-constructed OpenTelemetry logger as a sink.
func NewOTLPSinkFromLogger(logger opentelemetry.Logger) Sink {
	return &otlpSink{logger: logger}
}

// Emit implements the Sink interface.
func (s *otlpSink) Emit(ctx context.Context, rec *Record) error {
	ctx = withRemoteParent(ctx, rec)

	// The span name equals the event_name so decisions are identifiable in tracing tools.
	_, span := s.logger.StartSpan(ctx, rec.EventName)
	defer span.End()

	span.SetAttributes(
		attribute.String("adl.trace_id", rec.TraceID),
		attribute.String("adl.span_id", rec.SpanID),
		attribute.String("adl.event_name", rec.EventName),
		attribute.String(LabelEventName, rec.EventName),
		attribute.String(LabelDecision, DecisionLabel(rec)),
	)
	if sn := ServiceName(rec); sn != "" {
		span.SetAttributes(attribute.String(LabelServiceName, sn))
	}
	if rec.ParentSpanID != "" {
		span.SetAttributes(attribute.String("adl.parent_span_id", rec.ParentSpanID))
	}

	// Carry the complete record as a JSON attribute so no body/attributes field is
	// lost on the tracing path (the trace exporter drops everything not on the span).
	// The logs path (otlpLogSink) is preferred for Loki; this keeps traces lossless too.
	if raw, err := json.Marshal(rec); err == nil {
		span.SetAttributes(attribute.String("adl.record", string(raw)))
	}

	// Mirror the adl.core.policies reference so decisions can be filtered in tracing tools.
	if rec.Attributes != nil {
		if pol, ok := rec.Attributes[KeyPolicies]; ok {
			if raw, err := json.Marshal(pol); err == nil {
				span.SetAttributes(attribute.String(KeyPolicies, string(raw)))
			}
		}
	}

	switch rec.Status {
	case StatusError:
		span.SetStatus(codes.Error, "PDP could not evaluate the request")
	default:
		span.SetStatus(codes.Ok, "")
	}

	return nil
}

// withRemoteParent returns a context carrying a remote parent span context so the emitted
// span inherits the record's trace_id (preserving cross-organisation trace continuity).
func withRemoteParent(ctx context.Context, rec *Record) context.Context {
	tid, err := trace.TraceIDFromHex(rec.TraceID)
	if err != nil {
		return ctx
	}

	cfg := trace.SpanContextConfig{TraceID: tid, Remote: true}
	if rec.ParentSpanID != "" {
		if sid, err := trace.SpanIDFromHex(rec.ParentSpanID); err == nil {
			cfg.SpanID = sid
		}
	}
	return trace.ContextWithRemoteSpanContext(ctx, trace.NewSpanContext(cfg))
}
