// Package cerbos contains all logic for a functional component acting as the Policy Decision Point
// using Cerbos/CEL as the policy language.
package cerbos

import (
	"log/slog"
	"path/filepath"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/components"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/components/ldv"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/components/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/components/pdp"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/components/pip"
)

// Version defines the version of this Cerbos/CEL PDP.
const Version = "1.0.0"

// NewController instantiates a new Cerbos/CEL controller.
func NewController(pip pip.PIP, store string, recurse bool, logger *slog.Logger, logboek ldv.LDV) pdp.Controller {
	store, _ = filepath.Abs(store)

	c := &controller{
		Base: pdp.NewBase(components.CERBOS.String(), Version, logger, logboek),
	}

	c.SetPIP(pip)
	c.SetPAP(pap.New(nil, c.Logger(), c))
	c.PAP().LoadFromStore(store, recurse)

	c.Logger().Info("pbac controller initialized", "controller", c.String())
	return c
}

type controller struct {
	pdp.Base
}
