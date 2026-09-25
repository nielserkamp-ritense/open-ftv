package main

import (
	"flag"
	"time"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/envflag"
)

const (
	defaultPort       = 3001
	defaultPDPTimeout = 5 * time.Second
)

type cfg struct {
	Host    string
	Port    uint16
	PDP     string
	Timeout time.Duration
}

func loadConfig(args []string, lookupEnv func(string) (string, bool)) (*cfg, error) {
	c := &cfg{}

	fs := flag.NewFlagSet("openftv", flag.ContinueOnError)
	fs.StringVar(&c.Host, "host", "0.0.0.0", "Address to use for service")
	envflag.Uint16Var(fs, &c.Port, "port", defaultPort, "Port to use for service")
	fs.StringVar(&c.PDP, "pdp-address", "https://localhost:8443/authzen/v1/evaluation", "Address of the PDP")
	fs.DurationVar(&c.Timeout, "pdp-timeout", defaultPDPTimeout, "Timeout for accessing the PDP")

	if err := envflag.Parse(fs, args, "ENVOY_AUTHZEN_", lookupEnv); err != nil {
		return nil, err
	}

	return c, nil
}
