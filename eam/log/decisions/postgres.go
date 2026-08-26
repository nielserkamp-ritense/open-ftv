package decisions

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/postgresql/pool"
)

// A PostgreSQL logger is designed specifically for the Authorization Decision Log for OpenFTV.
func newPostgreSQL(dsn string, db pool.Pooler, maxLife time.Duration, maxConn int32) (*pgLogger, error) {
	opts := make([]pool.Option, 0, 2)
	if maxLife > 0 {
		opts = append(opts, pool.WithMaxLifetime(maxLife))
	}
	if maxConn > 0 {
		opts = append(opts, pool.WithMaxConnections(maxConn))
	}

	var p pool.Pooler
	var err error

	if db == nil {
		p, err = pool.NewPool(context.Background(), dsn, opts...)
	} else {
		p, err = pool.NewWithPooler(context.Background(), dsn, db, opts...)
	}

	if err != nil {
		return nil, err
	}
	return &pgLogger{db: p}, nil
}

// ExportSpans implements the SpanSyncer interface.
func (pl *pgLogger) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) (err error) {
	var t pgx.Tx
	if t, err = pl.db.Begin(ctx); err != nil {
		return
	}

	defer func() {
		if err == nil {
			err = t.Commit(ctx)
		} else {
			err = errors.Join(err, t.Rollback(ctx))
		}
	}()

	for i := range spans {
		span := spans[i]
		d := decisionFromSpan(span.StartTime().UTC(), span.Attributes())
		fallbackSpanIDs(d, span.SpanContext())

		if _, err = t.Exec(ctx, sql, decisionToParms(d)...); err != nil {
			return
		}
	}

	return nil
}

const sql = "INSERT INTO decision (timestamp,trace_id,span_id,parent_span_id,event_name,status,request_type,policies,body,attributes,information,engine,resource) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13) ON CONFLICT (trace_id, span_id) DO NOTHING"

func decisionFromSpan(timestamp time.Time, attrs []attribute.KeyValue) *Decision {
	d := Decision{Timestamp: timestamp}

	for _, item := range attrs {
		switch strings.ToLower(string(item.Key)) {
		case "request_type":
			d.RequestType = AuthRequestType(item.Value.AsInt64())
		case "policies":
			d.Policies = uint64(item.Value.AsInt64())
		case "trace_id":
			d.TraceID = item.Value.AsString()
		case "span_id":
			d.SpanID = item.Value.AsString()
		case "parent_span_id":
			d.ParentSpanID = item.Value.AsString()
		case "event_name":
			d.EventName = item.Value.AsString()
		case "status":
			d.Status = Status(item.Value.AsString())
		case "body":
			applyBodyAttribute(item.Value.AsString(), &d)
		case "attributes":
			applyAttributesAttribute(item.Value.AsString(), &d)
		case "information":
			d.Information = []byte(item.Value.AsString())
		case "engine":
			d.Engine = []byte(item.Value.AsString())
		case "resource":
			d.Resource = []byte(item.Value.AsString())
		}
	}

	d.applyDefaults()

	return &d
}

// fallbackSpanIDs fills in trace_id/span_id from the span's own identity when missing.
// Both are mandatory (Logius ADL §3.3.1/§3.3.2) and should always arrive as attributes handled by decisionFromSpan above.
func fallbackSpanIDs(d *Decision, sc trace.SpanContext) {
	if d.TraceID == "" {
		if tid := sc.TraceID(); tid.IsValid() {
			d.TraceID = tid.String()
		}
	}

	if d.SpanID == "" {
		if sid := sc.SpanID(); sid.IsValid() {
			d.SpanID = sid.String()
		}
	}
}

func decisionToParms(d *Decision) []any {
	d.applyDefaults()

	params := []any{
		d.Timestamp.UnixMilli(),
		nullableHex(d.TraceID),
		nullableHex(d.SpanID),
		nullableHex(d.ParentSpanID),
		d.EventName,
		string(d.Status),
		int64(d.RequestType),
		int64(d.Policies),
	}

	body := BodyFromDecision(d)
	attrs := AttributesFromDecision(d)

	if body != nil {
		params = append(params, body)
	} else {
		params = append(params, nil)
	}

	// attributes (e.g. adl.fsc.transaction_id, §3.3.7.6) must persist even without a body to go with it.
	if body != nil || len(attrs) > 0 {
		params = append(params, attrs)
	} else {
		params = append(params, nil)
	}

	if d.Information != nil {
		params = append(params, d.Information)
	} else {
		params = append(params, nil)
	}

	if d.Engine != nil {
		params = append(params, d.Engine)
	} else {
		params = append(params, nil)
	}

	if d.Resource != nil {
		params = append(params, d.Resource)
	} else {
		params = append(params, nil)
	}

	return params
}

func nullableHex(s string) any {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return nil
	}

	return s
}

// Shutdown implements the SpanSyncer interface.
func (pl *pgLogger) Shutdown(_ context.Context) error {
	return nil
}

type pgLogger struct {
	db pool.Pooler
}
