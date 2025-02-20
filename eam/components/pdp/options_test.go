package pdp

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/ldv"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pip"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/opentelemetry"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/slog"
)

func TestOptions(t *testing.T) {
	h := slog2.NewDummyHandler(slog.LevelInfo)
	logger := slog.New(h)

	ot, err := opentelemetry.New(&opentelemetry.LoggerConfig{
		Service:      "myService",
		URL:          "http://localhost",
		PrettyPrint:  false,
		Logger:       logger,
		BatchTimeout: time.Minute,
	})
	require.NoError(t, err)

	logboek := &lb{activityID: "a1", logger: ot}

	p1 := pip.New(pip.Config{
		Logger:        logger,
		NewAttributes: models.NewAttributeSet,
		NewEntities:   models.NewEntitySet,
	})
	require.NotNil(t, p1)

	p2 := pap.New(nil, logger)
	require.NotNil(t, p2)

	testCases := []struct {
		name        string
		options     []Option
		wantName    string
		wantVersion string
		wantFull    string
		wantStore   string
		wantRecurse bool
		wantLogger  *slog.Logger
		wantLogboek ldv.LDV
		wantPIP     pip.PIP
		wantPAP     pap.PAP
	}{
		{
			name: "no options",
		},
		{
			name:        "name + version",
			options:     []Option{WithNameVersion("x1", "v1")},
			wantName:    "x1",
			wantVersion: "v1",
			wantFull:    "x1 v1",
		},
		{
			name:        "store + recurse",
			options:     []Option{WithStore("/etc", true)},
			wantStore:   "/etc",
			wantRecurse: true,
		},
		{
			name:       "logger",
			options:    []Option{WithLogger(logger)},
			wantLogger: logger,
		},
		{
			name:        "logboek",
			options:     []Option{WithLogboek(logboek)},
			wantLogboek: logboek,
		},
		{
			name:    "pip",
			options: []Option{WithPIP(p1)},
			wantPIP: p1,
		},
		{
			name:    "pap",
			options: []Option{WithPAP(p2)},
			wantPAP: p2,
		},
		{
			name:        "all",
			options:     []Option{WithPAP(p2), WithLogger(logger), WithStore("/etc", true), WithPIP(p1), WithLogboek(logboek), WithNameVersion("x1", "v1")},
			wantName:    "x1",
			wantVersion: "v1",
			wantFull:    "x1 v1",
			wantStore:   "/etc",
			wantRecurse: true,
			wantLogger:  logger,
			wantLogboek: logboek,
			wantPIP:     p1,
			wantPAP:     p2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := NewBase(tc.options...)
			require.NotNil(t, got)

			assert.Equal(t, tc.wantName, got.Name())
			assert.Equal(t, tc.wantVersion, got.Version())
			assert.Equal(t, tc.wantFull, got.String())
			assert.Equal(t, tc.wantStore, got.store)
			assert.Equal(t, tc.wantRecurse, got.recurse)
			assert.Equal(t, tc.wantLogger, got.Logger())
			assert.Equal(t, tc.wantLogboek, got.Logboek())
			assert.Equal(t, tc.wantPIP, got.PIP())
			assert.Equal(t, tc.wantPAP, got.PAP())
		})
	}
}

func (l *lb) StartSpan(ctx context.Context, attributes ...attribute.KeyValue) (context.Context, trace.Span) {
	attributes = append(attributes, attribute.String("authz.activity.id", l.activityID))

	opts := []trace.SpanStartOption{
		trace.WithTimestamp(time.Now().UTC()),
		trace.WithAttributes(attributes...),
	}

	return l.logger.StartSpan(ctx, l.activityID, opts...)
}

func (l *lb) Shutdown(ctx context.Context) error {
	return l.logger.Shutdown(ctx)
}

type lb struct {
	activityID string
	logger     opentelemetry.Logger
}
