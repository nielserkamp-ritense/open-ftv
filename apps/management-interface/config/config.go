package config

import (
	"log/slog"
	"os"
	"time"

	"gitlab.com/gjuyn/go-config/config"

	config2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/config"
)

var (
	GitHash = "unknown"
	GitTag  = "unknown"
)

const envPrefix = "MANAGEMENT_INTERFACE_"

// runtimeEnvNames are the environment variables the browser app reads from window.__ENV__.
// They keep their names without the service prefix, so deployments of the former nginx image keep working.
var runtimeEnvNames = []string{
	"VITE_PAP_BASE_URL",
	"VITE_PIP_BASE_URL",
	"VITE_OIDC_AUTHORITY",
	"VITE_OIDC_CLIENT_ID",
	"MANAGER_MAX_LOGO_SIZE",
	"VITE_APP_VERSION",
}

// Config represents our configuration variables.
//
// The ports differ from the shared server configuration: the interface listens on 8080 like the former nginx image,
// so the health service moves to 8081.
type Config struct {
	config2.Log
	Host         string        `env:"ADDRESS,HOST"  flag:"address,host,a,h" desc:"Address to use for service (default 0.0.0.0)"          default:"0.0.0.0"`
	Port         uint16        `env:"PORT"          flag:"port,p"           desc:"Port to use for service (default 8080)"                default:"8080"`
	ReadTimeout  time.Duration `env:"READ_TIMEOUT"  flag:"read-timeout"     desc:"Read timeout for requests (default 30s)"               default:"30s"`
	WriteTimeout time.Duration `env:"WRITE_TIMEOUT" flag:"write-timeout"    desc:"Write timeout for requests (default 30s)"              default:"30s"`
	IdleTimeout  time.Duration `env:"IDLE_TIMEOUT"  flag:"idle-timeout"     desc:"Idle timeout for requests (default 300s)"              default:"300s"`
	HealthHost   string        `env:"HEALTH_HOST"   flag:"health-host"      desc:"Address to use for health service (default main host)"`
	HealthPort   uint16        `env:"HEALTH_PORT"   flag:"health-port"      desc:"Port to use for health service (default 8081)"         default:"8081"`
	HealthRead   time.Duration `env:"HEALTH_READ"   flag:"health-read"      desc:"Read timeout for health service (default 30s)"         default:"30s"`
	HealthWrite  time.Duration `env:"HEALTH_WRITE"  flag:"health-write"     desc:"Write timeout for health service (default 30s)"        default:"30s"`
	HealthIdle   time.Duration `env:"HEALTH_IDLE"   flag:"health-idle"      desc:"Idle timeout for health service (default 300s)"        default:"300s"`
}

// New instantiates a new configuration from environment variables and command-line flags.
func New(opts ...config.Option) (*Config, *slog.Logger) {
	cfg := &Config{}

	opts = append(
		append(
			[]config.Option{
				config.EnvironmentPrefix(envPrefix),
			},
			opts...,
		),
		config.NoHelpOnError(),
	)

	logger := config2.Load(cfg, &cfg.Log, opts...)

	if cfg.HealthHost == "" {
		cfg.HealthHost = cfg.Host
	}

	return cfg, logger
}

func RuntimeEnv() map[string]string {
	env := make(map[string]string, len(runtimeEnvNames))
	for _, name := range runtimeEnvNames {
		env[name] = os.Getenv(name)
	}

	return env
}

func (c *Config) LogSanitized(logger *slog.Logger) {
	logger.Info("git version", "hash", GitHash, "tag", GitTag)
}
