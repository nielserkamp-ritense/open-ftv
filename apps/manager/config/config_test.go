package config

import (
	"os"
	"testing"

	"gitlab.com/gjuyn/go-config/config-ext/yaml"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/gjuyn/go-config/config"
)

var goodCfg = `
svc:
  host: "127.0.0.1"
  port: 8080
  maxBody: 2048
log:
  format: "text"
  level: "info"
  output: "STDOUT"
  source: true
`

var badLogger = `
svc:
  host: "127.0.0.1"
  port: 8080
  maxBody: 2048
log:
  format: "json"
  level: "debug"
  output: "/this/is/not/a/valid/path.log"
`

func TestNew(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	testCases := []struct {
		name        string
		file        string
		data        string
		wantErr     bool
		wantHost    string
		wantPort    uint16
		wantMaxBody int
		wantLogFmt  string
		wantLogLvl  string
		wantLogOut  string
	}{
		{
			name:    "bad input",
			file:    "cfg1.yaml",
			data:    "%not yaml%",
			wantErr: true,
		},
		{
			name:    "bad log file",
			file:    "cfg2.yaml",
			data:    badLogger,
			wantErr: true,
		},
		{
			name:        "good file",
			file:        "cfg3.yaml",
			data:        goodCfg,
			wantHost:    "127.0.0.1",
			wantPort:    8080,
			wantMaxBody: 2048,
			wantLogFmt:  "text",
			wantLogLvl:  "info",
			wantLogOut:  "STDOUT",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			file := dir + tc.file
			err := os.WriteFile(file, []byte(tc.data), 0644)
			require.NoError(t, err)

			defer func() {
				e := recover()
				if tc.wantErr {
					require.NotNil(t, e)
				} else {
					require.Nil(t, e)
				}
			}()

			c, l := New(config.NoFlags(), yaml.FilesYAML(file))
			if tc.wantErr {
				require.Nil(t, c)
				require.Nil(t, l)
			} else {
				require.NotNil(t, c)
				require.NotNil(t, l)

				assert.Equal(t, tc.wantHost, c.Host)
				assert.Equal(t, tc.wantPort, c.Port)
				assert.Equal(t, tc.wantMaxBody, c.MaxBody)
				assert.Equal(t, tc.wantLogFmt, c.Format)
				assert.Equal(t, tc.wantLogLvl, c.Level)
				assert.Equal(t, tc.wantLogOut, c.Output)
				assert.True(t, c.Log.Source)
			}
		})
	}
}
