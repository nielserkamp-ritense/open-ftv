package cedar_embedded

import (
	"log/slog"
	"time"

	"github.com/cedar-policy/cedar-go"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

// Authorize implements the Controller interface.
func (c *controller) Authorize(uid string, req *models.PARC) (*models.Response, error) {
	var logger *slog.Logger
	if c.Logger.Enabled(nil, slog.LevelDebug) {
		logger = c.Logger.With("controller", c.String(), "request-uid", uid)
		logger.Debug("authorization request", "controller", c.String(), "request-uid", uid)
	}

	req2 := c.buildCedarRequest(req)

	c.AuthMutex.RLock()

	started := time.Now()
	decision, diagnostic := cedar.Authorize(c.pdp, c.entities, req2)
	duration := time.Since(started)

	c.AuthMutex.RUnlock()

	if decision {
		if logger != nil {
			logger.Debug("authorization granted", "pdp elapsed", duration.String())
		}
		return &models.Response{Allowed: true}, nil
	}

	if logger != nil {
		logger.Warn("authorization not granted", "diagnostic", diagnostic, "pdp elapsed", duration.String())
	}

	return &models.Response{Allowed: false, Message: "not authorized", Attributes: map[string]any{"diagnostic": diagnostic}}, nil
}

// Batch implements the Controller interface.
func (c *controller) Batch(uid string, req *models.Batch) ([]models.Response, error) {
	var logger *slog.Logger
	if c.Logger.Enabled(nil, slog.LevelDebug) {
		logger = c.Logger.With("controller", c.String(), "request-uid", uid)
		logger.Debug("batch request", "semantics", req.Semantics.String())
	}

	out := make([]models.Response, 0, len(req.Items))

	c.AuthMutex.RLock()

	for i := range req.Items {
		r1 := &req.Items[i]
		r2 := c.buildCedarRequest(r1)

		started := time.Now()
		decision, diagnostic := cedar.Authorize(c.pdp, c.entities, r2)
		duration := time.Since(started)

		if decision {
			if logger != nil {
				logger.Debug("authorization granted", "item#", i+1, "pdp elapsed", duration.String())
			}
			out = append(out, models.Response{Allowed: true})

			if req.Semantics == models.PermitOnFirstPermit {
				break
			}
		} else {
			if logger != nil {
				logger.Warn("authorization not granted", "item#", i+1, "diagnostic", diagnostic, "pdp elapsed", duration.String())
			}
			out = append(out, models.Response{Allowed: false, Message: "not authorized", Attributes: map[string]any{"diagnostic": diagnostic}})

			if req.Semantics == models.DenyOnFirstDeny {
				break
			}
		}
	}

	c.AuthMutex.RUnlock()

	return out, nil
}

func (c *controller) buildCedarRequest(parc *models.PARC) cedar.Request {
	parc = c.Map(parc)

	ca := make(cedar.RecordMap)
	parc.Context.IterateAttributes(func(attr *models.Attribute) {
		v, err := anyToValue(attr.Value())
		if err != nil {
			c.Logger.Warn("failed to convert context attribute to cedar format", "key", attr.Key(), "type", attr.Type, "value", attr.Value(), "err", err)
		}
		ca[cedar.String(attr.Key())] = v
	})

	return cedar.Request{
		Principal: cedar.NewEntityUID(cedar.EntityType(parc.Principal.Type()), cedar.String(parc.Principal.ID())),
		Action:    cedar.NewEntityUID(cedar.EntityType(parc.Action.Type()), cedar.String(parc.Action.ID())),
		Resource:  cedar.NewEntityUID(cedar.EntityType(parc.Resource.Type()), cedar.String(parc.Resource.ID())),
		Context:   cedar.NewRecord(ca),
	}
}
