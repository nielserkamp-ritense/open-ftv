package opa

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/open-policy-agent/opa/sdk"
	"go.opentelemetry.io/otel/trace"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/shared/types"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/standards"
)

// Authorize implements the Controller interface.
func (c *controller) Authorize(req *types.Request) (*types.Response, error) {

	var span trace.Span

	if ldv := c.LDV(); ldv != nil {
		_, span = ldv.StartSpan(context.Background())
	}

	debug := c.Logger().Enabled(nil, slog.LevelDebug)
	if debug {
		c.Logger().Debug("authorization request", "controller", c.String(), "request-uid", req.UID)
	}

	started := time.Now()
	resp, err := c.pdp.Decision(context.Background(), c.buildDecisionOptions(req))
	duration := time.Since(started)

	if span != nil {
		span.End()
	}

	if err != nil {
		c.Logger().Error("authorization failed", "controller", c.String(), "request-uid", req.UID, "err", err, "pdp elapsed", duration.String())
	} else {
		if m, ok := resp.Result.(map[string]any); ok {
			if allowed, ok2 := m["allow"].(bool); ok2 && allowed {
				if debug {
					c.Logger().Debug("authorization granted", "controller", c.String(), "request-uid", req.UID, "pdp elapsed", duration.String())
				}
				return &types.Response{Allowed: true}, nil
			}
		}

		if debug {
			c.Logger().Debug("authorization not granted", "controller", c.String(), "request-uid", req.UID, "pdp elapsed", duration.String())
		}
	}

	return &types.Response{Allowed: false, Message: "not authorized"}, nil
}

func (c *controller) buildDecisionOptions(req *types.Request) sdk.DecisionOptions {
	a, newURI := c.PIP().CollectAttributesFromRequest(req)
	m := types.MapFromAttributes(a)

	if newURI != "" {
		m["uri"] = newURI
	}

	p1, p2 := standards.DeterminePrincipal(a)

	return sdk.DecisionOptions{
		Now:        *req.RequestTime,
		Path:       fmt.Sprintf("/%s/%s", p1, p2),
		Input:      m,
		DecisionID: req.UID.String(),
	}
}
