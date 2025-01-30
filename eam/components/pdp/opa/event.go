package opa

import (
	"io"

	"github.com/open-policy-agent/opa/storage"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
)

// Handle implements the EventSink interface.
func (c *controller) Handle(event models.EventType, key string) {
	switch event {
	case models.PolicyAdded, models.PolicyReplaced:
		f, err := c.PAP().Get(key)
		if err != nil {
			c.Logger().Error("failed to get policy", "controller", c.String(), "policy-key", key, "error", err)
			return
		}

		d, _ := io.ReadAll(f.Content())

		t, _ := c.mem.NewTransaction(c.Context(), storage.TransactionParams{Write: true})
		if err = c.mem.UpsertPolicy(c.Context(), t, key, d); err != nil {
			c.Logger().Error("failed to upsert policy", "controller", c.String(), "policy-key", key, "error", err)
		} else {
			if err = c.mem.Commit(c.Context(), t); err != nil {
				c.Logger().Error("failed to commit transaction", "controller", c.String(), "policy-key", key, "error", err)
			} else {
				c.Logger().Info("policy added/replaced", "controller", c.String(), "policy-key", key)
			}
		}

	case models.PolicyRemoved:
		t, _ := c.mem.NewTransaction(c.Context(), storage.TransactionParams{Write: true})
		if err := c.mem.DeletePolicy(c.Context(), t, key); err != nil {
			c.Logger().Error("failed to remove policy", "controller", c.String(), "policy-key", key, "error", err)
		} else {
			if err = c.mem.Commit(c.Context(), t); err != nil {
				c.Logger().Error("failed to commit transaction", "controller", c.String(), "policy-key", key, "error", err)
			} else {
				c.Logger().Info("policy removed", "controller", c.String(), "policy-key", key)
			}
		}

	default:
		// TODO: attributes, entities, relations
	}
}
