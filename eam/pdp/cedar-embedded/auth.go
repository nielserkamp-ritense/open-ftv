package cedar_embedded

import (
	"log/slog"
	"time"

	"github.com/cedar-policy/cedar-go"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

// Authorize implements the Controller interface.
func (c *controller) Authorize(uid string, req *models.PARC) (*models.Response, error) {
	debug := c.Logger().Enabled(nil, slog.LevelDebug)
	if debug {
		c.Logger().Debug("authorization request", "controller", c.String(), "request-uid", uid)
	}

	req2 := c.buildCedarRequest(req)

	started := time.Now()
	decision, diagnostic := c.pdp.IsAuthorized(c.entities, req2)
	duration := time.Since(started)

	if decision {
		if debug {
			c.Logger().Debug("authorization granted", "controller", c.String(), "request-uid", uid, "pdp elapsed", duration.String())
		}
		return &models.Response{Allowed: true}, nil
	}

	if debug {
		c.Logger().Warn("authorization failed", "controller", c.String(), "request-uid", uid, "diagnostic", diagnostic, "pdp elapsed", duration.String())
	}

	return &models.Response{Allowed: false, Message: "not authorized", Attributes: map[string]any{"diagnostic": diagnostic}}, nil
}

func (c *controller) buildCedarRequest(parc *models.PARC) cedar.Request {
	parc = c.Map(parc)

	ca, ok := parc.Context.(*attributes)
	if !ok {
		a := NewAttributeSet(c.Logger(), parc.Context)
		ca, _ = a.(*attributes)
	}

	return cedar.Request{
		Principal: cedar.NewEntityUID(cedar.EntityType(parc.Principal.Type()), cedar.String(parc.Principal.ID())),
		Action:    cedar.NewEntityUID(cedar.EntityType(parc.Action.Type()), cedar.String(parc.Action.ID())),
		Resource:  cedar.NewEntityUID(cedar.EntityType(parc.Resource.Type()), cedar.String(parc.Resource.ID())),
		Context:   cedar.NewRecord(ca.cedarSet),
	}
}
