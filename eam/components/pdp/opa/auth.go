package opa

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/open-policy-agent/opa/sdk"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pep"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
)

// Authorize implements the Controller interface.
func (c *controller) Authorize(req *models.Request) (resp *models.Response, err error) {
	debug := c.Logger().Enabled(nil, slog.LevelDebug)
	if debug {
		c.Logger().Debug("authorization request", "controller", c.String(), "request-uid", req.UID)
	}

	var decision *sdk.DecisionResult
	started := time.Now()
	decision, err = c.pdp.Decision(context.Background(), c.buildDecisionOptions(req))
	duration := time.Since(started)

	if err != nil {
		c.Logger().Error("authorization failed", "controller", c.String(), "request-uid", req.UID, "err", err, "pdp elapsed", duration.String())
	} else {
		if m, ok := decision.Result.(map[string]any); ok {
			if allowed, ok2 := m["allow"].(bool); ok2 && allowed {
				if debug {
					c.Logger().Debug("authorization granted", "controller", c.String(), "request-uid", req.UID, "pdp elapsed", duration.String())
				}
				resp = &models.Response{Allowed: true}
				return
			}
		}

		if debug {
			c.Logger().Warn("authorization not granted", "controller", c.String(), "request-uid", req.UID, "pdp elapsed", duration.String())
		}
	}

	resp = &models.Response{Allowed: false, Message: "not authorized"}
	return
}

func (c *controller) buildDecisionOptions(req *models.Request) sdk.DecisionOptions {
	a, newURI := c.PIP().CollectAttributesFromRequest(req)
	m := models.MapFromAttributes(a)

	if newURI != "" {
		m["uri"] = newURI
	}

	p1, p2 := pep.DeterminePrincipal(a)

	return sdk.DecisionOptions{
		Now:        *req.RequestTime,
		Path:       fmt.Sprintf("/%s/%s", p1, p2),
		Input:      m,
		DecisionID: req.UID.String(),
	}
}
