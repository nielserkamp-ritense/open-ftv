package opentelemetry

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"

	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

func TestSlogLogger(t *testing.T) {
	t.Run("start span", func(t *testing.T) {
		h := slog2.NewDummyHandler(slog.LevelDebug)
		sl := slog.New(h)

		l, err := New(&LoggerConfig{
			Service:      "fsc-authz",
			URL:          "slog",
			Logger:       sl,
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

		assert.Equal(t, 1, h.Count())
	})
}
