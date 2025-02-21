package opa

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/open-policy-agent/opa/sdk"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
)

// Authorize implements the Controller interface.
func (c *controller) Authorize(req *models.Request) (resp *models.Response, err error) {
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
				resp = &models.Response{Allowed: true}
				return
			}
		}

		if debug {
			c.Logger().Debug("authorization not granted", "controller", c.String(), "request-uid", req.UID, "pdp elapsed", duration.String())
		}
	}

	resp = &models.Response{Allowed: false, Message: "not authorized"}
	return
}

func (c *controller) startLog(_ *models.Request) func(resp *models.Response) {
	ldv := c.Logboek()
	if ldv == nil {
		return func(*models.Response) {}
	}

	_, span := ldv.StartSpan(
		context.Background(),
		attribute.String("authz.policy.engine", c.String()),
	)

	return func(result *models.Response) {
		c.endLog(span, result)
	}
}

func (c *controller) endLog(span trace.Span, result *models.Response) {
	span.SetAttributes(
		attribute.Bool("authz.policy.allowed", result.Allowed),
		attribute.String("authz.policy.message", result.Message),
		attribute.String("authz.policy.key", result.PolicyKey),
		attribute.String("authz.policy.hash", result.PolicyHash),
	)

	span.End()
}

func (c *controller) buildDecisionOptions(req *models.Request) sdk.DecisionOptions {
	a, newURI := c.PIP().CollectAttributesFromRequest(req)
	m := models.MapFromAttributes(a)

	if newURI != "" {
		m["uri"] = newURI
	}

	p1, p2 := models.DeterminePrincipal(a)

	return sdk.DecisionOptions{
		Now:        *req.RequestTime,
		Path:       fmt.Sprintf("/%s/%s", p1, p2),
		Input:      m,
		DecisionID: req.UID.String(),
	}
}
