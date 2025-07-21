// Package pdp contains the base logic for a component acting as a Policy Decision Point.
package controller

import (
	"context"
	"log/slog"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/mapping"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pep"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
)

// Controller represents the interface for an EAM controller component,
// which may encompass a PEP, PIP, PAP and/or PDP.
//
// Each policy language specific PDP (in the submodules) must be represented by a Controller interface.
type Controller interface {
	String() string // name & version.
	Name() string
	Version() string
	Context() context.Context
	Logger() *slog.Logger
	PEP() *pep.PEP
	PAP() *pap.PAP
	PIP() *pip.PIP
	Authorize(uid string, req *models.PARC) (*models.Response, error)
}

// Base contains the common attributes of a controller.
type Base struct {
	ctx      context.Context
	logger   *slog.Logger
	name     string
	version  string
	fullName string
	pep      *pep.PEP
	pap      *pap.PAP
	pip      *pip.PIP
	mappers  []mapping.Mapper
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

// Context returns the context used by the controller.
func (b *Base) Context() context.Context {
	return b.ctx
}

// Logger returns the logger used by the controller.
func (b *Base) Logger() *slog.Logger {
	return b.logger
}

// PEP returns the PEP used by the controller.
func (b *Base) PEP() *pep.PEP {
	return b.pep
}

// PAP returns the PAP used by the controller.
func (b *Base) PAP() *pap.PAP {
	return b.pap
}

// PIP returns the PIP used by the controller.
func (b *Base) PIP() *pip.PIP {
	return b.pip
}

// Map performs the configured mappings on the given PARC.
func (b *Base) Map(parc *models.PARC) *models.PARC {
	p := parc
	for i := range b.mappers {
		p = b.mappers[i](p)
	}
	return p
}
