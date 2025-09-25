package decisions

import (
	"context"
	"encoding/hex"
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

const sql = "INSERT INTO decision (created,trace_id,span_id,request_type,policies,request,response,information,engine) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)"

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
		case "request":
			d.Request = []byte(item.Value.AsString())
		case "response":
			d.Response = []byte(item.Value.AsString())
		case "information":
			d.Information = []byte(item.Value.AsString())
		case "engine":
			d.Engine = []byte(item.Value.AsString())
		}
	}

	return &d
}

func decisionToParms(d *Decision) []any {
	traceID2, _ := hex.DecodeString(d.TraceID)
	spanID2, _ := hex.DecodeString(d.SpanID)

	params := []any{
		d.Timestamp,
		traceID2,
		spanID2,
		int64(d.RequestType),
		int64(d.Policies),
	}

	if d.Request != nil {
		params = append(params, d.Request)
	} else {
		params = append(params, nil)
	}
	if d.Response != nil {
		params = append(params, d.Response)
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

	return params
}

// Shutdown implements the SpanSyncer interface.
func (pl *pgLogger) Shutdown(_ context.Context) error {
	return nil
}

type pgLogger struct {
	db pool.Pooler
}
