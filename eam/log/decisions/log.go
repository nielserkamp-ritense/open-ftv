package decisions

import (
	"context"
	"fmt"
	"time"

	"github.com/goccy/go-json"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/opentelemetry"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/postgresql/pool"
)

// New instantiates a new OpenTelemetry logger for the Authorization Decision Log.
func New(ctx context.Context, service string, opts ...opentelemetry.Option) (*Logger, error) {
	ot, err := opentelemetry.New(ctx, service, opts...)
	if err != nil {
		return nil, err
	}
	return &Logger{ot: ot}, nil
}

// NewWithPostgreSQL instantiates a new OpenTelemetry logger for the Authorization Decision Log with a PostgreSQL backend.
func NewWithPostgreSQL(ctx context.Context, service string, dsn string, p pool.Pooler, maxLife time.Duration, maxConn int32, opts ...opentelemetry.Option) (*Logger, error) {
	if dsn == "" {
		return nil, fmt.Errorf("dsn must be filled")
	}

	e, err := newPostgreSQL(dsn, p, maxLife, maxConn)
	if err != nil {
		return nil, err
	}

	opts = append(opts, opentelemetry.WithExporter(e))

	ot, err2 := opentelemetry.New(ctx, service, opts...)
	if err2 != nil {
		return nil, err2
	}
	return &Logger{ot: ot}, nil
}

// Shutdown cleans up held resources.
//
// Do not call Decision() after Shutdown() has been called.
func (l *Logger) Shutdown(ctx context.Context) error {
	return l.ot.Shutdown(ctx)
}

// Decision writes an authorization decision to the log.
func (l *Logger) Decision(ctx context.Context, d *Decision) error {
	if d.EventName == "" && d.RequestType != 0 {
		d.EventName = d.RequestType.EventName()
	}

	if d.Status == "" {
		d.Status = StatusUnset
	}

	// Per Logius ADL §3.4, the span and the log record are correlated via a
	// shared trace_id/span_id — the span's own identity, not just an
	// attribute on it, so that an OTLP-transported span (PDP_DECISIONLOG_TYPE=otel)
	// actually correlates downstream.
	ctx = opentelemetry.ContextWithSpanIDs(ctx, d.TraceID, d.SpanID)

	_, span := l.ot.StartSpan(ctx, d.EventName, trace.WithTimestamp(d.Timestamp))

	span.SetAttributes(
		attribute.Int64("request_type", int64(d.RequestType)),
		attribute.Int64("policies", int64(d.Policies)),
		attribute.String("event_name", d.EventName),
		attribute.String("status", string(d.Status)),
	)

	if d.TraceID != "" {
		span.SetAttributes(attribute.String("trace_id", d.TraceID))
	}
	if d.SpanID != "" {
		span.SetAttributes(attribute.String("span_id", d.SpanID))
	}

	if d.ParentSpanID != "" {
		span.SetAttributes(attribute.String("parent_span_id", d.ParentSpanID))
	}

	if d.Request != nil || d.Response != nil {
		body, err := formatAny(BodyFromDecision(d))
		if err != nil {
			return err
		}
		// Level 1: request/response live in body, so attributes (source
		// references) stays empty per the Logius ADL spec.
		span.SetAttributes(
			attribute.String("body", body),
			attribute.String("attributes", "{}"),
		)
	}

	if d.Information != nil {
		info, err := formatAny(d.Information)
		if err != nil {
			return err
		}

		span.SetAttributes(attribute.String("information", info))
	}

	if d.Engine != nil {
		engine, err := formatAny(d.Engine)
		if err != nil {
			return err
		}

		span.SetAttributes(attribute.String("engine", engine))
	}

	if d.Resource != nil {
		resource, err := formatAny(d.Resource)
		if err != nil {
			return err
		}

		span.SetAttributes(attribute.String("resource", resource))
	}

	span.End()

	return nil
}

func formatAny(in any) (string, error) {
	if b, ok := in.([]byte); ok {
		return string(b), nil
	}

	b, err := json.Marshal(in)
	return string(b), err
}

// Logger implements an OpenTelemetry logger for the Authorization Decision Log.
type Logger struct {
	ot opentelemetry.Logger
}
