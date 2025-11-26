package main

import (
	"log/slog"
	"os"
	"time"

	"gitlab.com/gjuyn/go-config/config"
	"gitlab.com/gjuyn/go-config/config-ext/yaml"
)

type cfg struct {
	Host    string        `json:"host"    yaml:"svc.host,omitempty"    env:"ADDRESS,HOST" flag:"address,host,a,h" desc:"Address to use for service (default 0.0.0.0)"                              default:"0.0.0.0"`
	Port    uint16        `json:"port"    yaml:"svc.port,omitempty"    env:"PORT"         flag:"port,p"           desc:"Port to use for service (default 3001 )"                                   default:"3001"`
	PDP     string        `json:"pdp"     yaml:"pdp.address,omitempty" env:"PDP_ADDRESS"  flag:"pdp_address"      desc:"Address of the PDP (default https://localhost:8443/authzen/v1/evaluation)" default:"https://localhost:8443/authzen/v1/evaluation"`
	Timeout time.Duration `json:"timeout" yaml:"pdp.timeout,omitempty" env:"PDP_TIMEOUT"  flag:"pdp_timeout"      desc:"Timeout for accessing the PDP (default 5 seconds)"                         default:"5s"`
}

var paths = []string{
	"/etc/envoy/authzen.yaml",
	"./etc/authzen.yaml",
	"./authzen.yaml",
}

func newConfig() (*cfg, *slog.Logger) {
	c := &cfg{}
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	opts := []config.Option{
		yaml.FilesYAML(paths...),
		config.EnvironmentPrefix("ENVOY_AUTHZEN_"),
		config.AppName("Envoy AuthZEN plugin 1.0"),
		config.NoHelpOnError(),
	}

	if err := config.LoadConfig(c, opts...); err != nil {
		logger.Error("failed to load configuration", "error", err)
		panic("service halted")
	}

	logger.Info("configuration loaded", "config", c)

	return c, logger
}
