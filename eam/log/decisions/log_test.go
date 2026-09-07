package decisions

import (
	"context"
	"encoding/hex"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/opentelemetry"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

// slowExporter wraps a real exporter but sleeps first, so a caller that returns before delay has
// elapsed proves it didn't actually wait for the export to complete.
type slowExporter struct {
	sdktrace.SpanExporter
	delay time.Duration
}

func (s *slowExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	time.Sleep(s.delay)
	return s.SpanExporter.ExportSpans(ctx, spans)
}

func TestNew(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		service string
		opts    []opentelemetry.Option
		wantErr bool
	}{
		{name: "no service name", wantErr: true},
		{name: "no exporter", service: "test1", wantErr: true},
		{name: "good", service: "test2", opts: []opentelemetry.Option{opentelemetry.WithFile(os.Stdout, true)}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			got, err := New(ctx, tc.service, tc.opts...)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)

				err = got.Shutdown(ctx)
				require.NoError(t, err)
			}
		})
	}
}

func TestNewWithPostgreSQL(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		service string
		dsn     string
		maxLife time.Duration
		maxConn int32
		opts    []opentelemetry.Option
		wantErr bool
	}{
		{name: "no service name", dsn: "postgres://localhost:5432/mydb", wantErr: true},
		{name: "no dsn", service: "test1", wantErr: true},
		{name: "bad dsn", service: "test2", dsn: "ho ho", wantErr: true},
		{name: "good", service: "test3", dsn: "postgres://localhost:5432/mydb", maxLife: time.Second, maxConn: 5},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			got, err := NewWithPostgreSQL(ctx, tc.service, tc.dsn, nil, tc.maxLife, tc.maxConn, tc.opts...)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)

				err = got.Shutdown(ctx)
				require.NoError(t, err)
			}
		})
	}
}

func TestLogger_Decision(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	ch := make(chan bool)

	testCases := []struct {
		name    string
		d       *Decision
		wantErr bool
	}{
		{
			name: "only basics",
			d:    &Decision{Timestamp: now, RequestType: EvaluationEndpoint, Policies: 1},
		},
		{
			name: "with trace_id",
			d:    &Decision{Timestamp: now, RequestType: EvaluationEndpoint, Policies: 1, TraceID: "1234567812345678"},
		},
		{
			name: "with span_id",
			d:    &Decision{Timestamp: now, RequestType: EvaluationEndpoint, Policies: 1, SpanID: "12345678"},
		},
		{
			name:    "bad request",
			d:       &Decision{Timestamp: now, RequestType: EvaluationEndpoint, Policies: 1, Request: ch},
			wantErr: true,
		},
		{
			name: "good request",
			d:    &Decision{Timestamp: now, RequestType: EvaluationEndpoint, Policies: 1, Request: "hello"},
		},
		{
			name:    "bad response",
			d:       &Decision{Timestamp: now, RequestType: EvaluationEndpoint, Policies: 1, Response: ch},
			wantErr: true,
		},
		{
			name: "good response",
			d:    &Decision{Timestamp: now, RequestType: EvaluationEndpoint, Policies: 1, Response: "world"},
		},
		{
			name:    "bad information",
			d:       &Decision{Timestamp: now, RequestType: EvaluationEndpoint, Policies: 1, Information: ch},
			wantErr: true,
		},
		{
			name: "good information",
			d:    &Decision{Timestamp: now, RequestType: EvaluationEndpoint, Policies: 1, Information: "all good"},
		},
		{
			name:    "bad engine",
			d:       &Decision{Timestamp: now, RequestType: EvaluationEndpoint, Policies: 1, Engine: ch},
			wantErr: true,
		},
		{
			name: "good engine",
			d:    &Decision{Timestamp: now, RequestType: EvaluationEndpoint, Policies: 1, Engine: "cedar v1.0"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := slog2.NewDummyHandler(slog.LevelDebug)
			logger := slog.New(h)

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			got, err := New(ctx, "openftv", opentelemetry.WithSLog(logger, "openftv"), opentelemetry.WithBatchTimeout(10*time.Millisecond))
			require.NoError(t, err)
			require.NotNil(t, got)

			err = got.Decision(ctx, tc.d)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			time.Sleep(50 * time.Millisecond)

			err = got.Shutdown(ctx)
			require.NoError(t, err)

			if !tc.wantErr {
				assert.Equal(t, 1, h.Count())
			}
		})
	}
}

func TestLogger_Decision_ParentLinkage(t *testing.T) {
	t.Parallel()

	const (
		traceID      = "12345678123456781234567812345678"
		spanID       = "1234567812345678"
		parentSpanID = "8765432187654321"
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	exporter := tracetest.NewInMemoryExporter()
	l, err := New(ctx, "openftv", opentelemetry.WithExporter(exporter), opentelemetry.WithBatchTimeout(10*time.Millisecond))
	require.NoError(t, err)

	require.NoError(t, l.Decision(ctx, &Decision{
		Timestamp:   time.Now().UTC(),
		RequestType: EvaluationEndpoint,
		TraceID:     traceID,
		SpanID:      spanID,
	}))
	require.NoError(t, l.Decision(ctx, &Decision{
		Timestamp:    time.Now().UTC(),
		RequestType:  EvaluationEndpoint,
		TraceID:      traceID,
		SpanID:       spanID,
		ParentSpanID: parentSpanID,
	}))

	spans := exporter.GetSpans()
	require.Len(t, spans, 2)

	require.NoError(t, l.Shutdown(ctx))

	// No parent_span_id: the SDK-level span is correctly a root (as before).
	assert.False(t, spans[0].Parent.IsValid())

	// With parent_span_id: the SDK-level span now links to it directly, not just via a "parent_span_id"
	// attribute — this is what OTLP export relies on to render the trace tree.
	assert.True(t, spans[1].Parent.IsValid())
	assert.True(t, spans[1].Parent.IsRemote())
	assert.Equal(t, traceID, spans[1].Parent.TraceID().String())
	assert.Equal(t, parentSpanID, spans[1].Parent.SpanID().String())
	assert.Equal(t, spanID, spans[1].SpanContext.SpanID().String())
	assert.True(t, spans[1].SpanContext.IsSampled(), "ADL records must be recorded regardless of the sampled bit")
}

// TestLogger_Decision_SpanName covers §3.4 (SHOULD): the span's name equals the record's event_name.
func TestLogger_Decision_SpanName(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	exporter := tracetest.NewInMemoryExporter()
	l, err := New(ctx, "openftv", opentelemetry.WithExporter(exporter), opentelemetry.WithBatchTimeout(10*time.Millisecond))
	require.NoError(t, err)

	require.NoError(t, l.Decision(ctx, &Decision{Timestamp: time.Now().UTC(), RequestType: SearchActionEndpoint}))

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)
	require.NoError(t, l.Shutdown(ctx))

	assert.Equal(t, "adl.search_action", spans[0].Name)
}

func TestLogger_Decision_AttributesWithoutBody(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	exporter := tracetest.NewInMemoryExporter()
	l, err := New(ctx, "openftv", opentelemetry.WithExporter(exporter), opentelemetry.WithBatchTimeout(10*time.Millisecond))
	require.NoError(t, err)

	require.NoError(t, l.Decision(ctx, &Decision{
		Timestamp:        time.Now().UTC(),
		RequestType:      EvaluationEndpoint,
		FSCTransactionID: "abc-123",
	}))

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)
	require.NoError(t, l.Shutdown(ctx))

	var gotAttrs string

	var hasAttrs, hasBody bool

	for _, kv := range spans[0].Attributes {
		switch kv.Key {
		case "attributes":
			hasAttrs, gotAttrs = true, kv.Value.AsString()
		case "body":
			hasBody = true
		}
	}

	assert.True(t, hasAttrs, "adl.fsc.transaction_id must not ride on the body condition (§3.3.7.6)")
	assert.JSONEq(t, `{"adl.fsc.transaction_id":"abc-123"}`, gotAttrs)
	assert.False(t, hasBody)
}

func TestPgLogger_Decision(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		count int
	}{
		{name: "single", count: 1},
		{name: "few", count: 5},
		{name: "many", count: 99},
	}

	now := time.Now().UTC()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			traceID := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
			spanID := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}

			traceID2 := hex.EncodeToString(traceID)
			spanID2 := hex.EncodeToString(spanID)

			mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
			require.NoError(t, err)
			defer mock.Close()

			// Decision() now force-flushes before returning, so every call commits in its own
			// transaction instead of all tc.count decisions sharing one batched transaction.
			for i := range tc.count {
				mock.ExpectBegin()
				mock.ExpectExec(sql).
					WithArgs(
						now.UnixMilli(),
						traceID2, spanID2, nil,
						"adl.access_evaluation", string(StatusUnset),
						int64(1), int64(i+1),
						nil, nil, nil, nil, nil,
					).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				mock.ExpectCommit()
			}

			l, err2 := NewWithPostgreSQL(
				nil,
				"openftv",
				"postgres://localhost:5432/auth",
				mock,
				5*time.Minute,
				5,
				opentelemetry.WithBatchTimeout(time.Second),
			)
			require.NoError(t, err2)
			require.NotNil(t, l)

			for i := range tc.count {
				err = l.Decision(ctx, &Decision{
					Timestamp:   now,
					RequestType: EvaluationEndpoint,
					EventName:   EvaluationEndpoint.EventName(),
					Status:      StatusUnset,
					Policies:    uint64(i + 1),
					TraceID:     traceID2,
					SpanID:      spanID2,
				})
				require.NoError(t, err)
			}

			err = l.ot.Shutdown(ctx)
			require.NoError(t, err)

			err = mock.ExpectationsWereMet()
			require.NoError(t, err)
		})
	}
}

// TestLogger_Decision_BlocksUntilExportCompletes covers §2.3.2: Decision() must not return until the
// record has actually reached its exporter, not merely been queued with the batch processor — proven
// with a deliberately slow exporter, since a fast return would mean the write hadn't really finished.
func TestLogger_Decision_BlocksUntilExportCompletes(t *testing.T) {
	t.Parallel()

	const delay = 100 * time.Millisecond

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	inner := tracetest.NewInMemoryExporter()
	slow := &slowExporter{SpanExporter: inner, delay: delay}

	l, err := New(ctx, "openftv", opentelemetry.WithExporter(slow), opentelemetry.WithBatchTimeout(time.Second))
	require.NoError(t, err)

	start := time.Now()
	require.NoError(t, l.Decision(ctx, &Decision{Timestamp: time.Now().UTC(), RequestType: EvaluationEndpoint}))

	elapsed := time.Since(start)

	// Captured before Shutdown: InMemoryExporter.Shutdown() clears its recorded spans.
	spans := inner.GetSpans()

	require.NoError(t, l.Shutdown(ctx))

	assert.GreaterOrEqual(t, elapsed, delay, "Decision() must block until the export completes, not just queue the record")
	assert.Len(t, spans, 1, "the record must have actually reached the exporter by the time Decision() returned")
}
