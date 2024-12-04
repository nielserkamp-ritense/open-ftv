// Package cedar contains all logic for a functional component acting as the Policy Decision Point,
// using Cedar as the policy language.
package cedar

import (
	"log/slog"
	"path/filepath"

	"github.com/cedar-policy/cedar-go"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/shared/control"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/shared/ldv"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/shared/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/shared/pip"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/standards"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/module"
)

// Version defines the version of this Cedar PDP.
const Version = "1.0.0"

// NewController instantiates a new Cedar controller.
func NewController(pip pip.PIP, store string, recurse bool, logger *slog.Logger, logboek ldv.LDV) control.Controller {
	if store != "" {
		store, _ = filepath.Abs(store)
	}

	c := &controller{
		Base:     control.NewBase(standards.CEDAR.String(), Version, logger, logboek),
		pdp:      cedar.NewPolicySet(),
		entities: make(cedar.EntityMap),
	}

	pip.IterateEntities(func(entity standards.Entity) {
		if wrapped, ok := entity.(*WrappedEntity); ok {
			c.entities[wrapped.ce.UID] = *wrapped.ce
		}
	})

	c.SetPIP(pip)
	c.SetPAP(pap.New(nil, c.Logger(), c))
	c.PAP().LoadFromStore(store, recurse)

	mod := "github.com/cedar-policy/cedar-go"
	modVersion := module.GetModuleVersion(mod)

	c.Logger().Info("pbac controller initialized", "controller", c.String(), "module", mod, "module-version", modVersion)
	return c
}

type controller struct {
	control.Base
	pdp      *cedar.PolicySet
	entities cedar.EntityMap
}
