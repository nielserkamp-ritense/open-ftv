package decisions

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

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
		params := decisionToParms(decisionFromSpan(span.StartTime().UTC(), span.Attributes()))
		if _, err = t.Exec(ctx, sql, params...); err != nil {
			return
		}
	}

	return nil
}

const sql = "INSERT INTO decision (created,timestamp,trace_id,span_id,parent_span_id,event_name,status,request_type,policies,body,attributes,information,engine) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)"

func decisionFromSpan(timestamp time.Time, attrs []attribute.KeyValue) *Decision {
	d := Decision{Timestamp: timestamp, Status: StatusUnset}

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
		case "information":
			d.Information = []byte(item.Value.AsString())
		case "engine":
			d.Engine = []byte(item.Value.AsString())
		}
	}

	if d.EventName == "" && d.RequestType != 0 {
		d.EventName = d.RequestType.EventName()
	}
	if d.Status == "" {
		d.Status = StatusUnset
	}

	return &d
}

func decisionToParms(d *Decision) []any {
	if d.Status == "" {
		d.Status = StatusUnset
	}
	if d.EventName == "" && d.RequestType != 0 {
		d.EventName = d.RequestType.EventName()
	}

	params := []any{
		d.Timestamp,
		int64(d.Timestamp.UnixMilli()),
		nullableHex(d.TraceID),
		nullableHex(d.SpanID),
		nullableHex(d.ParentSpanID),
		d.EventName,
		string(d.Status),
		int64(d.RequestType),
		int64(d.Policies),
	}

	if body := BodyFromDecision(d); body != nil {
		params = append(params, body, DefaultADLAttributes())
	} else {
		params = append(params, nil, nil)
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
