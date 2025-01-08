package cerbos

import (
	"fmt"

	"github.com/cerbos/cerbos-sdk-go/cerbos"
	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/models"
)

// Handle implements the EventSink interface.
func (c *controller) Handle(event models.EventType, key string) {
	switch event {
	case models.PolicyAdded, models.PolicyReplaced:
		c.policyIDs[key] = c.getPolicyID(key)

		policy, err := c.PAP().Get(key)
		if err != nil {
			c.Logger().Error("failed to get policy", c.args(err, "policy-key", key)...)
			return
		}

		set := cerbos.NewPolicySet().AddPolicyFromReader(policy.Content())
		if err = set.Validate(); err != nil {
			c.Logger().Error("failed to decode policy", c.args(err, "policy-key", key)...)
		} else if err = c.admin.AddOrUpdatePolicy(c.Context(), set); err != nil {
			c.Logger().Error("failed to add/replace policy", c.args(err, "policy-key", key)...)
		} else {
			c.Logger().Info("policy added/replaced", c.args(nil, "policy-key", key)...)
		}

	case models.PolicyRemoved:
		if id, ok := c.policyIDs[key]; ok && id != "" {
			if _, err := c.admin.DisablePolicy(c.Context(), id); err != nil {
				c.Logger().Error("failed to remove policy", c.args(err, "policy-key", key, "policy-id", id)...)
			} else {
				delete(c.policyIDs, key)
				c.Logger().Info("policy removed", c.args(nil, "policy-key", key, "policy-id", id)...)
			}
		}

	default:
		// TODO: attributes, entities, relations
	}
}

func (c *controller) getPolicyID(key string) string {
	policy, err := c.PAP().Get(key)
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
					return fmt.Sprintf("resource.%s.v%s/%s", r, v, s)
				}
			}
		}
	}

	if rp, ok := m["principalPolicy"].(map[string]any); ok {
		if p, ok2 := rp["principal"].(string); ok2 {
			if v, ok3 := rp["version"].(string); ok3 {
				if s, ok4 := rp["scope"].(string); ok4 {
					return fmt.Sprintf("principal.%s.v%s/%s", p, v, s)
				}
			}
		}
	}

	if rp, ok := m["rolePolicy"].(map[string]any); ok {
		if r, ok2 := rp["role"].(string); ok2 {
			if s, ok3 := rp["scope"].(string); ok3 {
				return fmt.Sprintf("role.%s/%s", r, s)
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
