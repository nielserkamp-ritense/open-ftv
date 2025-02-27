package opa

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/open-policy-agent/opa/sdk"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/convert"
)

// Authorize implements the Controller interface.
func (c *controller) Authorize(uid string, parc *models.PARC) (resp *models.Response, err error) {
	debug := c.Logger().Enabled(nil, slog.LevelDebug)
	if debug {
		c.Logger().Debug("authorization request", "controller", c.String(), "request-uid", uid)
	}

	opts := c.buildDecisionOptions(uid, parc)

	var decision *sdk.DecisionResult
	started := time.Now()
	decision, err = c.pdp.Decision(context.Background(), opts)
	duration := time.Since(started)

	if err != nil {
		c.Logger().Error("authorization failed", "controller", c.String(), "request-uid", uid, "err", err, "pdp elapsed", duration.String())
	} else {
		if m, ok := decision.Result.(map[string]any); ok {
			if allowed, ok2 := m["allow"].(bool); ok2 && allowed {
				if debug {
					c.Logger().Debug("authorization granted", "controller", c.String(), "request-uid", uid, "pdp elapsed", duration.String())
				}
				resp = &models.Response{Allowed: true}
				return
			}
		}

		if debug {
			c.Logger().Warn("authorization not granted", "controller", c.String(), "request-uid", uid, "pdp elapsed", duration.String())
		}
	}

	resp = &models.Response{Allowed: false, Message: "not authorized"}
	return
}

func (c *controller) buildDecisionOptions(uid string, parc *models.PARC) sdk.DecisionOptions {
	t := convert.AnyToDateTime(parc.Context.GetAttributeValue(models.AttrTime))
	if t.IsZero() {
		t = time.Now().UTC()
	}

	m := models.MapFromAttributes(parc.Context)

	// TODO: pass the action and resource in the context.

	return sdk.DecisionOptions{
		DecisionID: uid,
		Now:        t,
		Path:       fmt.Sprintf("/%s/%s", parc.Principal.Type(), parc.Principal.ID()),
		Input:      m,
	}
}
