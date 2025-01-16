package fiber

import (
	"errors"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/oas/authzen"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/components"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/components/pdp"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/models"
)

// AuthZENAuthorizer represents the interface for handling AuthZEN authorization requests.
type AuthZENAuthorizer interface {
	Authorize(req *fiber.Ctx) error
}

// NewAuthHandlerZEN instantiates a new AuthZEN authorization handler.
func NewAuthHandlerZEN(logger *slog.Logger, controller pdp.Controller) AuthZENAuthorizer {
	return &authZEN{logger: logger, controller: controller}
}

// Authorize implements the AuthZENAuthorizer interface.
func (h *authZEN) Authorize(fc *fiber.Ctx) error {
	p := &authProcess{
		logger:     h.logger,
		controller: h.controller,
		fc:         fc,
		status:     fiber.StatusInternalServerError,
		started:    time.Now(),
	}

	if p.logger.Enabled(nil, slog.LevelInfo) {
		defer p.log()
	}

	req := p.verifyRequestAuthZEN()
	if p.err != nil {
		p.logger.Error("AuthZEN authorization handler failed", "error", p.err)
		return SendMessageResponse(fc, p.status, p.msg)
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
	actionAttrs := models.NewAttributeSet(req.Action.Properties)
	method, _ := actionAttrs.GetAttributeValue(models.AttrMethod).(string)

	principal := models.NewEntity(req.Subject.Type, req.Subject.Id, models.NewAttributeSet(req.Subject.Properties))
	action := models.NewEntity(models.EntityAction, req.Action.Name, actionAttrs)
	resource := models.NewEntity(req.Resource.Type, req.Resource.Id, models.NewAttributeSet(req.Resource.Properties))

	var attr map[string]any
	if req.Context != nil {
		attr = req.Context
	} else {
		attr = make(map[string]any)
	}

	uid, now := uuid.New(), time.Now().UTC()
	p.req = &components.Request{
		UID:         &uid,
		RequestTime: &now,
		Method:      method,
		Headers:     headers,
		Principal:   principal,
		Action:      action,
		Resource:    resource,
		Attributes:  attr,
	}
}

func (p *authProcess) authorizeAuthZEN() error {
	if p.resp, p.err = p.controller.Authorize(p.req); p.err != nil {
		p.msg = "AuthZEN authorization process failed"
		return SendMessageResponse(p.fc, p.status, p.msg)
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
		Context:  map[string]any{"en": msg},
	})
}

type authZEN struct {
	logger     *slog.Logger
	controller pdp.Controller
}
