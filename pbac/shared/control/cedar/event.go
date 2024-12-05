package cedar

import (
	"io"

	"github.com/cedar-policy/cedar-go"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/models"
)

// Handle implements the EventSink interface.
func (c *controller) Handle(t standards.EventType, key string) {
	switch t {
	case standards.PolicyAdded, standards.PolicyReplaced:
		f, err := c.PAP().Get(key)
		if err != nil {
			c.Logger().Error("failed to get policy", "controller", c.String(), "policy-key", key, "error", err)
			return
		}

		d, _ := io.ReadAll(f)

		var policy cedar.Policy
		if err = policy.UnmarshalCedar(d); err == nil {
			c.pdp.Add(cedar.PolicyID(key), &policy)
			c.Logger().Info("policy added/replaced", "controller", c.String(), "policy-key", key)
		} else {
			c.Logger().Error("error decoding policy", "controller", c.String(), "policy-key", key, "error", err)
		}

	case standards.PolicyRemoved:
		c.pdp.Remove(cedar.PolicyID(key))
		c.Logger().Info("policy removed", "controller", c.String(), "policy-key", key)
	}
}
