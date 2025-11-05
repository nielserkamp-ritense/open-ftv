// Package cedar_embedded contains all logic for a functional component acting as the Policy Decision Point,
// using Cedar as the policy language.
package cedar_embedded

import (
	"sync"

	"github.com/cedar-policy/cedar-go"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities-no-ci/module"
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

	c.Self = c

	if c.PIP != nil {
		// c.PIP.AddEventSink(c)
		c.PIP.IterateEntities(func(entity *models.Entity) {
			e2, err := entityToCedar(entity)
			if err != nil {
				c.Logger.Warn("failed to convert PIP entity to cedar format", "key", entity.UID(), "err", err)
			}
			c.entities[e2.UID] = *e2
		})
	}

	if c.PAP != nil {
		c.PAP.AddEventSink(c)
		c.PAP.LoadFiles()
	}

	if c.PIP != nil {
		c.PIP.AddEventSink(c)
	}

	mod := "github.com/cedar-policy/cedar-go"
	modVersion := module.GetModuleVersion(mod)

	if c.ADL != nil {
		c.ADL.NewEngine(map[string]any{
			"controller":        c.Name,
			"controllerVersion": c.Version,
			"language":          models.CEDAR.String(),
			"module":            mod,
			"moduleVersion":     modVersion,
		})
	}

	c.Logger.Info("pdp controller initialized", "controller", c.String(), "module", mod, "module-version", modVersion)
	return c
}

type controller struct {
	pdp.Base
	pdp      *cedar.PolicySet
	entities cedar.EntityMap
	pdpMutex sync.Mutex
}
