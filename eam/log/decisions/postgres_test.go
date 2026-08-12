package decisions

import (
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/postgresql/pool"
)

func TestNewPostgreSQL(t *testing.T) {
	t.Parallel()

	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer mock.Close()

	testCases := []struct {
		name    string
		dsn     string
		db      pool.Pooler
		maxLife time.Duration
		maxConn int32
		wantErr bool
	}{
		{name: "bad dsn without pool", dsn: "haha", wantErr: true},
		{name: "bad dsn with pool", dsn: "hihi", db: mock, wantErr: true},
		{name: "good dsn without pool", dsn: "postgres://localhost:5432/mydb", maxLife: 5 * time.Minute, maxConn: 5},
		{name: "good dsn without pool", dsn: "postgres://localhost:5432/mydb", db: mock, maxLife: 2 * time.Minute, maxConn: 10},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := newPostgreSQL(tc.dsn, tc.db, tc.maxLife, tc.maxConn)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
			}
		})
	}
}

func TestDecisionFromSpanToParams(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	ms := now.UnixMilli()
	base := []any{now, ms, nil, nil, nil, "", string(StatusUnset), int64(0), int64(0), nil, nil, nil, nil}
	adlAttrs := DefaultADLAttributes()

	testCases := []struct {
		name       string
		timestamp  time.Time
		attrs      []attribute.KeyValue
		want       Decision
		wantParams []any
	}{
		{
			name:       "only timestamp",
			timestamp:  now,
			want:       Decision{Timestamp: now, Status: StatusUnset},
			wantParams: base,
		},
		{
			name:      "timestamp + request_type (1)",
			timestamp: now,
			attrs:     []attribute.KeyValue{attribute.Int64("request_type", int64(EvaluationEndpoint))},
			want: Decision{
				Timestamp:   now,
				RequestType: EvaluationEndpoint,
				EventName:   "adl.access_evaluation",
				Status:      StatusUnset,
			},
			wantParams: []any{now, ms, nil, nil, nil, "adl.access_evaluation", string(StatusUnset), int64(EvaluationEndpoint), int64(0), nil, nil, nil, nil},
		},
		{
			name:      "timestamp + request_type (2)",
			timestamp: now,
			attrs:     []attribute.KeyValue{attribute.Int64("request_type", int64(SearchActionEndpoint))},
			want: Decision{
				Timestamp:   now,
				RequestType: SearchActionEndpoint,
				EventName:   "adl.search_action",
				Status:      StatusUnset,
			},
			wantParams: []any{now, ms, nil, nil, nil, "adl.search_action", string(StatusUnset), int64(SearchActionEndpoint), int64(0), nil, nil, nil, nil},
		},
		{
			name:       "timestamp + policies",
			timestamp:  now,
			attrs:      []attribute.KeyValue{attribute.Int64("policies", int64(55))},
			want:       Decision{Timestamp: now, Policies: 55, Status: StatusUnset},
			wantParams: []any{now, ms, nil, nil, nil, "", string(StatusUnset), int64(0), int64(55), nil, nil, nil, nil},
		},
		{
			name:       "timestamp + trace_id",
			timestamp:  now,
			attrs:      []attribute.KeyValue{attribute.String("trace_id", "12345678123456781234567812345678")},
			want:       Decision{Timestamp: now, TraceID: "12345678123456781234567812345678", Status: StatusUnset},
			wantParams: []any{now, ms, "12345678123456781234567812345678", nil, nil, "", string(StatusUnset), int64(0), int64(0), nil, nil, nil, nil},
		},
		{
			name:       "timestamp + span_id",
			timestamp:  now,
			attrs:      []attribute.KeyValue{attribute.String("span_id", "1234567812345678")},
			want:       Decision{Timestamp: now, SpanID: "1234567812345678", Status: StatusUnset},
			wantParams: []any{now, ms, nil, "1234567812345678", nil, "", string(StatusUnset), int64(0), int64(0), nil, nil, nil, nil},
		},
		{
			name:       "timestamp + parent_span_id",
			timestamp:  now,
			attrs:      []attribute.KeyValue{attribute.String("parent_span_id", "abcdefabcdefabcd")},
			want:       Decision{Timestamp: now, ParentSpanID: "abcdefabcdefabcd", Status: StatusUnset},
			wantParams: []any{now, ms, nil, nil, "abcdefabcdefabcd", "", string(StatusUnset), int64(0), int64(0), nil, nil, nil, nil},
		},
		{
			name:       "timestamp + body request",
			timestamp:  now,
			attrs:      []attribute.KeyValue{attribute.String("body", `{"request":{"x":123}}`)},
			want:       Decision{Timestamp: now, Request: []byte(`{"x":123}`), Status: StatusUnset},
			wantParams: []any{now, ms, nil, nil, nil, "", string(StatusUnset), int64(0), int64(0), map[string]any{"request": []byte(`{"x":123}`)}, adlAttrs, nil, nil},
		},
		{
			name:       "timestamp + body response",
			timestamp:  now,
			attrs:      []attribute.KeyValue{attribute.String("body", `{"response":{"x":123}}`)},
			want:       Decision{Timestamp: now, Response: []byte(`{"x":123}`), Status: StatusUnset},
			wantParams: []any{now, ms, nil, nil, nil, "", string(StatusUnset), int64(0), int64(0), map[string]any{"response": []byte(`{"x":123}`)}, adlAttrs, nil, nil},
		},
		{
			name:       "timestamp + information",
			timestamp:  now,
			attrs:      []attribute.KeyValue{attribute.String("information", `{"x":123}`)},
			want:       Decision{Timestamp: now, Information: []byte(`{"x":123}`), Status: StatusUnset},
			wantParams: []any{now, ms, nil, nil, nil, "", string(StatusUnset), int64(0), int64(0), nil, nil, []byte(`{"x":123}`), nil},
		},
		{
			name:       "timestamp + engine",
			timestamp:  now,
			attrs:      []attribute.KeyValue{attribute.String("engine", `{"x":123}`)},
			want:       Decision{Timestamp: now, Engine: []byte(`{"x":123}`), Status: StatusUnset},
			wantParams: []any{now, ms, nil, nil, nil, "", string(StatusUnset), int64(0), int64(0), nil, nil, nil, []byte(`{"x":123}`)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := decisionFromSpan(tc.timestamp, tc.attrs)
			require.NotNil(t, got)
			assert.EqualValues(t, tc.want, *got)

			params := decisionToParms(got)
			require.NotNil(t, params)
			require.EqualValues(t, tc.wantParams, params)
		})
	}
}
