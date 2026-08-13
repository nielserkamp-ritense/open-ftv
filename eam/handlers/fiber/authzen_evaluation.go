package fiber

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	server "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server/fiber"
	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/authzen"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/convert"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/xsd"
)

// Evaluation implements the AuthZENAuthorizer interface.
func (h *AuthZENAuthorizer) Evaluation(fc *fiber.Ctx) error {
	processHeaders(fc)

	p, finish := initAuthProcess(fc, h.logger, h.adl, h.controller)
	defer finish()

	req := p.verifyRequestAuthZEN()
	if p.err != nil {
		p.logger.Error("AuthZEN authorization handler failed", "request", req, "error", p.err)
		return server.SendMessageResponse(fc, p.status, p.msg)
	}

	p.createRequestAuthZEN(req)
	p.logger.Debug("AuthZEN evaluation request", "request", p.parc)

	return p.authorizeRequestAuthZEN()
}

func (p *authProcess) verifyRequestAuthZEN() *oas.EvaluationRequest {
	p.initVerification()

	req := new(oas.EvaluationRequest)
	if p.err = p.fc.BodyParser(req); p.err != nil {
		p.msg = "invalid input data"
		return nil
	}

	// From here on the body is a coherent AuthZEN object, even if it fails the checks below,
	// then log it (Logius ADL §3.3.6 lists missing attributes as an Error case).
	p.authReq = req
	applyTraceFallback(p.fc, models.NewAttributeSet(req.Context))

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

	p.parc = &models.PARC{Principal: principal, Action: action, Resource: resource, Context: ctx}
}

func (p *authProcess) authorizeRequestAuthZEN() error {
	p.controller.GetPIP().ReportDynamicData(func() {
		p.resp, p.err = p.controller.Authorize(p.reqUID, p.parc)
	})
	p.decided = time.Now()

	if p.err != nil {
		p.msg = "AuthZEN evaluation failed"
		p.logger.Error(p.msg, "request", p.parc, "error", p.err)
		return server.SendMessageResponse(p.fc, p.status, p.msg)
	}

	p.authResp = &oas.EvaluationDecision{
		Decision: p.resp.Allowed,
		Context:  reasonContext(p.resp),
	}
	return p.fc.JSON(p.authResp)
}
