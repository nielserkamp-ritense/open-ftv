package opentelemetry

import (
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"

	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

func TestOption(t *testing.T) {
	t.Parallel()

	h := slog2.NewDummyHandler(slog.LevelDebug)
	logger := slog.New(h)

	testCases := []struct {
		name         string
		opt          Option
		wantErr      bool
		wantExporter bool
		wantTimeout  time.Duration
		wantOpts     int
	}{
		{
			name:    "OT bad",
			opt:     WithOT("", true),
			wantErr: true,
		},
		{
			name:         "OT good",
			opt:          WithOT("http://localhost:12345", true),
			wantExporter: true,
			wantTimeout:  5 * time.Second,
		},
		{
			name:    "file bad",
			opt:     WithFile(nil, true),
			wantErr: true,
		},
		{
			name:         "file good",
			opt:          WithFile(os.Stdout, true),
			wantExporter: true,
			wantTimeout:  5 * time.Second,
		},
		{
			name:    "slog bad",
			opt:     WithSLog(nil, "hello world"),
			wantErr: true,
		},
		{
			name:         "slog good",
			opt:          WithSLog(logger, "my service"),
			wantExporter: true,
			wantTimeout:  5 * time.Second,
		},
		{
			name:    "exporter bad",
			opt:     WithExporter(nil),
			wantErr: true,
		},
		{
			name:         "exporter good",
			opt:          WithExporter(&slogLogger{}),
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
		{
			name:        "trace opts",
			opt:         WithTraceOpts(trace.WithInstrumentationVersion("x")),
			wantTimeout: 5 * time.Second,
			wantOpts:    1,
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
				assert.Equal(t, tc.wantOpts, len(e.opts))
			}
		})
	}
}
