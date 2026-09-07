package opentelemetry

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestNew(t *testing.T) {
	testCases := []struct {
		name    string
		service string
		opts    []Option
		wantErr bool
	}{
		{name: "empty", wantErr: true},
		{name: "no exporter", service: "fsc-authz", wantErr: true},
		{name: "bad exporter", service: "fsc-authz", opts: []Option{WithExporter(nil)}, wantErr: true},
		{name: "valid exporter", service: "fsc-authz", opts: []Option{WithExporter(tracetest.NewInMemoryExporter()), WithBatchTimeout(time.Second)}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			l, err := New(t.Context(), tc.service, tc.opts...)
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
		exporter := tracetest.NewInMemoryExporter()

		l, err := New(t.Context(), "fsc-authz", WithExporter(exporter), WithBatchTimeout(10*time.Millisecond))
		require.NoError(t, err)
		require.NotNil(t, l)

		ctx, span := l.StartSpan(context.Background(), "test")
		require.NotNil(t, ctx)
		require.NotNil(t, span)

		span.SetAttributes(attribute.String("key", "value"))
		span.End()

		time.Sleep(50 * time.Millisecond)

		assert.Len(t, exporter.GetSpans(), 1)

		err = l.Shutdown(ctx)
		require.NoError(t, err)
	})
}
