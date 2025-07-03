package opentelemetry

import (
	"context"
	"log"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"

	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

func TestNew(t *testing.T) {
	t.Parallel()

	h := slog2.NewDummyHandler(slog.LevelDebug)
	sl := slog.New(h)

	testCases := []struct {
		name    string
		service string
		url     string
		pretty  bool
		logger  *slog.Logger
		timeout time.Duration
		wantErr bool
	}{
		{name: "empty", wantErr: true},
		{name: "empty url", service: "fsc-authz"},
		{name: "stdout", service: "fsc-authz", url: "stdout", pretty: true},
		{name: "stderr", service: "fsc-authz", url: "stderr"},
		{name: "valid slog", service: "fsc-authz", url: "slog", logger: sl},
		{name: "invalid slog", service: "fsc-authz", url: "slog", wantErr: true},
		{name: "valid url", service: "fsc-authz", url: "localhost:12345", timeout: time.Second},
		{name: "invalid url", service: "fsc-authz", url: "\x00", wantErr: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := &LoggerConfig{
				Service:      tc.service,
				URL:          tc.url,
				PrettyPrint:  tc.pretty,
				Logger:       tc.logger,
				BatchTimeout: tc.timeout,
			}

			l, err := New(cfg)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, l)
			} else {
				require.NoError(t, err)
				require.NotNil(t, l)

				l2, ok := l.(*logger)
				require.True(t, ok)
				require.NotNil(t, l2)

				assert.NotNil(t, l2.tp)
				assert.NotNil(t, l2.tracer)

				err = l.Shutdown(context.Background())
				require.NoError(t, err)

				// double shutdown should not matter
				err = l.Shutdown(context.Background())
				require.NoError(t, err)
			}
		})
	}
}

func TestLogger_StartSpan(t *testing.T) {
	t.Parallel()

	t.Run("start span", func(t *testing.T) {
		q := quiet()
		defer func() {
			q()
		}()

		l, err := New(&LoggerConfig{
			Service:      "fsc-authz",
			BatchTimeout: 10 * time.Millisecond,
		})
		require.NoError(t, err)
		require.NotNil(t, l)

		ctx, span := l.StartSpan(context.Background(), "test")
		require.NotNil(t, ctx)
		require.NotNil(t, span)

		span.SetAttributes(attribute.String("key", "value"))
		span.End()

		time.Sleep(50 * time.Millisecond)

		err = l.Shutdown(ctx)
		require.NoError(t, err)

		err = l.Shutdown(ctx)
		require.NoError(t, err)
	})
}

func quiet() func() {
	null, _ := os.Open(os.DevNull)
	s1 := os.Stdout
	s2 := os.Stderr
	os.Stdout = null
	os.Stderr = null
	log.SetOutput(null)
	return func() {
		defer null.Close()
		os.Stdout = s1
		os.Stderr = s2
		log.SetOutput(os.Stderr)
	}
}
