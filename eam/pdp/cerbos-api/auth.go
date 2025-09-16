package cerbos_api

import (
	"log/slog"
	"time"

	"github.com/cerbos/cerbos-sdk-go/cerbos"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

// Authorize implements the Controller interface.
func (c *controller) Authorize(uid string, req *models.PARC) (*models.Response, error) {
	logger := c.logger.With("request-uid", uid)

	debug := c.logger.Enabled(nil, slog.LevelDebug)
	if debug {
		logger.Debug("authorization request")
	}

	principal, resource, action := c.buildCerbosRequest(req)
	batch := cerbos.NewResourceBatch().Add(resource, action)

	c.AuthMutex.RLock()

	started := time.Now()
	decision, err := c.engine.CheckResources(c.Ctx, principal, batch)
	duration := time.Since(started)

	c.AuthMutex.RUnlock()

	if err == nil {
		match := decision.GetResource(resource.ID())
		if match.IsAllowed(action) {
			if debug {
				logger.Debug("authorization granted", "pdp elapsed", duration.String())
			}
			return &models.Response{Allowed: true}, nil
		}

		if debug {
			logger.Debug("authorization not granted", "pdp elapsed", duration.String(), "diagnostic", match)
		}
		return &models.Response{Allowed: false, Message: "not authorized"}, nil
	}

	logger.Warn("authorization failed", "pdp elapsed", duration.String(), "error", err)
	return &models.Response{Allowed: false, Message: "not authorized"}, nil
}

// Batch implements the Controller interface.
func (c *controller) Batch(uid string, req *models.Batch) ([]models.Response, error) {
	logger := c.logger.With("request-uid", uid)

	debug := c.logger.Enabled(nil, slog.LevelDebug)
	if debug {
		logger.Debug("batch request", "semantics", req.Semantics.String())
	}

	out := make([]models.Response, 0, len(req.Items))

	c.AuthMutex.RLock()

	for i := range req.Items {
		r := &req.Items[i]

		principal, resource, action := c.buildCerbosRequest(r)
		batch := cerbos.NewResourceBatch().Add(resource, action)

		started := time.Now()
		decision, err := c.engine.CheckResources(c.Ctx, principal, batch)
		duration := time.Since(started)

		if err == nil {
			match := decision.GetResource(resource.ID())
			if match.IsAllowed(action) {
				if debug {
					logger.Debug("authorization granted", "item#", i+1, "pdp elapsed", duration.String())
				}
				out = append(out, models.Response{Allowed: true})

				if req.Semantics == models.PermitOnFirstPermit {
					break
				} else {
					continue
				}
			}

			if debug {
				logger.Debug("authorization not granted", "item#", i+1, "pdp elapsed", duration.String(), "diagnostic", match)
			}
			out = append(out, models.Response{Allowed: false, Message: "not authorized"})
		} else {
			logger.Warn("authorization failed", "item#", i+1, "pdp elapsed", duration.String(), "error", err)
			out = append(out, models.Response{Allowed: false, Message: "not authorized"})
		}

		if req.Semantics == models.DenyOnFirstDeny {
			break
		}
	}

	c.AuthMutex.RUnlock()

	return out, nil
}

func (c *controller) buildCerbosRequest(parc *models.PARC) (*cerbos.Principal, *cerbos.Resource, string) {
	parc = c.Map(parc)

	// principal := cerbos.NewPrincipal(fmt.Sprintf("%s:%s", parc.Principal.Type(), parc.Principal.ID()), "doelbinding").WithAttributes(schema.MapFromAttributes(parc.Principal.Attributes()))
	// resource := cerbos.NewResource(parc.Resource.Type(), fmt.Sprintf("%s:%s", parc.Resource.Type(), parc.Resource.ID())).WithAttributes(schema.MapFromAttributes(parc.Resource.Attributes()))
	principal := cerbos.NewPrincipal(parc.Principal.ID(), "doelbinding").WithAttributes(models.MapFromAttributes(parc.Principal.Attributes()))
	resource := cerbos.NewResource(parc.Resource.Type(), parc.Resource.ID()).WithAttributes(models.MapFromAttributes(parc.Resource.Attributes()))
	action := parc.Action.ID()

	// TODO: context???

	return principal, resource, action
}
