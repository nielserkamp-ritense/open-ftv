package opa_embedded

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/open-policy-agent/opa/sdk"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pep"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/convert"
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
	parc = c.Map(parc)

	t := convert.AnyToDateTime(parc.Context.GetAttributeValue(models.AttrTime))
	if t.IsZero() {
		t = time.Now().UTC()
	}

	data := map[string]any{
		"principal": models.EntityToAttribute(parc.Principal).Value(),
		"action":    models.EntityToAttribute(parc.Action).Value(),
		"resource":  models.EntityToAttribute(parc.Resource).Value(),
		"context":   models.MapFromAttributes(parc.Context),
	}

	return sdk.DecisionOptions{
		DecisionID: uid,
		Now:        t,
		Path:       c.determinePath(parc),
		Input:      data,
	}
}

func (c *controller) determinePath(parc *models.PARC) string {
	rvvaID, doelbinding, ok := identifiersFromPrincipal(parc.Principal)
	if !ok {
		rvvaID, doelbinding, ok = identifiersFromContext(parc.Context)
		if !ok {
			rvvaID, doelbinding = identifiersFromHeaders(parc.Context.GetAttributeValue("headers"))
		}
	}

	switch {
	case rvvaID != "":
		return fmt.Sprintf("/activity/%s", rvvaID)
	case doelbinding != "":
		return fmt.Sprintf("/doelbinding/%s", doelbinding)
	default:
		return "/authz"
	}
}

func identifiersFromPrincipal(principal models.Entity) (string, string, bool) {
	switch principal.Type() {
	case pep.PrincipalRVVA:
		return principal.ID(), "", true
	case pep.PrincipalDoelbinding:
		return "", principal.ID(), true
	default:
		return "", "", false
	}
}

func identifiersFromContext(context models.AttributeSet) (string, string, bool) {
	rvvaID := convert.AnyToString(context.GetAttributeValue(models.AttrRvvaID))
	doelbinding := convert.AnyToString(context.GetAttributeValue(models.AttrDoelbinding))
	return rvvaID, doelbinding, rvvaID != "" || doelbinding != ""
}

func identifiersFromHeaders(headers any) (string, string) {
	if m, ok := headers.(map[string]any); ok {
		for k := range m {
			switch strings.ToLower(k) {
			case models.HeaderRvvaID, models.HeaderObsoleteRvvaID:
				return firstWord(convert.AnyToString(m[k])), ""
			case models.HeaderDoelbinding:
				return "", firstWord(convert.AnyToString(m[k]))
			}
		}
	}

	if m, ok := headers.(map[string]string); ok {
		for k := range m {
			switch strings.ToLower(k) {
			case models.HeaderRvvaID, models.HeaderObsoleteRvvaID:
				return firstWord(m[k]), ""
			case models.HeaderDoelbinding:
				return "", firstWord(m[k])
			}
		}
	}

	return "", ""
}

func firstWord(in string) string {
	if i := strings.Index(in, ","); i > 0 {
		return in[:i]
	}
	return in
}
