package opa

import (
	"io"
	"strings"

	"github.com/open-policy-agent/opa/storage"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
)

// Handle implements the EventSink interface.
func (c *controller) Handle(event models.EventType, key string) {
	switch event {
	case models.PolicyAdded, models.PolicyReplaced:
		language, id := pap.SplitPolicyKey(key)
		if !strings.EqualFold(language, components.REGO.Language()) {
			return
		}

		f, err := c.PAP().Read(language, id)
		if err != nil {
			c.Logger().Error("failed to get policy", "controller", c.String(), "policy-id", id, "error", err)
			return
		}

		d, _ := io.ReadAll(f.Content())

		t, _ := c.mem.NewTransaction(c.Context(), storage.TransactionParams{Write: true})
		if err = c.mem.UpsertPolicy(c.Context(), t, id, d); err != nil {
			c.Logger().Error("failed to upsert policy", "controller", c.String(), "policy-id", id, "error", err)
		} else {
			if err = c.mem.Commit(c.Context(), t); err != nil {
				c.Logger().Error("failed to commit transaction", "controller", c.String(), "policy-id", id, "error", err)
			} else {
				c.Logger().Info("policy added/replaced", "controller", c.String(), "policy-id", id)
			}
		}

	case models.PolicyRemoved:
		language, id := pap.SplitPolicyKey(key)
		if !strings.EqualFold(language, components.REGO.Language()) {
			return
		}

		t, _ := c.mem.NewTransaction(c.Context(), storage.TransactionParams{Write: true})
		if err := c.mem.DeletePolicy(c.Context(), t, id); err != nil {
			c.Logger().Error("failed to remove policy", "controller", c.String(), "policy-id", id, "error", err)
		} else {
			if err = c.mem.Commit(c.Context(), t); err != nil {
				c.Logger().Error("failed to commit transaction", "controller", c.String(), "policy-id", id, "error", err)
			} else {
				c.Logger().Info("policy removed", "controller", c.String(), "policy-id", id)
			}
		}

	default:
		// TODO: attributes, entities, relations
	}
}
