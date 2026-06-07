package main

import (
	"context"
	"log/slog"
	"os"
	"sync"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pep"
)

func newConfig() any {
	return &Config{}
}

func (c *Config) ensurePEP() {
	c.once.Do(func() {
		logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
		if opt, err := pepOptionFromConfig(c); err == nil && opt != nil {
			c.pep = pep.New(context.Background(), logger, opt)
		} else {
			c.pep = pep.New(context.Background(), logger)
		}
	})
}

func newPEPWith(opt pep.Option) *pep.PEP {
	return pep.New(context.Background(), slog.New(slog.NewJSONHandler(os.Stdout, nil)), opt)
}

type Config struct {
	PDP        string `json:"pdp"`
	Timeout    int    `json:"timeout"`
	Issuer     string `json:"issuer"`
	JWKSURL    string `json:"jwks_url"`
	Audience   string `json:"audience"`
	RolesClaim string `json:"roles_claim"`
	// hidden fields
	once sync.Once
	pep  *pep.PEP
}
