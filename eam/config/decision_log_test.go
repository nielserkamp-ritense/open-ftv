package config

import (
	"os"
	"testing"
	"time"

	"gitlab.com/gjuyn/go-config/config-ext/yaml"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/gjuyn/go-config/config"
)

func TestOpenTel(t *testing.T) {
	t.Parallel()

	const goodCfg = `
log:
  decisions:
    type: "opentelemetry"
    service: "authlog"
    timeout: "15s"
    postgresql:
      url: "postgres://localhost:5432/mydb?sslmode=disable"
    otel:
      url: "localhost:111"
      insecure: true
    file:
      prettyPrint: true
`
	dir := t.TempDir()

	testCases := []struct {
		name           string
		file           string
		data           string
		args           []string
		wantErr        bool
		wantType       string
		wantService    string
		wantTimeout    time.Duration
		wantPgURL      string
		wantPgMaxLife  time.Duration
		wantPgMaxConn  int32
		wantOtURL      string
		wantOtInsecure bool
		wantPretty     bool
	}{
		{
			name:    "bad input",
			file:    "adl1.yaml",
			data:    "%not yaml%",
			wantErr: true,
		},
		{
			name:           "good file",
			file:           "adl2.yaml",
			data:           goodCfg,
			wantType:       "opentelemetry",
			wantService:    "authlog",
			wantTimeout:    15 * time.Second,
			wantPgURL:      "postgres://localhost:5432/mydb?sslmode=disable",
			wantPgMaxLife:  5 * time.Minute,
			wantPgMaxConn:  100,
			wantOtURL:      "localhost:111",
			wantOtInsecure: true,
			wantPretty:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			file := dir + tc.file
			err := os.WriteFile(file, []byte(tc.data), 0644)

			cfg := &DecisionLog{}

			opts := []config.Option{
				yaml.FilesYAML(file),
				config.WithArgs(tc.args...),
				config.EnvironmentPrefix("CFG_"),
				config.AppName("test 1.0"),
				config.NoHelpOnError(),
			}

			err = config.LoadConfig(cfg, opts...)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.wantType, cfg.Type)
				assert.Equal(t, tc.wantService, cfg.Service)
				assert.Equal(t, tc.wantTimeout, cfg.Timeout)
				assert.Equal(t, tc.wantPgURL, cfg.PgURL)
				assert.Equal(t, tc.wantPgMaxLife, cfg.PgMaxLife)
				assert.Equal(t, tc.wantPgMaxConn, cfg.PgMaxConn)
				assert.Equal(t, tc.wantOtURL, cfg.OtelURL)
				assert.Equal(t, tc.wantOtInsecure, cfg.OtelInsecure)
				assert.Equal(t, tc.wantPretty, cfg.Pretty)
			}
		})
	}
}
