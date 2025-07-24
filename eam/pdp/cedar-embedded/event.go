package cedar_embedded

import (
	"io"
	"strings"

	"github.com/cedar-policy/cedar-go"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

// Handle implements the EventSink interface.
func (c *controller) Handle(t models.EventType, key string) {
	switch t {
	case models.PolicyAdded, models.PolicyReplaced:
		language, id := models.SplitPolicyKey(key)
		if !strings.EqualFold(language, models.CEDAR.Language()) && !strings.EqualFold(language, models.CEDAR.String()) {
			return
		}

		f, _, err := c.PAP().Read(language, id)
		if err != nil {
			c.Logger().Error("failed to get policy", "controller", c.String(), "policy-id", id, "error", err)
			return
		}

		d, _ := io.ReadAll(f.Content())

		var policy cedar.Policy
		if err = policy.UnmarshalCedar(d); err == nil {
			c.pdp.Add(cedar.PolicyID(id), &policy)
			c.Logger().Info("policy added/replaced", "controller", c.String(), "policy-id", id)
		} else {
			c.Logger().Error("error decoding policy", "controller", c.String(), "policy-id", id, "error", err)
		}

	case models.PolicyRemoved:
		language, id := models.SplitPolicyKey(key)
		if !strings.EqualFold(language, models.CEDAR.String()) {
			return
		}

		c.pdp.Remove(cedar.PolicyID(id))
		c.Logger().Info("policy removed", "controller", c.String(), "policy-id", id)

	default:
		// TODO: attributes, entities, relations
	}
}
