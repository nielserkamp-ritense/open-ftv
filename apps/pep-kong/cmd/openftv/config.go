package main

import (
	"context"
	"log/slog"
	"os"
	"sync"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pep"
)

type Config struct {
	PDP string `json:"pdp"`
	// Timeout is in milliseconds; without one, the PDP call gets defaultPDPTimeout.
	Timeout    int    `json:"timeout"`
	Issuer     string `json:"issuer"`
	JWKSURL    string `json:"jwks_url"`
	Audience   string `json:"audience"`
	RolesClaim string `json:"roles_claim"`
	// EvaluateResponse also asks the PDP for a decision on the upstream response.
	EvaluateResponse bool `json:"evaluate_response"`
	// hidden fields
	once sync.Once
	pep  *pep.PEP
}

func newConfig() any {
	return &Config{}
}

func (c *Config) ensurePEP() {
	c.once.Do(func() {
		logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

		opt, err := pepOptionFromConfig(c)
		if err != nil {
			logger.Error("OpenFTV AuthZEN: JWT validation is off, so bearer tokens are refused", "error", err)
		}

		if opt != nil {
			c.pep = pep.New(context.Background(), logger, opt)
		} else {
			c.pep = pep.New(context.Background(), logger)
		}
	})
}
