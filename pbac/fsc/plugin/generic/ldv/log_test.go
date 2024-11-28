package ldv

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
			name:      "bad url",
			url:       string([]byte{0, 1, 127, 128, 129}),
			service:   "fsc-auth",
			wantLog:   2,
			wantValue: true,
		},
		{
			name:      "no url",
			service:   "fsc-auth",
			wantLog:   1,
			wantValue: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
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
				err := l.Shutdown(context.Background())
				require.NoError(t, err)
			}
		})
	}
}
