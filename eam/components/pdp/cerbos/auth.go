package cerbos

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/cerbos/cerbos-sdk-go/cerbos"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pep"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
)

// Authorize implements the Controller interface.
func (c *controller) Authorize(req *models.Request) (*models.Response, error) {
	debug := c.Logger().Enabled(nil, slog.LevelDebug)
	if debug {
		c.logger.Debug("authorization request", "request-uid", req.UID)
	}

	principal, resource, action := c.buildCerbosRequest(req)
	batch := cerbos.NewResourceBatch().Add(resource, action)

	started := time.Now()
	decision, err := c.engine.CheckResources(c.Context(), principal, batch)
	duration := time.Since(started)

	if err == nil {
		match := decision.GetResource(resource.ID())
		if match.IsAllowed(action) {
			if debug {
				c.logger.Debug("authorization granted", "request-uid", req.UID, "pdp elapsed", duration.String())
			}
			return &models.Response{Allowed: true}, nil
		}

		c.logger.Debug("authorization not granted", "request-uid", req.UID, "pdp elapsed", duration.String(), "diagnostic", match)
	}

	if err != nil {
		c.logger.Warn("authorization failed", "request-uid", req.UID, "pdp elapsed", duration.String(), "error", err)

	}
	return &models.Response{Allowed: false, Message: "not authorized"}, nil
}

func (c *controller) buildCerbosRequest(req *models.Request) (*cerbos.Principal, *cerbos.Resource, string) {
	a, uri := c.PIP().CollectAttributesFromRequest(req)
	if uri == "" && req.URL != nil {
		uri = req.URL.String()
	}

	p1, p2 := pep.DeterminePrincipal(a)

	principal := cerbos.NewPrincipal(fmt.Sprintf("%s:%s", p1, p2), "doelbinding").WithAttributes(models.MapFromAttributes(a))
	resource := cerbos.NewResource(uri, uri)
	action := req.Method

	return principal, resource, action
}
