// Package controller contains the base logic for a component acting as a Policy Decision Point.
package controller

import (
	"context"
	"log/slog"
	"sync"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/bundles"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/mapping"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller/adl"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pep"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
)

type Evaluator interface {
	Authorize(uid string, req *models.PARC) (*models.Response, error) // must be implemented by the PDP wrapper (/evaluation).
	Batch(uid string, req *models.Batch) ([]models.Response, error)   // must be implemented by the PDP wrapper (/evaluations).
}

// Controller represents the interface for an EAM controller component,
// which embeds a PDP-wrapper for authorization requests, along with a PAP, a PIP and an optional PEP.
//
// Each PDP-wrapper, in the language-specific submodules, must implement the Controller.Evaluator interface.
// Each PDP-wrapper must also embed the Base struct to support all generic functions.
type Controller interface {
	Evaluator
	Search(uid string, req *models.PARC) ([]string, error) // implemented by Base.
	GetContext() context.Context                           // implemented by Base.
	GetLogger() *slog.Logger                               // implemented by Base.
	GetPEP() *pep.PEP                                      // implemented by Base.
	GetPAP() *pap.PAP                                      // implemented by Base.
	GetPIP() *pip.PIP                                      // implemented by Base.
	GetADL() *adl.ADL                                      // implemented by Base.
	PARCFromRequest(req *models.Request) *models.PARC      // implemented by Base (requires the PEP).
	NewBundle(bundle *bundles.Bundle) (uint64, error)      // implemented by Base.
}

// Base contains the common attributes of a controller.
type Base struct {
	Ctx           context.Context
	Logger        *slog.Logger
	Name          string
	Version       string
	PEP           *pep.PEP
	PAP           *pap.PAP
	PIP           *pip.PIP
	ADL           *adl.ADL
	AuthMutex     *sync.RWMutex
	BundleVersion uint64
	Self          Evaluator
	// hidden fields
	fullName string
	mappers  []mapping.Mapper
}

// NewBase instantiates a new controller base.
func NewBase(options ...Option) Base {
	b := Base{Ctx: context.Background(), AuthMutex: &sync.RWMutex{}}

	for i := range options {
		options[i](&b)
	}

	if b.PIP != nil && b.ADL != nil {
		pip.WithDynamicReporter(b.ADL.NewInformation)(b.PIP)
	}

	b.fullName = b.Name
	if b.Version != "" {
		b.fullName += " " + b.Version
	}

	return b
}

// String returns the full name of the controller.
func (b *Base) String() string {
	return b.fullName
}

// GetContext returns the context of the controller.
func (b *Base) GetContext() context.Context {
	return b.Ctx
}

// GetLogger returns the logger of the controller.
func (b *Base) GetLogger() *slog.Logger {
	return b.Logger
}

// GetPEP returns the PEP of the controller.
func (b *Base) GetPEP() *pep.PEP {
	return b.PEP
}

// GetPAP returns the PAP of the controller.
func (b *Base) GetPAP() *pap.PAP {
	return b.PAP
}

// GetPIP returns the PIP of the controller.
func (b *Base) GetPIP() *pip.PIP {
	return b.PIP
}

// GetADL returns the ADL of the controller.
func (b *Base) GetADL() *adl.ADL {
	return b.ADL
}

// Map performs the configured mappings on the given PARC.
func (b *Base) Map(parc *models.PARC) *models.PARC {
	p := parc
	for i := range b.mappers {
		p = b.mappers[i](p)
	}
	return p
}

// PARCFromRequest convert an authorization-request into an AuthZEN compatible PARC struct.
func (b *Base) PARCFromRequest(req *models.Request) *models.PARC {
	return b.PEP.PARCFromRequest(req, b.PIP.GetEntity)
}
