package cedar

import (
	"log/slog"
	"time"

	"github.com/cedar-policy/cedar-go"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
)

// Authorize implements the Controller interface.
func (c *controller) Authorize(req *models.Request) (*models.Response, error) {
	debug := c.Logger().Enabled(nil, slog.LevelDebug)
	if debug {
		c.Logger().Debug("authorization request", "controller", c.String(), "request-uid", req.UID)
	}

	req2 := c.buildCedarRequest(req)

	started := time.Now()
	decision, diagnostic := c.pdp.IsAuthorized(c.entities, req2)
	duration := time.Since(started)

	if decision {
		if debug {
			c.Logger().Debug("authorization granted", "controller", c.String(), "request-uid", req.UID, "pdp elapsed", duration.String())
		}
		return &models.Response{Allowed: true}, nil
	}

	if debug {
		c.Logger().Error("authorization failed", "controller", c.String(), "request-uid", req.UID, "diagnostic", diagnostic, "pdp elapsed", duration.String())
	}

	return &models.Response{Allowed: false, Message: "not authorized", Attributes: map[string]any{"diagnostic": diagnostic}}, nil
}

func (c *controller) buildCedarRequest(req *models.Request) cedar.Request {
	a, uri := c.PIP().CollectAttributesFromRequest(req)
	if uri == "" && req.URL != nil {
		uri = req.URL.String()
	}

	ca, ok := a.(*attributes)
	if !ok {
		a = NewAttributeSet(c.Logger(), a)
		ca, _ = a.(*attributes)
	}

	p1, p2 := models.DeterminePrincipal(a)

	req.Principal = models.NewEntity(p1, p2, nil)
	req.Action = models.NewEntity(TypeAction, req.Method, nil)
	req.Resource = models.NewEntity(TypeService, uri, nil)

	return cedar.Request{
		Principal: cedar.NewEntityUID(cedar.EntityType(req.Principal.Type()), cedar.String(req.Principal.ID())),
		Action:    cedar.NewEntityUID(cedar.EntityType(req.Action.Type()), cedar.String(req.Action.ID())),
		Resource:  cedar.NewEntityUID(cedar.EntityType(req.Resource.Type()), cedar.String(req.Resource.ID())),
		Context:   cedar.NewRecord(ca.cedarSet),
	}
}
