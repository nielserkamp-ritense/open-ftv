// Package cedar contains all logic for a functional component acting as the Policy Decision Point,
// using Cedar as the policy language.
package cedar

import (
	"github.com/cedar-policy/cedar-go"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pdp"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/module"
)

// Version defines the version of this Cedar PDP.
const Version = "1.0.0"

// NewController instantiates a new Cedar controller.
func NewController(options ...pdp.Option) pdp.Controller {
	options = append(options, pdp.WithNameVersion(models.CEDAR.String(), Version))

	c := &controller{
		Base:     pdp.NewBase(options...),
		pdp:      cedar.NewPolicySet(),
		entities: make(cedar.EntityMap),
	}

	if c.PIP() != nil {
		c.PIP().IterateEntities(func(entity models.Entity) {
			if wrapped, ok := entity.(*WrappedEntity); ok {
				c.entities[wrapped.ce.UID] = *wrapped.ce
			} else {
				e2 := entityToCedar(entity)
				c.entities[e2.UID] = *e2
			}
		})
	}

	if c.PAP() != nil {
		c.PAP().AddEventSink(c)
		c.PAP().LoadFiles()
	}

	mod := "github.com/cedar-policy/cedar-go"
	modVersion := module.GetModuleVersion(mod)

	c.Logger().Info("pdp controller initialized", "controller", c.String(), "module", mod, "module-version", modVersion)
	return c
}

type controller struct {
	pdp.Base
	pdp      *cedar.PolicySet
	entities cedar.EntityMap
}
