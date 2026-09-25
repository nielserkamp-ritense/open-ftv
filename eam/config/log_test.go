package config

import (
	"flag"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLog(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		args       []string
		wantErr    bool
		wantOutput string
		wantFormat string
		wantLevel  string
		wantSource bool
	}{
		{
			name:    "bad source",
			args:    []string{"--log-source=x"},
			wantErr: true,
		},
		{
			name:       "good args",
			args:       []string{"--log-output=/logs/mylog.txt", "--log-format=TEXT", "--log-level=info", "--log-source=true"},
			wantOutput: "/logs/mylog.txt",
			wantFormat: "TEXT",
			wantLevel:  "info",
			wantSource: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := &Log{}

			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			fs.SetOutput(io.Discard)
			cfg.RegisterFlags(fs)

			err := fs.Parse(tc.args)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.wantOutput, cfg.Output)
				assert.Equal(t, tc.wantFormat, cfg.Format)
				assert.Equal(t, tc.wantLevel, cfg.Level)
				assert.Equal(t, tc.wantSource, cfg.Source)
			}
		})
	}
}
