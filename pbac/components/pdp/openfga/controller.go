// Package openfga contains all logic for a functional component acting as the Policy Decision Point
// using OpenFGA as the policy language.
//
// Even though OpenFGA is advertised as Relation Based Access Control,
// it provides the necessary support for RBAC and ABAC,
// so it can be made to work like any other PBAC engine.
package openfga

import (
	"log/slog"
	"path/filepath"
	"sync"

	"github.com/openfga/openfga/pkg/server"
	"github.com/openfga/openfga/pkg/storage/memory"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/components"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/components/ldv"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/components/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/components/pdp"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/components/pip"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/module"
)

// Version defines the version of this OpenFGA PDP.
const Version = "1.0.0"

// NewController instantiates a new OpenFGA controller.
func NewController(pip pip.PIP, store string, recurse bool, logger *slog.Logger, logboek ldv.LDV) pdp.Controller {
	store, _ = filepath.Abs(store)

	c := &controller{
		Base:   pdp.NewBase(components.OPENFGA.String(), Version, logger, logboek),
		stores: make(map[string]string),
		models: make(map[string]string),
	}

	if c.newServer(); c.engine == nil {
		return nil
	}

	c.SetPIP(pip)
	c.SetPAP(pap.New(nil, c.Logger(), c))
	c.PAP().LoadFromStore(store, recurse)

	mod := "github.com/openfga/openfga"
	modVersion := module.GetModuleVersion(mod)

	c.Logger().Info("pbac controller initialized", "controller", c.String(), "module", mod, "module-version", modVersion)
	return c
}

func (c *controller) newServer() {
	engine, err := server.NewServerWithOpts(
		server.WithDatastore(memory.New()),
		server.WithLogger(NewZapper(c.Logger())),
	)

	if err != nil {
		c.Logger().Error("Failed to create server", "error", err)
		return
	}

	c.engine = engine
}

type controller struct {
	pdp.Base
	engine *server.Server
	stores map[string]string // key = principal, valid = storeID.
	models map[string]string // key = storeID, value = policyID.
	mutex  sync.RWMutex
}
