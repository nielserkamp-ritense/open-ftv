package opentelemetry

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestOption(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		opt          Option
		wantErr      bool
		wantExporter bool
		wantTimeout  time.Duration
	}{
		{
			name:    "exporter bad",
			opt:     WithExporter(nil),
			wantErr: true,
		},
		{
			name:         "exporter good",
			opt:          WithExporter(tracetest.NewInMemoryExporter()),
			wantExporter: true,
			wantTimeout:  5 * time.Second,
		},
		{
			name:    "batch timeout bad",
			opt:     WithBatchTimeout(0),
			wantErr: true,
		},
		{
			name:        "batch timeout good",
			opt:         WithBatchTimeout(10 * time.Second),
			wantTimeout: 10 * time.Second,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			e := &Exporter{service: "test", timeout: 5 * time.Second}

			tc.opt(e)
			if tc.wantErr {
				require.Error(t, e.err)
			} else {
				if tc.wantExporter {
					require.NotNil(t, e.exporter)
				}

				assert.Equal(t, tc.wantTimeout, e.timeout)
			}
		})
	}
}
