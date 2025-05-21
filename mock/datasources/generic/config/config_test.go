package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/gjuyn/go-config/config"
	"gitlab.com/gjuyn/go-config/config-ext/yaml"
)

var goodCfg = `
svc:
  host: "127.0.0.1"
  port: 8080
  maxBody: 2048
log:
  format: "text"
  level: "warn"
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
			wantLogLvl:  "warn",
			wantLogOut:  "STDOUT",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			file := dir + tc.file
			err := os.WriteFile(file, []byte(tc.data), 0644)
			require.NoError(t, err)

			c, l, err2 := New(config.NoFlags(), yaml.FilesYAML(file))
			if tc.wantErr {
				require.Error(t, err2)
				require.Nil(t, c)
			} else {
				require.NoError(t, err2)
				require.NotNil(t, l)
				require.NotNil(t, c)

				assert.Equal(t, tc.wantHost, c.Host)
				assert.Equal(t, tc.wantPort, c.Port)
				assert.Equal(t, tc.wantMaxBody, c.MaxBody)
				assert.Equal(t, tc.wantLogFmt, c.Log.Format)
				assert.Equal(t, tc.wantLogLvl, c.Log.Level)
				assert.Equal(t, tc.wantLogOut, c.Log.Output)
				assert.True(t, c.Log.Source)
			}
		})
	}
}
