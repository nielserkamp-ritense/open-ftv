package envflag

import (
	"flag"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type values struct {
	level   string
	source  bool
	timeout time.Duration
	port    uint16
}

func newFlagSet(v *values) *flag.FlagSet {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.StringVar(&v.level, "log-level", "info", "")
	fs.BoolVar(&v.source, "log-source", true, "")
	fs.DurationVar(&v.timeout, "read-timeout", 30*time.Second, "")
	Uint16Var(fs, &v.port, "port", 8443, "")

	return fs
}

func lookup(env map[string]string) func(string) (string, bool) {
	return func(name string) (string, bool) {
		value, ok := env[name]
		return value, ok
	}
}

func TestParse(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		args    []string
		env     map[string]string
		wantErr bool
		want    values
	}{
		{
			name: "defaults",
			want: values{level: "info", source: true, timeout: 30 * time.Second, port: 8443},
		},
		{
			name: "env overrides default",
			env:  map[string]string{"APP_LOG_LEVEL": "debug", "APP_LOG_SOURCE": "false", "APP_READ_TIMEOUT": "1m", "APP_PORT": "9000"},
			want: values{level: "debug", source: false, timeout: time.Minute, port: 9000},
		},
		{
			name: "flag overrides env",
			args: []string{"--log-level=warn", "-port", "9001"},
			env:  map[string]string{"APP_LOG_LEVEL": "debug", "APP_PORT": "9000"},
			want: values{level: "warn", source: true, timeout: 30 * time.Second, port: 9001},
		},
		{
			name: "env without prefix is ignored",
			env:  map[string]string{"LOG_LEVEL": "debug"},
			want: values{level: "info", source: true, timeout: 30 * time.Second, port: 8443},
		},
		{
			name:    "invalid flag value",
			args:    []string{"--port=x"},
			wantErr: true,
		},
		{
			name:    "invalid env value",
			env:     map[string]string{"APP_READ_TIMEOUT": "soon"},
			wantErr: true,
		},
		{
			name:    "port out of range",
			env:     map[string]string{"APP_PORT": "65536"},
			wantErr: true,
		},
		{
			name:    "unknown flag",
			args:    []string{"--nope"},
			wantErr: true,
		},
		{
			name:    "bool value as a separate argument",
			args:    []string{"--log-source", "false", "--log-level=debug"},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var got values

			fs := newFlagSet(&got)
			fs.SetOutput(io.Discard)

			err := Parse(fs, tc.args, "APP_", lookup(tc.env))
			if tc.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestEnvName(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "PDP_AUTHORIZATION_FAIL_CLOSED_ON_EMPTY", envName("PDP_", "authorization-fail-closed-on-empty"))
}
