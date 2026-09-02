package config

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

// SelfAuthzControllerBuilder builds the embedded PDP that authorizes an app's own HTTP API.
//
// Every app needs a self-authz PIP and, for CERBOS, connection details, so those are required
// up front. What varies per app is the PAP it enforces from and any extra pdp.Option it needs;
// WithPAP/WithBundledPAP and WithOptions configure those, in any order, before Build.
type SelfAuthzControllerBuilder struct {
	ctx       context.Context
	logger    *slog.Logger
	language  models.Language
	pipCfg    *PIP
	cerbosCfg *Cerbos
	pap       *pap.PAP
	papCfg    *PAP
	opts      []pdp.Option
}

// BuildSelfAuthzController starts building the embedded PDP that authorizes this app's own
// HTTP API for language, using pipCfg for its self-authz PIP and cerbosCfg when language is
// CERBOS. Call WithPAP or WithBundledPAP before Build to supply the PAP it enforces from.
func BuildSelfAuthzController(
	ctx context.Context, logger *slog.Logger, language models.Language, pipCfg *PIP, cerbosCfg *Cerbos,
) *SelfAuthzControllerBuilder {
	return &SelfAuthzControllerBuilder{ctx: ctx, logger: logger, language: language, pipCfg: pipCfg, cerbosCfg: cerbosCfg}
}

// WithPAP enforces the given live, administered PAP, e.g. the same one an app's own admin API
// manages (pap, manager). Overrides any WithBundledPAP already set.
func (b *SelfAuthzControllerBuilder) WithPAP(ap *pap.PAP) *SelfAuthzControllerBuilder {
	b.pap, b.papCfg = ap, nil
	return b
}

// WithBundledPAP builds a private PAP from papCfg's bundled attribute/policy files, for an app
// with no live PAP of its own (pip, pdp). Overrides any WithPAP already set.
func (b *SelfAuthzControllerBuilder) WithBundledPAP(papCfg *PAP) *SelfAuthzControllerBuilder {
	b.papCfg, b.pap = papCfg, nil
	return b
}

// WithOptions appends extra pdp.Option, applied last so they can override any default, e.g. a
// caller-specific pdp.WithPEP or pdp.WithADL.
func (b *SelfAuthzControllerBuilder) WithOptions(opts ...pdp.Option) *SelfAuthzControllerBuilder {
	b.opts = append(b.opts, opts...)
	return b
}

// Build assembles the controller from the configuration collected so far.
func (b *SelfAuthzControllerBuilder) Build() (pdp.Controller, *pip.PIP, error) {
	ap := b.pap

	if ap == nil && b.papCfg != nil {
		var err error
		if ap, err = b.papCfg.NewSelfAuthzPAP(b.ctx, b.logger); err != nil {
			return nil, nil, err
		}
	}

	ip, err := b.pipCfg.NewSelfAuthzPIP(b.ctx, b.logger, b.language)
	if err != nil {
		return nil, nil, err
	}

	options := append([]pdp.Option{
		pdp.WithContext(b.ctx),
		pdp.WithLogger(b.logger),
		pdp.WithPEP(pep.New(b.ctx, b.logger)),
		pdp.WithPIP(ip),
		pdp.WithPAP(ap),
	}, b.opts...)

	switch b.language {
	case models.CEDAR:
		return cedar_embedded.NewController(options...), ip, nil
	case models.REGO:
		return opa_embedded.NewController(options...), ip, nil
	case models.OPENFGA:
		return openfga_embedded.NewController(options...), ip, nil
	case models.CERBOS:
		cfg := cerbos_api.Config{Addr1: b.cerbosCfg.Address, Addr2: b.cerbosCfg.AdminAddress, CA: b.cerbosCfg.CA, User: b.cerbosCfg.User, Pswd: b.cerbosCfg.Pswd}
		return cerbos_api.NewController(cfg, options...), ip, nil
	default:
		return nil, nil, fmt.Errorf("unsupported policy language '%s'", b.language)
	}
}
