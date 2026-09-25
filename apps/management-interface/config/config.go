package config

import (
	"flag"
	"os"
	"time"

	config2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/config"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/envflag"
)

var (
	GitHash = "unknown"
	GitTag  = "unknown"
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

const (
	defaultPort        = 8080
	defaultHealthPort  = 8081
	defaultTimeout     = 30 * time.Second
	defaultIdleTimeout = 300 * time.Second
)

// Config represents our configuration variables.
//
// The ports differ from the shared server configuration: the interface listens on 8080 like the former nginx image,
// so the health service moves to 8081.
type Config struct {
	Log          config2.Log
	Host         string
	Port         uint16
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	HealthHost   string
	HealthPort   uint16
	HealthRead   time.Duration
	HealthWrite  time.Duration
	HealthIdle   time.Duration
}

// Load reads the configuration from the command-line args, falling back to MANAGEMENT_INTERFACE_ environment variables.
func Load(args []string, lookupEnv func(string) (string, bool)) (*Config, error) {
	cfg := &Config{}

	fs := flag.NewFlagSet("management-interface", flag.ContinueOnError)
	cfg.Log.RegisterFlags(fs)
	fs.StringVar(&cfg.Host, "host", "0.0.0.0", "Address to use for service")
	envflag.Uint16Var(fs, &cfg.Port, "port", defaultPort, "Port to use for service")
	fs.DurationVar(&cfg.ReadTimeout, "read-timeout", defaultTimeout, "Read timeout for requests")
	fs.DurationVar(&cfg.WriteTimeout, "write-timeout", defaultTimeout, "Write timeout for requests")
	fs.DurationVar(&cfg.IdleTimeout, "idle-timeout", defaultIdleTimeout, "Idle timeout for requests")
	fs.StringVar(&cfg.HealthHost, "health-host", "", "Address to use for health service (default main host)")
	envflag.Uint16Var(fs, &cfg.HealthPort, "health-port", defaultHealthPort, "Port to use for health service")
	fs.DurationVar(&cfg.HealthRead, "health-read", defaultTimeout, "Read timeout for health service")
	fs.DurationVar(&cfg.HealthWrite, "health-write", defaultTimeout, "Write timeout for health service")
	fs.DurationVar(&cfg.HealthIdle, "health-idle", defaultIdleTimeout, "Idle timeout for health service")

	if err := envflag.Parse(fs, args, "MANAGEMENT_INTERFACE_", lookupEnv); err != nil {
		return nil, err
	}

	if cfg.HealthHost == "" {
		cfg.HealthHost = cfg.Host
	}

	return cfg, nil
}

func RuntimeEnv() map[string]string {
	env := make(map[string]string, len(runtimeEnvNames))
	for _, name := range runtimeEnvNames {
		env[name] = os.Getenv(name)
	}

	return env
}
