package slog

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		output    string
		format    string
		level     string
		wantErr   bool
		testLevel slog.Level
		wantLevel bool
		want      slog.Handler
	}{
		{
			name:    "bad file",
			output:  "/this/is/not/a/valid/file.duh",
			wantErr: true,
		},
		{
			name:      "all empty",
			testLevel: slog.LevelInfo,
			wantLevel: true,
			want:      &slog.JSONHandler{},
		},
		{
			name:      "stdout - no fmt - no level",
			output:    "STDOUT",
			testLevel: slog.LevelInfo,
			wantLevel: true,
			want:      &slog.JSONHandler{},
		},
		{
			name:      "stderr - no fmt - no level",
			output:    "stderr",
			format:    "",
			level:     "",
			testLevel: slog.LevelInfo,
			wantLevel: true,
			want:      &slog.JSONHandler{},
		},
		{
			name:      "stdout - no fmt - warning",
			output:    "STDOUT",
			level:     "warning",
			testLevel: slog.LevelInfo,
			want:      &slog.JSONHandler{},
		},
		{
			name:      "stdout - text - debug",
			output:    "stdout",
			format:    "text",
			level:     "debug",
			testLevel: slog.LevelInfo,
			wantLevel: true,
			want:      &slog.TextHandler{},
		},
		{
			name:      "stdout - console - debug",
			output:    "STDOUT",
			format:    "CONSOLE",
			level:     "DEBUG",
			testLevel: slog.LevelInfo,
			wantLevel: true,
			want:      &ConsoleHandler{},
		},
		{
			name:      "stdout - json - error",
			output:    "STDOUT",
			format:    "JSON",
			level:     "error",
			testLevel: slog.LevelWarn,
			want:      &slog.JSONHandler{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, gotErr := Init(tc.output, tc.format, tc.level, false)
			if tc.wantErr {
				assert.Error(t, gotErr)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, gotErr)
				require.NotNil(t, got)

				gotLevel := got.Enabled(nil, tc.testLevel)
				assert.Equal(t, tc.wantLevel, gotLevel)

				assert.IsType(t, tc.want, got.Handler())
			}
		})
	}
}
