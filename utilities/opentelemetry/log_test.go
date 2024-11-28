package opentelemetry

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func TestNew(t *testing.T) {
	testCases := []struct {
		name    string
		service string
		url     string
		timeout time.Duration
		opts    []trace.TracerOption
		wantErr bool
	}{
		{name: "empty", wantErr: true},
		{name: "stdout", service: "fsc-authz"},
		{name: "with url", service: "fsc-authz", url: "localhost:12345"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			l, err := New(tc.service, tc.url, tc.timeout, tc.opts...)
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
	t.Run("start span", func(t *testing.T) {
		q := quiet()
		defer func() {
			q()
		}()

		l, err := New("fsc-authz", "", 10*time.Millisecond)
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
