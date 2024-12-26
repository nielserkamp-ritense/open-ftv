// Package cedar contains all logic for a functional component acting as the Policy Decision Point,
// using Cedar as the policy language.
package cedar

import (
	"path/filepath"

	"github.com/cedar-policy/cedar-go"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/components"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/components/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/components/pdp"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/module"
)

// Version defines the version of this Cedar PDP.
const Version = "1.0.0"

// NewController instantiates a new Cedar controller.
func NewController(options ...pdp.Option) pdp.Controller {
	options = append(options, pdp.WithNameVersion(components.CEDAR.String(), Version))
	c := &controller{Base: pdp.NewBase(options...)}

	c.pdp = cedar.NewPolicySet()
	c.entities = make(cedar.EntityMap)

	c.PIP().IterateEntities(func(entity models.Entity) {
		if wrapped, ok := entity.(*WrappedEntity); ok {
			c.entities[wrapped.ce.UID] = *wrapped.ce
		}
	})

	c.SetPAP(pap.New(c.Context(), c.Logger(), c))

	store, recurse := c.Store()
	if store != "" {
		store, _ = filepath.Abs(store)
		c.PAP().LoadFromStore(store, recurse)
	}

	mod := "github.com/cedar-policy/cedar-go"
	modVersion := module.GetModuleVersion(mod)

	c.Logger().Info("pbac controller initialized", "controller", c.String(), "module", mod, "module-version", modVersion)
	return c
}

type controller struct {
	pdp.Base
	pdp      *cedar.PolicySet
	entities cedar.EntityMap
}
