package opa

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/open-policy-agent/opa/sdk"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/shared/types"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/standards"
)

// Authorize implements the Controller interface.
func (c *controller) Authorize(req *types.Request) (resp *types.Response, err error) {
	finish := c.startLog(req)
	defer func() { finish(resp) }()

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
				resp = &types.Response{Allowed: true}
				return
			}
		}

		if debug {
			c.Logger().Debug("authorization not granted", "controller", c.String(), "request-uid", req.UID, "pdp elapsed", duration.String())
		}
	}

	resp = &types.Response{Allowed: false, Message: "not authorized"}
	return
}

func (c *controller) startLog(req *types.Request) func(resp *types.Response) {
	ldv := c.LDV()
	if ldv == nil {
		return func(*types.Response) {}
	}

	_, span := ldv.StartSpan(
		context.Background(),
		attribute.String("authz.policy.engine", c.String()),
	)

	return func(result *types.Response) {
		c.endLog(span, result)
	}
}

func (c *controller) endLog(span trace.Span, result *types.Response) {
	span.SetAttributes(
		attribute.Bool("authz.policy.allowed", result.Allowed),
		attribute.String("authz.policy.message", result.Message),
		attribute.String("authz.policy.key", result.PolicyKey),
		attribute.String("authz.policy.hash", result.PolicyHash),
	)

	span.End()
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
