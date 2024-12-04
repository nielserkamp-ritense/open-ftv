package handlers

import (
	"errors"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/oas/authzen"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/standards"
)

// AuthZEN implements the authorization handler for AuthZEN requests.
func (h *authHandler) AuthZEN(fc *fiber.Ctx) error {
	p := &authProcess{
		authHandler: *h,
		fc:          fc,
		status:      fiber.StatusInternalServerError,
		started:     time.Now(),
	}

	if p.logger.Enabled(nil, slog.LevelInfo) {
		defer p.log()
	}

	req := p.verifyRequestAuthZEN()
	if p.err != nil {
		p.logger.Error("authorization handler failed", "error", p.err)
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

	if req.Action.Name == nil || *req.Action.Name == "" {
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
	actionAttrs := standards.NewAttributeSet(req.Action.Properties)
	method, _ := actionAttrs.GetAttribute(standards.AttrMethod).(string)

	principal := standards.NewEntity(req.Subject.Type, req.Subject.Id, standards.NewAttributeSet(req.Subject.Properties))
	action := standards.NewEntity(standards.EntityAction, *req.Action.Name, actionAttrs)
	resource := standards.NewEntity(req.Resource.Type, req.Resource.Id, standards.NewAttributeSet(req.Resource.Properties))

	var attr map[string]any
	if req.Context != nil {
		attr = *req.Context
	} else {
		attr = make(map[string]any)
	}

	uid, now := uuid.New(), time.Now().UTC()
	p.req = &standards.Request{
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
		p.msg = "authorization process failed"
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
		Context:  &authzen.ReasonField{"en": msg},
	})
}
