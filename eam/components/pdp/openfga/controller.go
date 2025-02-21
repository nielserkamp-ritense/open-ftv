// Package openfga contains all logic for a functional component acting as the Policy Decision Point
// using OpenFGA as the policy language.
//
// Even though OpenFGA is advertised as Relation Based Access Control,
// it provides the necessary support for RBAC and ABAC,
// so it can be made to work like any other PDP.
package openfga

import (
	"sync"

	"github.com/openfga/openfga/pkg/server"
	"github.com/openfga/openfga/pkg/storage/memory"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pdp"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/module"
)

// Version defines the version of this OpenFGA PDP.
const Version = "1.0.0"

// NewController instantiates a new OpenFGA controller.
func NewController(options ...pdp.Option) pdp.Controller {
	options = append(options, pdp.WithNameVersion(models.OPENFGA.String(), Version))
	c := &controller{Base: pdp.NewBase(options...), stores: make(map[string]*details)}
	if c.newServer(); c.engine == nil {
		return nil
	}

	if c.PAP() != nil {
		c.PAP().AddEventSink(c)
		c.PAP().LoadFiles()
	}

	mod := "github.com/openfga/openfga"
	modVersion := module.GetModuleVersion(mod)

	c.Logger().Info("eam controller initialized", "controller", c.String(), "module", mod, "module-version", modVersion)
	return c
}

func (c *controller) newServer() {
	engine, err := server.NewServerWithOpts(
		server.WithDatastore(memory.New()),
		server.WithLogger(newZapper(c.Logger())),
	)

	if err != nil {
		c.Logger().Error("Failed to create server", "error", err)
		return
	}

	c.engine = engine
}

type controller struct {
	pdp.Base
	engine *server.Server      // OpenFGA PDP
	stores map[string]*details // key = principal type
	mutex  sync.RWMutex
}

type details struct {
	store       string              // key of the store (principal type).
	storeID     string              // id of the store in the OpenFGA engine.
	authModelID string              // id of the authorization model in the OpenFGA store.
	relations   map[string]struct{} // keys of the relations in the OpenFGA store.
}
