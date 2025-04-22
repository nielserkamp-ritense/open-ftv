package fiber

import (
	"errors"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/log/authlog"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/pdp/controller"
	server "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/server/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/oas/authzen"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/convert"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/xsd"
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

	return p.fc.JSON(&authzen.AuthorizationResponse{
		Decision: allowed,
		Context:  &authzen.ReasonObject{Id: "0", ReasonUser: authzen.ReasonField{"en": msg}},
	})
}

type authZEN struct {
	logger     *slog.Logger
	authLogger authlog.Logger
	controller pdp.Controller
}
