package fiber

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/decisions"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller/adl"
	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/authzen"
	otel "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/opentelemetry"
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

// statusAttr pulls the "status" span attribute out of a captured span stub.
func statusAttr(t *testing.T, span *tracetest.SpanStub) string {
	t.Helper()

	for _, kv := range span.Attributes {
		if kv.Key == "status" {
			return kv.Value.AsString()
		}
	}

	t.Fatal("status attribute not found")

	return ""
}

func TestAuthProcess_LogDecision(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		err  error
	}{
		{name: "PDP produced a decision"},
		{name: "PDP could not produce a decision", err: errors.New("engine fault")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			h := slog2.NewDummyHandler(slog.LevelDebug)
			l1 := slog.New(h)

			logger, err := decisions.New(ctx, "test", otel.WithSLog(l1, "msg"), otel.WithBatchTimeout(10*time.Millisecond))
			require.NoError(t, err)

			p := &authProcess{
				logger:  slog.New(slog2.NewDummyHandler(slog.LevelError)),
				adl:     adl.New(logger),
				authReq: &oas.EvaluationRequest{Subject: oas.Entity{Type: "gebruiker", Id: "0012"}},
				err:     tc.err,
				started: time.Now().UTC(),
			}
			if tc.err == nil {
				p.authResp = &oas.EvaluationDecision{Decision: true}
			}

			p.logDecision(ctx)

			time.Sleep(50 * time.Millisecond)

			err = logger.Shutdown(ctx)
			require.NoError(t, err)

			// Exactly one ADL record either way. The Logius spec says a completed decision and a failed evaluation
			// attempt each MUST produce exactly one log record.
			assert.Equal(t, 1, h.Count())
		})
	}
}

// TestAuthProcess_LogDecision_UsesDecidedNotStarted covers §3.3.5: the record's timestamp MUST be when
// the decision was actually made (p.decided, set right after the PDP call returns), not when the
// request arrived (p.started).
func TestAuthProcess_LogDecision_UsesDecidedNotStarted(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	exporter := tracetest.NewInMemoryExporter()
	logger, err := decisions.New(ctx, "test", otel.WithExporter(exporter), otel.WithBatchTimeout(10*time.Millisecond))
	require.NoError(t, err)

	started := time.Now().UTC().Add(-time.Hour)
	decided := time.Now().UTC()

	p := &authProcess{
		logger:   slog.New(slog2.NewDummyHandler(slog.LevelError)),
		adl:      adl.New(logger),
		authReq:  &oas.EvaluationRequest{Subject: oas.Entity{Type: "gebruiker", Id: "0012"}},
		authResp: &oas.EvaluationDecision{Decision: true},
		started:  started,
		decided:  decided,
	}

	p.logDecision(ctx)

	spans := exporter.GetSpans()

	require.NoError(t, logger.Shutdown(ctx))

	require.Len(t, spans, 1)
	assert.True(t, spans[0].StartTime.Equal(decided), "must use p.decided as the record's timestamp")
	assert.False(t, spans[0].StartTime.Equal(started), "must not fall back to p.started when p.decided is set")
}

// TestAuthProcess_LogDecision_DenialIsNotFailure covers §3.3.6: a denied request is still a successful
// decision — Status must be Ok, never Error. Error is reserved for the PDP failing to reach a decision
// at all (covered by the "PDP could not produce a decision" case in TestAuthProcess_LogDecision).
func TestAuthProcess_LogDecision_DenialIsNotFailure(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	exporter := tracetest.NewInMemoryExporter()
	logger, err := decisions.New(ctx, "test", otel.WithExporter(exporter), otel.WithBatchTimeout(10*time.Millisecond))
	require.NoError(t, err)

	p := &authProcess{
		logger:   slog.New(slog2.NewDummyHandler(slog.LevelError)),
		adl:      adl.New(logger),
		authReq:  &oas.EvaluationRequest{Subject: oas.Entity{Type: "gebruiker", Id: "0012"}},
		authResp: &oas.EvaluationDecision{Decision: false}, // denied, but the PDP did produce a decision
		started:  time.Now().UTC(),
		// p.err is deliberately nil: the PDP evaluated successfully and denied, it did not fail.
	}

	p.logDecision(ctx)

	spans := exporter.GetSpans()

	require.NoError(t, logger.Shutdown(ctx))

	require.Len(t, spans, 1)
	assert.Equal(t, string(decisions.StatusOk), statusAttr(t, &spans[0]))
}

// TestAuthProcess_Finish_BlocksUntilDecisionWritten covers §2.3.2: the handler must not answer its
// caller until the ADL record has reached durable storage. Proven end-to-end through a real fiber
// request/response cycle with a deliberately slow decision-log exporter: if the HTTP response came back
// before the write finished, app.Test would return faster than the artificial delay.
func TestAuthProcess_Finish_BlocksUntilDecisionWritten(t *testing.T) {
	t.Parallel()

	const delay = 100 * time.Millisecond

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	inner := tracetest.NewInMemoryExporter()
	slow := &slowExporter{SpanExporter: inner, delay: delay}

	logger, err := decisions.New(ctx, "test", otel.WithExporter(slow), otel.WithBatchTimeout(time.Second))
	require.NoError(t, err)

	app := fiber.New()
	app.Get("/", func(fc *fiber.Ctx) error {
		p, finish := initAuthProcess(fc, slog.New(slog2.NewDummyHandler(slog.LevelError)), adl.New(logger), nil)
		p.authReq = &oas.EvaluationRequest{Subject: oas.Entity{Type: "gebruiker", Id: "0012"}}
		p.authResp = &oas.EvaluationDecision{Decision: true}

		defer finish()

		return fc.SendStatus(fiber.StatusOK)
	})

	start := time.Now()
	resp, err2 := app.Test(httptest.NewRequest(fiber.MethodGet, "/", http.NoBody), -1)
	require.NoError(t, err2)
	resp.Body.Close()

	elapsed := time.Since(start)

	spans := inner.GetSpans()

	require.NoError(t, logger.Shutdown(ctx))

	assert.GreaterOrEqual(t, elapsed, delay, "the response must not come back before the ADL write completes")
	assert.Len(t, spans, 1)
}
