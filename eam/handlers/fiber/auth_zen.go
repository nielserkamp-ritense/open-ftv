package fiber

import (
	"errors"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/authlog"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller"
	server "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/authzen"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/convert"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/xsd"
)

// AuthZENVersion is the full semantic API version for the AuthZEN endpoints.
const AuthZENVersion = "1.0.0"

// AuthZENAuthorizer represents the interface for handling AuthZEN authorization requests.
type AuthZENAuthorizer interface {
	Authorize(req *fiber.Ctx) error
}

// NewAuthHandlerZEN instantiates a new AuthZEN authorization handler.
func NewAuthHandlerZEN(logger *slog.Logger, authLogger authlog.Logger, controller pdp.Controller) AuthZENAuthorizer {
	return &authZEN{logger: logger, authLogger: authLogger, controller: controller}
}

// Authorize implements the AuthZENAuthorizer interface.
func (h *authZEN) Authorize(fc *fiber.Ctx) error {
	p := &authProcess{
		logger:     h.logger,
		authLogger: h.authLogger,
		controller: h.controller,
		fc:         fc,
		status:     fiber.StatusInternalServerError,
		started:    time.Now(),
	}

	if p.logger.Enabled(nil, slog.LevelInfo) {
		defer p.log()
	}
	if p.authLogger != nil {
		defer p.authLog()
	}

	p.fc.Set(HeaderVersion, AuthZENVersion)

	req := p.verifyRequestAuthZEN()
	if p.err != nil {
		p.logger.Error("AuthZEN authorization handler failed", "error", p.err)
		return server.SendMessageResponse(fc, p.status, p.msg)
	}

	p.newAuthRequestAuthZEN(req, fc.GetReqHeaders())
	return p.authorizeAuthZEN()
}

func (p *authProcess) verifyRequestAuthZEN() *authzen.AuthorizationRequest {
	p.status = fiber.StatusBadRequest

	if req := p.fc.Request(); len(req.Header.ContentType()) == 0 {
		req.Header.SetContentType(fiber.MIMEApplicationJSON)
	}

	req := &authzen.AuthorizationRequest{}
	if p.err = p.fc.BodyParser(req); p.err != nil {
		p.msg = "invalid input data"
		return nil
	}

	if req.Subject.Type == "" || req.Subject.Id == "" {
		p.msg, p.err = "invalid subject", errors.New("subject type&id must be filled")
		return nil
	}

	if req.Action.Name == "" {
		p.msg, p.err = "invalid action", errors.New("action name must be filled")
		return nil
	}

	if req.Resource.Type == "" || req.Resource.Id == "" {
		p.msg, p.err = "invalid resource", errors.New("resource type&id must be filled")
		return nil
	}

	p.status = fiber.StatusOK
	return req
}

func (p *authProcess) newAuthRequestAuthZEN(req *authzen.AuthorizationRequest, headers map[string][]string) {
	principal := models.NewEntity(req.Subject.Type, req.Subject.Id, models.NewAttributeSet(req.Subject.Properties))
	action := models.NewEntity(models.EntityTypeName, req.Action.Name, models.NewAttributeSet(req.Action.Properties))
	resource := models.NewEntity(req.Resource.Type, req.Resource.Id, models.NewAttributeSet(req.Resource.Properties))
	ctx := models.NewAttributeSet(req.Context)

	if t := convert.AnyToDateTime(ctx.GetAttributeValue(models.AttrTime)); t.IsZero() {
		ctx.AddAttributeWithType(models.AttrTime, time.Now().UTC(), xsd.PrefixDateTime)
	}

	p.reqUID = uuid.New().String()
	p.parc = &models.PARC{
		Principal: principal,
		Action:    action,
		Resource:  resource,
		Context:   ctx,
	}
}

func (p *authProcess) authorizeAuthZEN() error {
	if reqID := p.fc.Get("X-Request-ID"); reqID != "" {
		p.fc.Set("X-Request-ID", reqID)
	}

	if p.resp, p.err = p.controller.Authorize(p.reqUID, p.parc); p.err != nil {
		p.msg = "AuthZEN authorization process failed"
		return server.SendMessageResponse(p.fc, p.status, p.msg)
	}

	allowed, msg := p.resp.Allowed, p.resp.Message
	if msg == "" {
		if allowed {
			msg = "ok"
		} else {
			msg = "not authorized"
		}
	}

	return p.fc.JSON(&authZENResponse{
		Decision: allowed,
		Context:  authZENContext(p.resp, msg),
	})
}

// authZENResponse mirrors authzen.AuthorizationResponse but carries the extended
// decision context below. It keeps the AuthZEN wire shape ("context" before
// "decision") so existing clients (and the exact-match tests) are unaffected.
type authZENResponse struct {
	Context  *authZENDecisionContext `json:"context,omitempty"`
	Decision bool                    `json:"decision"`
}

// authZENDecisionContext extends the NLGov AuthZEN ReasonObject with two fields
// the BRO PEP's and the mock-PDP conventions read (findings A9/B19/B20):
//
//   - reason: a flat human-readable string (the mock-PDP and the BRO PEP's
//     ResponseContext.DenyReason read context.reason; they cannot read the
//     AuthZEN reasonUser/reasonAdmin maps, which are typed as strings there);
//   - audit_identifiers.policy_version: the governing policy uid/hash, which the
//     PEP's ResponseContext.PolicyRef reads as the policy reference.
//
// The map-shaped reasonUser/reasonAdmin and the obligations array are kept for
// AuthZEN conformance; the new fields are additive and backward compatible.
type authZENDecisionContext struct {
	Id               string              `json:"id"`
	Reason           string              `json:"reason,omitempty"`
	ReasonAdmin      authzen.ReasonField `json:"reasonAdmin,omitempty"`
	ReasonUser       authzen.ReasonField `json:"reasonUser,omitempty"`
	Obligations      []map[string]any    `json:"obligations,omitempty"`
	AuditIdentifiers map[string]any      `json:"audit_identifiers,omitempty"`
}

// authZENContext builds the AuthZEN Decision context from the PDP response: the
// reason (reasonUser + flat reason), the governing policy uid (id + reasonAdmin
// + audit_identifiers.policy_version) and any ODRL obligations/duties that the
// PEP must fulfil.
func authZENContext(resp *models.Response, msg string) *authZENDecisionContext {
	ctx := &authZENDecisionContext{Id: "0", ReasonUser: authzen.ReasonField{"en": msg}}
	if resp == nil {
		return ctx
	}

	if resp.PolicyKey != "" {
		ctx.Id = resp.PolicyKey
		ctx.ReasonAdmin = authzen.ReasonField{"policy": resp.PolicyKey}
		// audit_identifiers.policy_version: the policy reference the mock-PDP and
		// the BRO PEP's PolicyRef read (B20).
		ctx.AuditIdentifiers = map[string]any{"policy_version": resp.PolicyKey}
	}
	if reason, ok := resp.Attributes["decision_reason"].(string); ok && reason != "" {
		if ctx.ReasonAdmin == nil {
			ctx.ReasonAdmin = authzen.ReasonField{}
		}
		ctx.ReasonAdmin["reason"] = reason
		// flat context.reason for the BRO PEP / mock-PDP (A9/B19).
		ctx.Reason = reason
	}
	if obs := obligationsFromAttributes(resp.Attributes["obligations"]); len(obs) > 0 {
		ctx.Obligations = obs
	}
	return ctx
}

// obligationsFromAttributes normalises the obligations carried in
// models.Response.Attributes (produced by the ODRL/ODRL-geo engine) to the
// AuthZEN obligation array shape.
func obligationsFromAttributes(v any) []map[string]any {
	switch o := v.(type) {
	case []map[string]any:
		return o
	case []any:
		out := make([]map[string]any, 0, len(o))
		for _, e := range o {
			if m, ok := e.(map[string]any); ok {
				out = append(out, m)
			}
		}
		return out
	}
	return nil
}

type authZEN struct {
	logger     *slog.Logger
	authLogger authlog.Logger
	controller pdp.Controller
}
