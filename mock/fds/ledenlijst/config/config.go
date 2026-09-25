package config

import (
	"flag"
	"time"

	config2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/config"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/envflag"
)

const (
	defaultPort        = 8090
	defaultTimeout     = 30 * time.Second
	defaultIdleTimeout = 300 * time.Second
)

type Config struct {
	Host         string
	Port         uint16
	CA           string
	Cert         string
	Key          string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	Log          config2.Log
}

// Load reads the configuration from the command-line args, falling back to FDS_LEDENLIJST_ environment variables.
func Load(args []string, lookupEnv func(string) (string, bool)) (*Config, error) {
	cfg := &Config{}

	fs := flag.NewFlagSet("ledenlijst", flag.ContinueOnError)
	fs.StringVar(&cfg.Host, "host", "0.0.0.0", "Address to use for service")
	envflag.Uint16Var(fs, &cfg.Port, "port", defaultPort, "Port to use for https service")
	fs.StringVar(&cfg.CA, "tls-ca", "", "TLS Certificate authority to use for https service")
	fs.StringVar(&cfg.Cert, "tls-cert", "", "TLS certificate to use for https service")
	fs.StringVar(&cfg.Key, "tls-key", "", "TLS key authority to use for https service")
	fs.DurationVar(&cfg.ReadTimeout, "read-timeout", defaultTimeout, "Read timeout for API requests")
	fs.DurationVar(&cfg.WriteTimeout, "write-timeout", defaultTimeout, "Write timeout for API requests")
	fs.DurationVar(&cfg.IdleTimeout, "idle-timeout", defaultIdleTimeout, "Idle timeout for API requests")
	cfg.Log.RegisterFlags(fs)

	if err := envflag.Parse(fs, args, "FDS_LEDENLIJST_", lookupEnv); err != nil {
		return nil, err
	}

	return cfg, nil
}

const AppName = "FDS Ledenlijst 0.1"
