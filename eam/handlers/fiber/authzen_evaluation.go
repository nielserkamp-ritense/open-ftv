package fiber

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	server "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server/fiber"
	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/authzen"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/convert"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/xsd"
)

func (p *authProcess) verifyRequestAuthZEN() *oas.EvaluationRequest {
	p.status = fiber.StatusBadRequest

	if req := p.fc.Request(); len(req.Header.ContentType()) == 0 {
		req.Header.SetContentType(fiber.MIMEApplicationJSON)
	}

	req := new(oas.EvaluationRequest)
	if p.err = p.fc.BodyParser(req); p.err != nil {
		p.msg = "invalid input data"
		return nil
	}

	if !p.checkEvaluationObject(req) {
		return nil
	}

	p.status = fiber.StatusOK
	return req
}

func (p *authProcess) checkEvaluationObject(req *oas.EvaluationObject) bool {
	if req.Subject.Type == "" || req.Subject.Id == "" {
		p.msg, p.err = "invalid subject", errors.New("subject type&id must be filled")
		return false
	}

	if req.Action.Name == "" {
		p.msg, p.err = "invalid action", errors.New("action name must be filled")
		return false
	}

	if req.Resource.Type == "" || req.Resource.Id == "" {
		p.msg, p.err = "invalid resource", errors.New("resource type&id must be filled")
		return false
	}

	return true
}

func (p *authProcess) createRequestAuthZEN(req *oas.EvaluationRequest) {
	principal := models.NewEntity(req.Subject.Type, req.Subject.Id, models.NewAttributeSet(req.Subject.Properties))
	action := models.NewEntity(models.EntityTypeName, req.Action.Name, models.NewAttributeSet(req.Action.Properties))
	resource := models.NewEntity(req.Resource.Type, req.Resource.Id, models.NewAttributeSet(req.Resource.Properties))
	ctx := models.NewAttributeSet(req.Context)

	if t := convert.AnyToDateTime(ctx.GetAttributeValue(models.AttrTime)); t.IsZero() {
		ctx.AddAttributeKVWithType(models.AttrTime, time.Now().UTC(), xsd.PrefixDateTime)
	}

	p.reqUID = uuid.New().String()
	p.parc = &models.PARC{
		Principal: principal,
		Action:    action,
		Resource:  resource,
		Context:   ctx,
	}
}

func (p *authProcess) authorizeRequestAuthZEN() error {
	if p.resp, p.err = p.controller.Authorize(p.reqUID, p.parc); p.err != nil {
		p.msg = "AuthZEN evaluation failed"
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

	return p.fc.JSON(&oas.EvaluationDecision{
		Decision: allowed,
		Context:  oas.ReasonObject{Id: "0", ReasonUser: oas.ReasonField{"en": msg}},
	})
}
