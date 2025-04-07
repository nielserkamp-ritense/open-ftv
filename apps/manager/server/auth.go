// Package server contains the HTTP request handlers for the app.
package server

import (
	"fmt"
	"log/slog"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pdp"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pdp/cedar"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pdp/cerbos"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pdp/opa"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pdp/openfga"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pep"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pip"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
)

// AuthHandler represents the interface for handling authorization requests.
type AuthHandler interface {
	Controller() pdp.Controller
}

func (s *service) newAuth() AuthHandler {
	controller, err := s.newController()
	if controller == nil {
		s.logger.Error("failed to initialize EAM controller", "error", err)
		return nil
	}
	return &authHandler{logger: s.logger, controller: controller}
}

func (s *service) newController() (pdp.Controller, error) {
	ep := pep.New(s.ctx, s.logger)
	ip := pip.New(s.ctx, s.logger, s.pipOptions()...)
	ap := pap.New(s.ctx, s.logger, s.papOptions()...)

	options := []pdp.Option{pdp.WithContext(s.ctx), pdp.WithLogger(s.logger), pdp.WithPEP(ep), pdp.WithPIP(ip), pdp.WithPAP(ap)}

	switch s.l {
	case models.CEDAR:
		return cedar.NewController(options...), nil
	case models.REGO:
		return opa.NewController(options...), nil
	case models.OPENFGA:
		return openfga.NewController(options...), nil
	case models.CERBOS:
		cerbosCFG := cerbos.Config{Addr1: s.cfg.Cerbos.Address, Addr2: s.cfg.Cerbos.AdminAddress, CA: s.cfg.Cerbos.CA, User: s.cfg.Cerbos.User, Pswd: s.cfg.Cerbos.Pswd}
		return cerbos.NewController(cerbosCFG, options...), nil
	default:
		return nil, fmt.Errorf("unsupported policy language '%s'", s.cfg.PAP.Language)
	}
}

func (s *service) pipOptions() []pip.Option {
	opts := []pip.Option{pip.WithFileStore(s.cfg.PIP.Store, s.cfg.PIP.StoreRecurse)}

	if s.cfg.PIP.PullConfigs != "" {
		opts = append(opts, pip.WithPullConfigs(s.cfg.PIP.PullConfigs))
	}

	if s.l == models.CEDAR {
		opts = append(opts, pip.WithFactories(cedar.NewAttributeBuilder(s.logger), cedar.NewEntityBuilder(s.logger)))
	}

	return opts
}

func (s *service) papOptions() []pap.Option {
	return []pap.Option{pap.WithLanguage(s.l.Language()), pap.WithFileStore(s.cfg.PAP.Store, s.cfg.PAP.StoreRecurse)}
}

// Controller returns the PDP controller.
func (h *authHandler) Controller() pdp.Controller { return h.controller }

type authHandler struct {
	logger     *slog.Logger
	controller pdp.Controller
}
