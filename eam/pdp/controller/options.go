package controller

import (
	"log/slog"

	"golang.org/x/net/context"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/mapping"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pep"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
)

// Option is the function signature for configuring a new controller.
type Option func(c *Base)

// WithContext adds a context to the controller.
func WithContext(ctx context.Context) Option {
	return func(c *Base) {
		c.ctx = ctx
	}
}

// WithNameVersion adds the name of the policy engine and the version to a controller.
func WithNameVersion(name, version string) Option {
	return func(c *Base) {
		c.name = name
		c.version = version
	}
}

// WithLogger adds a logger instance to the controller.
func WithLogger(logger *slog.Logger) Option {
	return func(c *Base) {
		c.logger = logger
	}
}

// WithPEP adds a Policy Enforcement Point to the controller.
func WithPEP(p pep.PEP) Option {
	return func(c *Base) {
		c.pep = p
	}
}

// WithPAP adds a Policy Administration Point to the controller.
func WithPAP(p pap.PAP) Option {
	return func(c *Base) {
		c.pap = p
	}
}

// WithPIP adds a Policy Information Point to the controller.
func WithPIP(p pip.PIP) Option {
	return func(c *Base) {
		c.pip = p
	}
}

// WithMappings adds one or more required mappings on PARC schema.
//
// Note that for mappings which target the same property,
// each match will overwrite a previous match, so the last match wins.
// E.g. if mapping1 is more important than mapping2,
// mapping1 should be supplied after mapping2: WithMappings(mapping2, mapping1).
func WithMappings(mappers ...mapping.Mapper) Option {
	return func(c *Base) {
		c.mappers = mappers
	}
}
