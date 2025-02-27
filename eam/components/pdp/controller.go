// Package pdp contains the base logic for a component acting as a Policy Decision Point.
package pdp

import (
	"context"
	"log/slog"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pep"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pip"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
)

// Controller represents the interface for an EAM controller component,
// which may encompass a PEP, PIP, PAP and/or PDP.
//
// Each policy language specific PDP (in the submodules) must be represented by a Controller interface.
type Controller interface {
	String() string // name & version.
	Name() string
	Version() string
	PEP() pep.PEP
	PAP() pap.PAP
	PIP() pip.PIP
	Authorize(uid string, req *models.PARC) (*models.Response, error)
}

// Base contains the common attributes of a controller.
type Base struct {
	ctx      context.Context
	logger   *slog.Logger
	name     string
	version  string
	fullName string
	pep      pep.PEP
	pap      pap.PAP
	pip      pip.PIP
}

// NewBase instantiates a new controller base.
func NewBase(options ...Option) Base {
	b := Base{ctx: context.Background()}

	for i := range options {
		options[i](&b)
	}

	b.fullName = b.name
	if b.version != "" {
		b.fullName += " " + b.version
	}

	return b
}

// String implements the Controller interface.
func (b *Base) String() string {
	return b.fullName
}

// Name implements the Controller interface.
func (b *Base) Name() string {
	return b.name
}

// Version implements the Controller interface.
func (b *Base) Version() string {
	return b.version
}

// Logger returns the logger used by the controller.
func (b *Base) Logger() *slog.Logger {
	return b.logger
}

// Context returns the context used by the controller.
func (b *Base) Context() context.Context {
	return b.ctx
}

// PEP returns the PEP used by the controller.
func (b *Base) PEP() pep.PEP {
	return b.pep
}

// PAP returns the PAP used by the controller.
func (b *Base) PAP() pap.PAP {
	return b.pap
}

// PIP returns the PIP used by the controller.
func (b *Base) PIP() pip.PIP {
	return b.pip
}
