package fiber

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/decisions"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller/adl"
	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/authzen"
	otel "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/opentelemetry"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

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

			// Exactly one ADL record either way, per the Logius spec: a
			// completed decision and a failed evaluation attempt each MUST
			// produce exactly one log record.
			assert.Equal(t, 1, h.Count())
		})
	}
}
