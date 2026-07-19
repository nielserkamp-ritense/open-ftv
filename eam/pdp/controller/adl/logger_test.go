package adl

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/decisions"
	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/authzen"
	otel "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/opentelemetry"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

func TestADL_Evaluation(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()

	testCases := []struct {
		name    string
		opts    []Option
		req     oas.EvaluationRequest
		resp    oas.EvaluationResponse
		traceID string
		spanID  string
	}{
		{
			name: "basic",
			opts: []Option{
				WithBundleVersion(12),
				WithEngine(map[string]any{"engine": "OPA"}),
			},
			req:  oas.EvaluationRequest{},
			resp: oas.EvaluationResponse{},
		},
		{
			name: "full",
			opts: []Option{
				WithBundleVersion(15),
				WithEngine(map[string]any{"engine": "OPA", "version": "v1.8.13"}),
				WithInformation(map[string]any{"x": "hello world", "y": true, "z": 18}),
			},
			req: oas.EvaluationRequest{
				Subject:  oas.Entity{Type: "gebruiker", Id: "0012"},
				Action:   oas.Action{Name: "GET"},
				Resource: oas.Entity{Type: "document", Id: "news.pdf"},
			},
			resp:    oas.EvaluationResponse{Decision: true},
			traceID: "12345678123456781234567812345678",
			spanID:  "1234567812345678",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			if tc.traceID != "" {
				ctx = context.WithValue(ctx, "trace_id", tc.traceID)
			}
			if tc.spanID != "" {
				ctx = context.WithValue(ctx, "span_id", tc.spanID)
			}

			h := slog2.NewDummyHandler(slog.LevelDebug)
			l1 := slog.New(h)

			l2, err := decisions.New(ctx, "test", otel.WithSLog(l1, "msg"), otel.WithBatchTimeout(10*time.Millisecond))
			require.NoError(t, err)
			require.NotNil(t, l2)

			logger := New(l2, tc.opts...)
			require.NotNil(t, logger)

			err = logger.Evaluation(ctx, now, &tc.req, &tc.resp)
			require.NoError(t, err)

			time.Sleep(50 * time.Millisecond)

			err = l2.Shutdown(ctx)
			require.NoError(t, err)

			assert.Equal(t, 1, h.Count())
		})
	}
}

func TestADL_Evaluations(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()

	testCases := []struct {
		name    string
		opts    []Option
		req     oas.EvaluationsRequest
		resp    oas.EvaluationsResponse
		traceID string
		spanID  string
	}{
		{
			name: "basic",
			opts: []Option{
				WithBundleVersion(12),
				WithEngine(map[string]any{"engine": "OPA"}),
			},
			req:  oas.EvaluationsRequest{},
			resp: oas.EvaluationsResponse{},
		},
		{
			name: "full",
			opts: []Option{
				WithBundleVersion(15),
				WithEngine(map[string]any{"engine": "OPA", "version": "v1.8.13"}),
				WithInformation(map[string]any{"x": "hello world", "y": true, "z": 18}),
			},
			req: oas.EvaluationsRequest{
				Subject:  oas.Entity{Type: "gebruiker", Id: "0012"},
				Action:   oas.Action{Name: "GET"},
				Resource: oas.Entity{Type: "document", Id: "news.pdf"},
			},
			resp:    oas.EvaluationsResponse{Evaluations: []oas.EvaluationDecision{{Decision: true}}},
			traceID: "12345678123456781234567812345678",
			spanID:  "1234567812345678",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			if tc.traceID != "" {
				ctx = context.WithValue(ctx, "trace_id", tc.traceID)
			}
			if tc.spanID != "" {
				ctx = context.WithValue(ctx, "span_id", tc.spanID)
			}

			h := slog2.NewDummyHandler(slog.LevelDebug)
			l1 := slog.New(h)

			l2, err := decisions.New(ctx, "test", otel.WithSLog(l1, "msg"), otel.WithBatchTimeout(10*time.Millisecond))
			require.NoError(t, err)
			require.NotNil(t, l2)

			logger := New(l2, tc.opts...)
			require.NotNil(t, logger)

			err = logger.Evaluations(ctx, now, &tc.req, &tc.resp)
			require.NoError(t, err)

			time.Sleep(50 * time.Millisecond)

			err = l2.Shutdown(ctx)
			require.NoError(t, err)

			assert.Equal(t, 1, h.Count())
		})
	}
}

func TestADL_SearchSubject(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()

	testCases := []struct {
		name    string
		opts    []Option
		req     oas.SearchRequest
		resp    oas.SearchResponse
		traceID string
		spanID  string
	}{
		{
			name: "basic",
			opts: []Option{
				WithBundleVersion(12),
				WithEngine(map[string]any{"engine": "OPA"}),
			},
			req:  oas.SearchRequest{},
			resp: oas.SearchResponse{},
		},
		{
			name: "full",
			opts: []Option{
				WithBundleVersion(15),
				WithEngine(map[string]any{"engine": "OPA", "version": "v1.8.13"}),
				WithInformation(map[string]any{"x": "hello world", "y": true, "z": 18}),
			},
			req: oas.SearchRequest{
				Subject:  oas.SearchEntity{Type: "gebruiker", Id: "0012"},
				Action:   oas.Action{Name: "GET"},
				Resource: oas.SearchEntity{Type: "document", Id: "news.pdf"},
			},
			resp:    oas.SearchResponse{Results: []oas.SearchResult{{Type: "user", Id: "12"}}},
			traceID: "12345678123456781234567812345678",
			spanID:  "1234567812345678",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			if tc.traceID != "" {
				ctx = context.WithValue(ctx, "trace_id", tc.traceID)
			}
			if tc.spanID != "" {
				ctx = context.WithValue(ctx, "span_id", tc.spanID)
			}

			h := slog2.NewDummyHandler(slog.LevelDebug)
			l1 := slog.New(h)

			l2, err := decisions.New(ctx, "test", otel.WithSLog(l1, "msg"), otel.WithBatchTimeout(10*time.Millisecond))
			require.NoError(t, err)
			require.NotNil(t, l2)

			logger := New(l2, tc.opts...)
			require.NotNil(t, logger)

			err = logger.SearchSubject(ctx, now, &tc.req, &tc.resp)
			require.NoError(t, err)

			time.Sleep(50 * time.Millisecond)

			err = l2.Shutdown(ctx)
			require.NoError(t, err)

			assert.Equal(t, 1, h.Count())
		})
	}
}

func TestADL_NewBundle(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		version uint64
	}{
		{name: "zero"},
		{name: "1", version: 1},
		{name: "500", version: 500},
		{name: "99999999", version: 99999999},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			logger, err := decisions.New(ctx, "test", otel.WithFile(os.Stdout, true))
			require.NoError(t, err)
			require.NotNil(t, logger)

			l := New(logger)
			require.NotNil(t, l)

			l.NewBundle(tc.version)
			assert.Equal(t, tc.version, l.bundleVersion)
		})
	}
}

func TestADL_NewInfo(t *testing.T) {
	t.Parallel()

	m1 := map[string]any{"x": "hello world", "y": 123.456}
	m2 := map[string]any{"y": "hello jupiter", "z": true}

	testCases := []struct {
		name string
		info map[string]any
	}{
		{name: "empty"},
		{name: "m1", info: m1},
		{name: "m2", info: m2},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			logger, err := decisions.New(ctx, "test", otel.WithFile(os.Stdout, true))
			require.NoError(t, err)
			require.NotNil(t, logger)

			l := New(logger)
			require.NotNil(t, l)

			l.NewInformation(tc.info)
			if len(tc.info) > 0 {
				b, _ := json.Marshal(tc.info)
				assert.EqualValues(t, b, l.information)
			} else {
				assert.Nil(t, l.information)
			}
		})
	}
}

func TestADL_NewEngine(t *testing.T) {
	t.Parallel()

	m3 := map[string]any{"engine": "Cerbos", "version": "v1.6.3"}

	testCases := []struct {
		name   string
		engine map[string]any
	}{
		{name: "empty"},
		{name: "m3", engine: m3},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			logger, err := decisions.New(ctx, "test", otel.WithFile(os.Stdout, true))
			require.NoError(t, err)
			require.NotNil(t, logger)

			l := New(logger)
			require.NotNil(t, l)

			l.NewEngine(tc.engine)

			if len(tc.engine) > 0 {
				b, _ := json.Marshal(tc.engine)
				assert.EqualValues(t, b, l.engine)
			} else {
				assert.Nil(t, l.engine)
			}
		})
	}
}
