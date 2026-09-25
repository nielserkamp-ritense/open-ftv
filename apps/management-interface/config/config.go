package config

import (
	"log/slog"
	"os"
	"time"

	"gitlab.com/gjuyn/go-config/config"
	"gitlab.com/gjuyn/go-config/config-ext/yaml"

	config2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/config"
)

var (
	GitHash = "unknown"
	GitTag  = "unknown"
)

const (
	envPrefix = "MANAGEMENT_INTERFACE_"
	cfg1      = "/etc/management-interface/config.yaml"
	cfg2      = "./etc/management-interface.yaml"
	cfg3      = "./management-interface.yaml"
)

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
	Host         string        `json:"mainHost"             yaml:"svc.host,omitempty"                 env:"ADDRESS,HOST"  flag:"address,host,a,h" desc:"Address to use for service (default 0.0.0.0)"          default:"0.0.0.0"`
	Port         uint16        `json:"mainPort"             yaml:"svc.port,omitempty"                 env:"PORT"          flag:"port,p"           desc:"Port to use for service (default 8080)"                default:"8080"`
	ReadTimeout  time.Duration `json:"mainReadTimeout"      yaml:"svc.timeout.read,omitempty"         env:"READ_TIMEOUT"  flag:"read-timeout"     desc:"Read timeout for requests (default 30s)"               default:"30s"`
	WriteTimeout time.Duration `json:"mainWriteTimeout"     yaml:"svc.timeout.write,omitempty"        env:"WRITE_TIMEOUT" flag:"write-timeout"    desc:"Write timeout for requests (default 30s)"              default:"30s"`
	IdleTimeout  time.Duration `json:"mainIdleTimeout"      yaml:"svc.timeout.idle,omitempty"         env:"IDLE_TIMEOUT"  flag:"idle-timeout"     desc:"Idle timeout for requests (default 300s)"              default:"300s"`
	HealthHost   string        `json:"healthHost,omitempty" yaml:"svc.health.host,omitempty"          env:"HEALTH_HOST"   flag:"health-host"      desc:"Address to use for health service (default main host)"`
	HealthPort   uint16        `json:"healthPort"           yaml:"svc.health.port,omitempty"          env:"HEALTH_PORT"   flag:"health-port"      desc:"Port to use for health service (default 8081)"         default:"8081"`
	HealthRead   time.Duration `json:"healthReadTimeout"    yaml:"svc.health.timeout.read,omitempty"  env:"HEALTH_READ"   flag:"health-read"      desc:"Read timeout for health service (default 30s)"         default:"30s"`
	HealthWrite  time.Duration `json:"healthWriteTimeout"   yaml:"svc.health.timeout.write,omitempty" env:"HEALTH_WRITE"  flag:"health-write"     desc:"Write timeout for health service (default 30s)"        default:"30s"`
	HealthIdle   time.Duration `json:"healthIdleTimeout"    yaml:"svc.health.timeout.idle,omitempty"  env:"HEALTH_IDLE"   flag:"health-idle"      desc:"Idle timeout for health service (default 300s)"        default:"300s"`
}

// New instantiates a new configuration from one or more files, environment variables and command-line flags.
func New(opts ...config.Option) (*Config, *slog.Logger) {
	cfg := &Config{}

	opts = append(
		append(
			[]config.Option{
				yaml.FilesYAML(cfg1, cfg2, cfg3),
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
