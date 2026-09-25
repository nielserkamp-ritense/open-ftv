// Package selfauthz builds the embedded PDP that authorizes an app's own HTTP API.
package selfauthz

import (
	"context"
	"fmt"
	"log/slog"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	cedar_embedded "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/cedar-embedded"
	cerbos_api "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/cerbos-api"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller"
	opa_embedded "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/opa-embedded"
	openfga_embedded "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/openfga-embedded"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pep"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
)

// ControllerBuilder builds the embedded PDP that authorizes an app's own HTTP API.
//
// Every app needs a self-authz PIP and, for CERBOS, connection details, so those are required
// up front. What varies per app is the PAP it enforces from and any extra pdp.Option it needs;
// WithPAP/WithBundledPAP and WithOptions configure those, in any order, before Build.
type ControllerBuilder struct {
	ctx      context.Context
	logger   *slog.Logger
	language models.Language
	pip      *pip.PIP
	cerbos   *cerbos_api.Config
	pap      *pap.PAP
	bundled  *bundledPAP
	opts     []pdp.Option
}

type bundledPAP struct {
	language string
	store    string
	recurse  bool
}

// BuildController starts building the embedded PDP that authorizes this app's own HTTP API for
// language, backed by ip and using cerbos when language is CERBOS. Call WithPAP or WithBundledPAP
// before Build to supply the PAP it enforces from.
func BuildController(
	ctx context.Context, logger *slog.Logger, language models.Language, ip *pip.PIP, cerbos *cerbos_api.Config,
) *ControllerBuilder {
	return &ControllerBuilder{ctx: ctx, logger: logger, language: language, pip: ip, cerbos: cerbos}
}

// WithPAP enforces the given live, administered PAP, e.g. the same one an app's own admin API
// manages (pap, manager). Overrides any WithBundledPAP already set.
func (b *ControllerBuilder) WithPAP(ap *pap.PAP) *ControllerBuilder {
	b.pap, b.bundled = ap, nil
	return b
}

// WithBundledPAP builds a private PAP from the bundled policy files in store, for an app with no
// live PAP of its own (pip, pdp). Overrides any WithPAP already set.
func (b *ControllerBuilder) WithBundledPAP(language, store string, recurse bool) *ControllerBuilder {
	b.bundled, b.pap = &bundledPAP{language: language, store: store, recurse: recurse}, nil
	return b
}

// WithOptions appends extra pdp.Option, applied last so they can override any default, e.g. a
// caller-specific pdp.WithPEP or pdp.WithADL.
func (b *ControllerBuilder) WithOptions(opts ...pdp.Option) *ControllerBuilder {
	b.opts = append(b.opts, opts...)
	return b
}

// Build assembles the controller from the configuration collected so far.
func (b *ControllerBuilder) Build() (pdp.Controller, error) {
	ap := b.pap

	if ap == nil && b.bundled != nil {
		var err error
		if ap, err = pap.NewSelfAuthz(b.ctx, b.logger, b.bundled.language, b.bundled.store, b.bundled.recurse); err != nil {
			return nil, err
		}
	}

	options := append([]pdp.Option{
		pdp.WithContext(b.ctx),
		pdp.WithLogger(b.logger),
		pdp.WithPEP(pep.New(b.ctx, b.logger)),
		pdp.WithPIP(b.pip),
		pdp.WithPAP(ap),
	}, b.opts...)

	switch b.language {
	case models.CEDAR:
		return cedar_embedded.NewController(options...), nil
	case models.REGO:
		return opa_embedded.NewController(options...), nil
	case models.OPENFGA:
		return openfga_embedded.NewController(options...), nil
	case models.CERBOS:
		return cerbos_api.NewController(*b.cerbos, options...), nil
	default:
		return nil, fmt.Errorf("unsupported policy language '%s'", b.language)
	}
}
