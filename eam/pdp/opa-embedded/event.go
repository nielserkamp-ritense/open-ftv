package opa_embedded

import (
	"io"
	"strings"

	"github.com/open-policy-agent/opa/storage"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

// Handle implements the EventSink interface.
func (c *controller) Handle(event models.EventType, key string) {
	switch event {
	case models.PolicyAdded, models.PolicyReplaced:
		language, id := models.SplitPolicyKey(key)
		if !strings.EqualFold(language, models.REGO.Language()) && !strings.EqualFold(language, models.REGO.String()) {
			return
		}

		f, _, err := c.PAP.Read(language, id)
		if err != nil {
			c.Logger.Error("failed to get policy", "controller", c.String(), "policy-id", id, "error", err)
			return
		}

		d, _ := io.ReadAll(f.Content())

		c.pdpMutex.Lock()
		defer c.pdpMutex.Unlock()

		t, _ := c.mem.NewTransaction(c.Ctx, storage.TransactionParams{Write: true})
		if err = c.mem.UpsertPolicy(c.Ctx, t, id, d); err != nil {
			c.Logger.Error("failed to upsert policy", "controller", c.String(), "policy-id", id, "error", err)
		} else {
			if err = c.mem.Commit(c.Ctx, t); err != nil {
				c.Logger.Error("failed to commit transaction", "controller", c.String(), "policy-id", id, "error", err)
			} else {
				c.Logger.Info("policy added/replaced", "controller", c.String(), "policy-id", id)
			}
		}

	case models.PolicyRemoved:
		language, id := models.SplitPolicyKey(key)
		if !strings.EqualFold(language, models.REGO.Language()) {
			return
		}

		c.pdpMutex.Lock()
		defer c.pdpMutex.Unlock()

		t, _ := c.mem.NewTransaction(c.Ctx, storage.TransactionParams{Write: true})
		if err := c.mem.DeletePolicy(c.Ctx, t, id); err != nil {
			c.Logger.Error("failed to remove policy", "controller", c.String(), "policy-id", id, "error", err)
		} else {
			if err = c.mem.Commit(c.Ctx, t); err != nil {
				c.Logger.Error("failed to commit transaction", "controller", c.String(), "policy-id", id, "error", err)
			} else {
				c.Logger.Info("policy removed", "controller", c.String(), "policy-id", id)
			}
		}

	default:
		// TODO: attributes, entities, relations
	}
}
