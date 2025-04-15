package opa

import (
	"log/slog"
	"testing"

	"github.com/open-policy-agent/opa/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/slog"
)

func TestWrappedLogger_GetLevel(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		level slog.Level
		want  logging.Level
	}{
		{name: "debug", level: slog.LevelDebug, want: logging.Debug},
		{name: "info", level: slog.LevelInfo, want: logging.Info},
		{name: "warn", level: slog.LevelWarn, want: logging.Warn},
		{name: "error", level: slog.LevelError, want: logging.Error},
		{name: "-1", level: -1, want: logging.Info},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := slog2.NewDummyHandler(tc.level)
			w := &wrappedLogger{logger: slog.New(h)}

			got := w.GetLevel()
			assert.Equal(t, tc.want, got)

			w.SetLevel(logging.Debug) // this should do nothing!

			got = w.GetLevel()
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestWrappedLogger_Write(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		level   slog.Level
		debug   []string
		info    []string
		warn    []string
		error   []string
		wantLog int
	}{
		{
			name:  "debug - none",
			level: slog.LevelDebug,
		},
		{
			name:    "debug - 2x debug",
			level:   slog.LevelDebug,
			debug:   []string{"debug1", "debug2"},
			wantLog: 2,
		},
		{
			name:    "debug - 3x info",
			level:   slog.LevelDebug,
			info:    []string{"info1", "info2", "info3"},
			wantLog: 3,
		},
		{
			name:    "debug - many",
			level:   slog.LevelDebug,
			debug:   []string{"debug1", "debug2"},
			info:    []string{"info1", "info2", "info3"},
			warn:    []string{"warn1", "warn2", "warn3", "warn4"},
			error:   []string{"err1", "err2", "err3", "err4", "err5"},
			wantLog: 14,
		},
		{
			name:  "info - 2x debug",
			level: slog.LevelInfo,
			debug: []string{"debug1", "debug2"},
		},
		{
			name:    "info - 3x info",
			level:   slog.LevelInfo,
			info:    []string{"info1", "info2", "info3"},
			wantLog: 3,
		},
		{
			name:    "info - many",
			level:   slog.LevelInfo,
			debug:   []string{"debug1", "debug2"},
			info:    []string{"info1", "info2", "info3"},
			warn:    []string{"warn1", "warn2", "warn3", "warn4"},
			error:   []string{"err1", "err2", "err3", "err4", "err5"},
			wantLog: 12,
		},
		{
			name:  "error - 2x debug",
			level: slog.LevelError,
			debug: []string{"debug1", "debug2"},
		},
		{
			name:  "error - 3x info",
			level: slog.LevelError,
			info:  []string{"info1", "info2", "info3"},
		},
		{
			name:    "error - many",
			level:   slog.LevelError,
			debug:   []string{"debug1", "debug2"},
			info:    []string{"info1", "info2", "info3"},
			warn:    []string{"warn1", "warn2", "warn3", "warn4"},
			error:   []string{"err1", "err2", "err3", "err4", "err5"},
			wantLog: 5,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := slog2.NewDummyHandler(tc.level)
			w := &wrappedLogger{logger: slog.New(h)}

			for i := range tc.debug {
				w.Debug(tc.debug[i])
			}

			for i := range tc.info {
				w.Info(tc.info[i])
			}

			for i := range tc.warn {
				w.Warn(tc.warn[i])
			}

			for i := range tc.error {
				w.Error(tc.error[i])
			}

			assert.Equal(t, tc.wantLog, h.Count())
		})
	}
}

func TestWrappedLogger_WithFields(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		fields map[string]any
		want   []any
	}{
		{
			name: "no fields",
		},
		{
			name:   "one",
			fields: map[string]any{"one": 1},
			want:   []any{"one", 1},
		},
		{
			name:   "few",
			fields: map[string]any{"one": 1, "two": true, "three": 14.25, "a": "b"},
			want:   []any{"a", "b", "one", 1, "three", 14.25, "two", true},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := slog2.NewDummyHandler(slog.LevelInfo)
			w := &wrappedLogger{logger: slog.New(h)}

			w2 := w.WithFields(tc.fields)
			require.NotNil(t, w2)

			w3, ok := w2.(*wrappedLogger)
			require.True(t, ok)
			require.NotNil(t, w3)
			assert.Equal(t, w.logger, w3.logger)

			assert.EqualValues(t, tc.want, w3.fields)
		})
	}
}
