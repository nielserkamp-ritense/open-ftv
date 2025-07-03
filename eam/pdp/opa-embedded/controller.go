// Package opa_embedded contains all logic for a functional component acting as the Policy Decision Point
// using OPA/Rego as the policy language.
package opa_embedded

import (
	"bytes"
	"context"

	"github.com/open-policy-agent/opa/hooks"
	"github.com/open-policy-agent/opa/sdk"
	"github.com/open-policy-agent/opa/storage"
	"github.com/open-policy-agent/opa/storage/inmem"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller"
)

// Version defines the version of this OPA/Rego PDP.
const Version = "1.0.0"

// NewController instantiates a new OPA/Rego controller.
func NewController(options ...pdp.Option) pdp.Controller {
	wait := make(chan struct{})

	options = append(options, pdp.WithNameVersion(models.REGO.String(), Version))
	c := &controller{Base: pdp.NewBase(options...), mem: inmem.New()}

	var err error
	c.pdp, err = sdk.New(context.Background(), sdk.Options{
		RegoVersion:   1,
		ID:            "opa-controller",
		Config:        bytes.NewReader([]byte(cfg)),
		ConsoleLogger: &wrappedLogger{logger: c.Logger()},
		Ready:         wait,
		Hooks:         hooks.Hooks{},
		Store:         c.mem,
	})
	if err != nil {
		c.Logger().Error("Failed to initialize OPA SDK", "error", err)
		return nil
	}

	// wait for OPA to be ready, before loading other data!
	select {
	case <-wait:
	}

	c.loadAttributes()
	c.loadEntities()

	if c.PAP() != nil {
		c.PAP().AddEventSink(c)
		c.PAP().LoadFiles()
	}

	c.Logger().Info("pdp controller initialized", "controller", c.String())
	return c
}

func (c *controller) loadAttributes() {
	m := make(map[string]any)
	c.PIP().IterateAttributes(func(attr models.Attribute) {
		m[attr.Key()] = attr.Value()
	})
	c.loadData(m, "attributes")
}

func (c *controller) loadEntities() {
	m := make(map[string]any)
	c.PIP().IterateEntities(func(entity models.Entity) {
		m2, ok := m[entity.Type()].(map[string]any)
		if !ok || m2 == nil {
			m2 = make(map[string]any)
		}
		m2[entity.ID()] = models.MapFromAttributes(entity.Attributes())
		m[entity.Type()] = m2
	})
	c.loadData(m, "entities")
}

func (c *controller) loadData(m map[string]any, key string) {
	if len(m) == 0 {
		return
	}

	t, _ := c.mem.NewTransaction(c.Context(), storage.TransactionParams{Write: true})
	if err := c.mem.Write(c.Context(), t, storage.AddOp, storage.Path{key}, m); err != nil {
		c.Logger().Error("failed to add/replace data", "controller", c.String(), "document-key", key, "error", err)
	}
	if err := c.mem.Commit(c.Context(), t); err != nil {
		c.Logger().Error("failed to commit transaction to add/replace data", "controller", c.String(), "document-key", key, "error", err)
	} else {
		c.Logger().Info("data added/replaced successfully", "controller", c.String(), "document-key", key)
	}
}

type controller struct {
	pdp.Base
	pdp *sdk.OPA
	mem storage.Store
	m   map[string]any
}

const cfg = `{
	"decision_logs": {
		"console": true
	}
}`
