package cerbos

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/cerbos/cerbos-sdk-go/cerbos"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
)

// Authorize implements the Controller interface.
func (c *controller) Authorize(uid string, req *models.PARC) (*models.Response, error) {
	debug := c.Logger().Enabled(nil, slog.LevelDebug)
	if debug {
		c.logger.Debug("authorization request", "request-uid", uid)
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
				c.logger.Debug("authorization granted", "request-uid", uid, "pdp elapsed", duration.String())
			}
			return &models.Response{Allowed: true}, nil
		}

		c.logger.Debug("authorization not granted", "request-uid", uid, "pdp elapsed", duration.String(), "diagnostic", match)
	}

	if err != nil {
		c.logger.Warn("authorization failed", "request-uid", uid, "pdp elapsed", duration.String(), "error", err)

	}
	return &models.Response{Allowed: false, Message: "not authorized"}, nil
}

func (c *controller) buildCerbosRequest(parc *models.PARC) (*cerbos.Principal, *cerbos.Resource, string) {
	principal := cerbos.NewPrincipal(fmt.Sprintf("%s:%s", parc.Principal.Type(), parc.Principal.ID()), "doelbinding").WithAttributes(models.MapFromAttributes(parc.Principal.Attributes()))
	resource := cerbos.NewResource(parc.Resource.Type(), parc.Resource.ID()).WithAttributes(models.MapFromAttributes(parc.Resource.Attributes()))
	action := parc.Action.ID()

	return principal, resource, action
}
