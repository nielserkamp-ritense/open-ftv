// Package openfga contains all logic for a functional component acting as the Policy Decision Point
// using OpenFGA as the policy language.
//
// Even though OpenFGA is advertised as Relation Based Access Control,
// it provides the necessary support for RBAC and ABAC,
// so it can be made to work like any other PBAC engine.
package openfga

import (
	"context"
	"log/slog"
	"path/filepath"

	openfgav1 "github.com/openfga/api/proto/openfga/v1"
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

	c := &controller{Base: pdp.NewBase(components.OPENFGA.String(), Version, logger, logboek)}

	if c.newServer(); c.pdp == nil {
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
	pdp, err := server.NewServerWithOpts(
		server.WithDatastore(memory.New()),
		server.WithLogger(NewZapper(c.Logger())),
	)

	if err != nil {
		c.Logger().Error("Failed to create server", "error", err)
		return
	}

	store, err2 := pdp.CreateStore(
		context.Background(),
		&openfgav1.CreateStoreRequest{Name: "demo"},
	)
	if err2 != nil {
		c.Logger().Error("Failed to create store", "error", err2)
		return
	}

	c.pdp = pdp
	c.storeID = store.GetId()
}

type controller struct {
	pdp.Base
	pdp     *server.Server
	storeID string
}
