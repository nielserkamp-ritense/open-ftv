package cedar

import (
	"log/slog"
	"time"

	"github.com/cedar-policy/cedar-go"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/shared/control"
)

// Authorize implements the Controller interface.
func (c *controller) Authorize(req *control.Request) (*control.Response, error) {
	debug := c.Logger().Enabled(nil, slog.LevelDebug)
	if debug {
		c.Logger().Debug("authorization request", "controller", c.String(), "request-uid", req.UID)
	}

	started := time.Now()
	decision, diagnostic := c.pdp.IsAuthorized(c.entities, c.buildCedarRequest(req))
	duration := time.Since(started)

	if decision {
		if debug {
			c.Logger().Debug("authorization granted", "controller", c.String(), "request-uid", req.UID, "pdp elapsed", duration.String())
		}
		return &control.Response{Allowed: true}, nil
	}

	if debug {
		c.Logger().Error("authorization failed", "controller", c.String(), "request-uid", req.UID, "diagnostic", diagnostic, "pdp elapsed", duration.String())
	}

	return &control.Response{Allowed: false, Message: "not authorized"}, nil
}

func (c *controller) buildCedarRequest(req *control.Request) cedar.Request {
	a, uri := c.PIP().CollectAttributesFromRequest(req)
	if uri == "" && req.URL != nil {
		uri = req.URL.String()
	}

	ca, ok := a.(*attributes)
	if !ok {
		a = NewAttributeSet(c.Logger(), a)
		ca, _ = a.(*attributes)
	}

	p1, p2 := DeterminePrincipal(ca)

	return cedar.Request{
		Principal: cedar.NewEntityUID(p1, p2),
		Action:    cedar.NewEntityUID(TypeAction, cedar.String(req.Method)),
		Resource:  cedar.NewEntityUID(TypeService, cedar.String(uri)),
		Context:   cedar.NewRecord(ca.set),
	}
}
