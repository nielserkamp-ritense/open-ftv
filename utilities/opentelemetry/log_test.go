package opentelemetry

import (
	"context"
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
	h := slog2.NewDummyHandler(slog.LevelDebug)
	sl := slog.New(h)

	testCases := []struct {
		name    string
		service string
		opts    []Option
		wantErr bool
	}{
		{name: "empty", wantErr: true},
		{name: "no exporter", service: "fsc-authz", wantErr: true},
		{name: "bad file", service: "fsc-authz", opts: []Option{WithFile(nil, true)}, wantErr: true},
		{name: "bad slog", service: "fsc-authz", opts: []Option{WithSLog(nil, "yoyo")}, wantErr: true},
		{name: "stdout", service: "fsc-authz", opts: []Option{WithFile(os.Stdout, true)}},
		{name: "stderr", service: "fsc-authz", opts: []Option{WithFile(os.Stderr, false)}},
		{name: "valid slog", service: "fsc-authz", opts: []Option{WithSLog(sl, "hello world")}},
		{name: "valid url", service: "fsc-authz", opts: []Option{WithOT("localhost:12345", true), WithBatchTimeout(time.Second)}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			l, err := New(nil, tc.service, tc.opts...)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, l)
			} else {
				require.NoError(t, err)
				require.NotNil(t, l)

				assert.NotNil(t, l.tp)
				assert.NotNil(t, l.tracer)

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
	t.Run("start span", func(t *testing.T) {
		h := slog2.NewDummyHandler(slog.LevelDebug)
		sl := slog.New(h)

		l, err := New(nil, "fsc-authz", WithSLog(sl, "log"))
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
	})
}
