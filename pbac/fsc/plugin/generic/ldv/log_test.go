package ldv

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

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/fsc/plugin/generic/config"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/slog"
)

func TestNew(t *testing.T) {
	testCases := []struct {
		name      string
		url       string
		service   string
		aid       string
		timeout   time.Duration
		wantLog   int
		wantValue bool
	}{
		{
			name:    "empty",
			wantLog: 1,
		},
		{
			name:    "bad activityID",
			service: "fsc-auth",
			aid:     "not-a-uuid",
			wantLog: 1,
		},
		{
			name:      "bad url",
			url:       string([]byte{0, 1, 127, 128, 129}),
			service:   "fsc-auth",
			aid:       "8abcf4a091574795a328a8e83e77e3f8",
			wantLog:   2,
			wantValue: true,
		},
		{
			name:      "without url",
			service:   "fsc-auth",
			aid:       "8abcf4a091574795a328a8e83e77e3f8",
			wantLog:   1,
			wantValue: true,
		},
		{
			name:      "with url",
			url:       "http://localhost:3000",
			service:   "fsc-auth",
			aid:       "8abcf4a091574795a328a8e83e77e3f8",
			wantLog:   1,
			wantValue: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			q := quiet()
			defer func() {
				q()
			}()

			cfg := &config.Config{
				OpenTelURL:         tc.url,
				OpenTelServiceName: tc.service,
				OpenTelActivityID:  tc.aid,
				OpenTelTimeout:     tc.timeout,
			}

			h := slog2.NewDummyHandler(slog.LevelDebug)
			logger := slog.New(h)

			l := New(cfg, logger)
			assert.Equal(t, tc.wantLog, h.Count())
			assert.Equal(t, tc.wantValue, l != nil)

			if l != nil {
				ctx, span := l.StartSpan(context.Background())
				require.NotNil(t, ctx)
				require.NotNil(t, span)

				span.SetAttributes(attribute.String("key", "value"))
				span.End()

				time.Sleep(50 * time.Millisecond)

				err := l.Shutdown(context.Background())
				require.NoError(t, err)
			}
		})
	}
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
