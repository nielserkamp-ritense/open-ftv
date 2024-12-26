package pdp

import (
	"log/slog"

	"golang.org/x/net/context"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/components/ldv"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/components/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/components/pip"
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

// WithLogboek adds an interface for the Logboek Dataverwerkingen to the controller.
func WithLogboek(ldv ldv.LDV) Option {
	return func(c *Base) {
		c.logboek = ldv
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

// WithStore adds a file storage location to the controller.
func WithStore(store string, recurse bool) Option {
	return func(c *Base) {
		c.store = store
		c.recurse = recurse
	}
}
