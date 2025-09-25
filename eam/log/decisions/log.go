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
	_, span := l.ot.StartSpan(ctx, "test", trace.WithTimestamp(d.Timestamp))

	span.SetAttributes(
		attribute.Int64("request_type", int64(d.RequestType)),
		attribute.Int64("policies", int64(d.Policies)),
	)

	if d.TraceID != "" {
		span.SetAttributes(attribute.String("trace_id", d.TraceID))
	}
	if d.SpanID != "" {
		span.SetAttributes(attribute.String("span_id", d.SpanID))
	}

	if d.Request != nil {
		request, err := formatAny(d.Request)
		if err != nil {
			return err
		}
		span.SetAttributes(attribute.String("request", request))
	}

	if d.Response != nil {
		response, err := formatAny(d.Response)
		if err != nil {
			return err
		}
		span.SetAttributes(attribute.String("response", response))
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
