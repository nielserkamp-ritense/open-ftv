package cerbos_api

import (
	"fmt"
	"strings"

	"github.com/cerbos/cerbos-sdk-go/cerbos"
	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

// Handle implements the EventSink interface.
func (c *controller) Handle(event models.EventType, key string) {
	if c.admin == nil {
		c.logger.Error("failed to process event; admin client not initialized", "policy-id", key)
		return
	}

	switch event {
	case models.PolicyAdded, models.PolicyReplaced:
		language, id := models.SplitPolicyKey(key)
		if !strings.EqualFold(language, models.CERBOS.Language()) && !strings.EqualFold(language, models.CERBOS.String()) {
			return
		}

		c.pdpMutex.Lock()
		defer c.pdpMutex.Unlock()

		id2 := c.getPolicyID(language, id)

		oldID, ok := c.policyIDs[id]
		if ok && oldID != id2 {
			c.deletePolicy(key, oldID)
		}

		c.upsertPolicy(language, id, id2)

	case models.PolicyRemoved:
		language, id := models.SplitPolicyKey(key)
		if !strings.EqualFold(language, models.CERBOS.Language()) && !strings.EqualFold(language, models.CERBOS.String()) {
			return
		}

		c.pdpMutex.Lock()
		defer c.pdpMutex.Unlock()

		if id2, ok := c.policyIDs[id]; ok && id2 != "" {
			c.deletePolicy(id, id2)
		}

	default:
		// TODO: attributes, entities, relations
	}
}

func (c *controller) upsertPolicy(language, id, id2 string) {
	policy, _, err := c.PAP.Read(id)
	if err != nil {
		c.logger.Error("failed to get policy", "policy-id", id, "error", err)
		return
	}

	set := cerbos.NewPolicySet().AddPolicyFromReader(policy.Content())
	if err = set.Validate(); err != nil {
		c.logger.Error("failed to decode policy", "policy-id", id, "error", err)
	} else if err = c.admin.AddOrUpdatePolicy(c.Ctx, set); err != nil {
		c.logger.Error("failed to add/replace policy", "policy-id", id, "error", err)
	} else {
		c.policyIDs[id] = id2
		c.logger.Info("policy added/replaced", "policy-id", id, "policy-key", id2)
	}
}

func (c *controller) deletePolicy(id, id2 string) {
	if _, err := c.admin.DisablePolicy(c.Ctx, id2); err != nil {
		c.logger.Error("failed to remove policy", "policy-id", id, "policy-key", id2, "error", err)
	} else {
		delete(c.policyIDs, id)
		c.logger.Info("policy removed", "policy-id", id, "policy-key", id2)
	}
}

func (c *controller) getPolicyID(language, id string) string {
	policy, _, err := c.PAP.Read(id)
	if err != nil {
		return ""
	}

	var m map[string]any
	if err = yaml.NewDecoder(policy.Content()).Decode(&m); err != nil {
		if err = json.NewDecoder(policy.Content()).Decode(&m); err != nil {
			return ""
		}
	}

	if rp, ok := m["resourcePolicy"].(map[string]any); ok {
		if r, ok2 := rp["resource"].(string); ok2 {
			if v, ok3 := rp["version"].(string); ok3 {
				if s, ok4 := rp["scope"].(string); ok4 {
					return fmt.Sprintf("resource.%s.v%s@%s", r, v, s)
				} else {
					return fmt.Sprintf("resource.%s.v%s@default", r, v)
				}
			}
		}
	}

	if rp, ok := m["principalPolicy"].(map[string]any); ok {
		if p, ok2 := rp["principal"].(string); ok2 {
			if v, ok3 := rp["version"].(string); ok3 {
				if s, ok4 := rp["scope"].(string); ok4 {
					return fmt.Sprintf("principal.%s.v%s@%s", p, v, s)
				} else {
					return fmt.Sprintf("principal.%s.v%s@default", p, v)
				}
			}
		}
	}

	if rp, ok := m["rolePolicy"].(map[string]any); ok {
		if r, ok2 := rp["role"].(string); ok2 {
			if s, ok3 := rp["scope"].(string); ok3 {
				return fmt.Sprintf("role.%s@%s", r, s)
			} else {
				return fmt.Sprintf("role.%s@default", r)
			}
		}
	}

	if rp, ok := m["exportVariables"].(map[string]any); ok {
		if n, ok2 := rp["name"].(string); ok2 {
			return fmt.Sprintf("export_variables.%s", n)
		}
	}

	if rp, ok := m["exportConstants"].(map[string]any); ok {
		if n, ok2 := rp["name"].(string); ok2 {
			return fmt.Sprintf("export_constants.%s", n)
		}
	}

	return ""
}
