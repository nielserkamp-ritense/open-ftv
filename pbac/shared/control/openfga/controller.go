// Package openfga contains all logic for a functional component acting as the Policy Decision Point
// using OpenFGA as the policy language.
//
// Even though OpenFGA is advertised as Relation Based Access Control,
// it contains the necessary support for RBAC and ABAC, so it can work like any other PBAC engine.
package openfga

import (
	"log/slog"
	"path/filepath"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/fsc/plugin/generic/ldv"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/shared/control"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/shared/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/shared/pip"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/shared/types"
)

// Version defines the version of this OpenFGA PDP.
const Version = "1.0.0"

// NewController instantiates a new OpenFGA controller.
func NewController(pip pip.PIP, store string, recurse bool, logger *slog.Logger, ldv ldv.LDV) control.Controller {
	store, _ = filepath.Abs(store)

	c := &controller{
		Base: control.NewBase(types.OPENFGA.String(), Version, logger, ldv),
	}

	c.SetPIP(pip)
	c.SetPAP(pap.New(nil, c.Logger(), c))
	c.PAP().LoadFromStore(store, recurse)

	c.Logger().Info("pbac controller initialized", "controller", c.String())
	return c
}

type controller struct {
	control.Base
}
