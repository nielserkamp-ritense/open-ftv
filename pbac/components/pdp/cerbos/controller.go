// Package cerbos contains all logic for a functional component acting as the Policy Decision Point
// using Cerbos/CEL as the policy language.
package cerbos

import (
	"path/filepath"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/components"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/components/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/components/pdp"
)

// Version defines the version of this Cerbos/CEL PDP.
const Version = "1.0.0"

// NewController instantiates a new Cerbos/CEL controller.
func NewController(options ...pdp.Option) pdp.Controller {
	options = append(options, pdp.WithNameVersion(components.CERBOS.String(), Version))

	c := &controller{Base: pdp.NewBase(options...)}
	c.SetPAP(pap.New(c.Context(), c.Logger(), c))

	if store, recurse := c.Store(); store != "" {
		store, _ = filepath.Abs(store)
		c.PAP().LoadFromStore(store, recurse)
	}

	c.Logger().Info("pbac controller initialized", "controller", c.String())
	return c
}

type controller struct {
	pdp.Base
}
