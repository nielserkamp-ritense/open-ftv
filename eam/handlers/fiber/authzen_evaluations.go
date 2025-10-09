package fiber

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	server "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server/fiber"
	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/authzen"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/convert"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/xsd"
)

// Evaluations implements the AuthZENAuthorizer interface.
func (h *AuthZENAuthorizer) Evaluations(fc *fiber.Ctx) error {
	processHeaders(fc)

	if !h.hasEvaluations {
		return server.SendMessageResponse(fc, fiber.StatusNotImplemented, utils.StatusMessage(fiber.StatusNotImplemented))
	}

	p, finish := initAuthProcess(fc, h.logger, h.adl, h.controller)
	defer finish()

	batch := p.verifyBatchAuthZEN()
	if p.err != nil {
		p.logger.Error("AuthZEN batch authorization handler failed", "request", batch, "error", p.err)
		return server.SendMessageResponse(fc, p.status, p.msg)
	}
	p.authReq = batch

	p.createBatchAuthZEN(batch)
	p.logger.Debug("AuthZEN evaluations request", "request", p.batch)

	return p.authorizeBatchAuthZEN()
}

func (p *authProcess) verifyBatchAuthZEN() *oas.EvaluationsRequest {
	p.initVerification()

	req := new(oas.EvaluationsRequest)
	if p.err = p.fc.BodyParser(req); p.err != nil {
		p.msg = "invalid input data"
		return nil
	}

	if len(req.Evaluations) == 0 {
		p.msg, p.err = "invalid batch request", errors.New("evaluations array must not be empty")
		return nil
	}

	for i := range req.Evaluations {
		e := &req.Evaluations[i]
		fixDefaults(e, req)

		if !p.checkEvaluationObject(e) {
			p.msg, p.err = "invalid batch request", fmt.Errorf("evaluation #%d invalid: %w", i, p.err)
			return nil
		}
	}

	p.status = fiber.StatusOK
	return req
}

func fixDefaults(entity *oas.EvaluationObject, req *oas.EvaluationsRequest) {
	fixDefaultEntity(&entity.Subject, &req.Subject)
	fixDefaultEntity(&entity.Resource, &req.Resource)

	if entity.Action.Name == "" {
		entity.Action.Name = req.Action.Name
	}
	if len(entity.Action.Properties) == 0 {
		entity.Action.Properties = req.Action.Properties
	}
}

func fixDefaultEntity(entity, def *oas.Entity) {
	if entity.Type == "" {
		entity.Type = def.Type
	}
	if entity.Id == "" {
		entity.Id = def.Id
	}
	if len(entity.Properties) == 0 {
		entity.Properties = def.Properties
	}
}

func (p *authProcess) createBatchAuthZEN(req *oas.EvaluationsRequest) {
	p.batch = &models.Batch{Items: make([]models.PARC, 0, len(req.Evaluations))}

	switch strings.ToLower(req.Options.EvaluationSemantics) {
	case "deny_on_first_deny":
		p.batch.Semantics = models.DenyOnFirstDeny
	case "permit_on_first_permit":
		p.batch.Semantics = models.PermitOnFirstPermit
	default:
		p.batch.Semantics = models.EvaluateAll
	}

	now := time.Now().UTC()

	for i := range req.Evaluations {
		e := &req.Evaluations[i]

		principal := models.NewEntity(e.Subject.Type, e.Subject.Id, models.NewAttributeSet(e.Subject.Properties))
		action := models.NewEntity(models.EntityTypeName, e.Action.Name, models.NewAttributeSet(e.Action.Properties))
		resource := models.NewEntity(e.Resource.Type, e.Resource.Id, models.NewAttributeSet(e.Resource.Properties))
		ctx := models.NewAttributeSet(e.Context)

		if t := convert.AnyToDateTime(ctx.GetAttributeValue(models.AttrTime)); t.IsZero() {
			ctx.AddAttributeKVWithType(models.AttrTime, now, xsd.PrefixDateTime)
		}

		p.batch.Items = append(p.batch.Items, models.PARC{
			Principal: principal,
			Action:    action,
			Resource:  resource,
			Context:   ctx,
		})
	}
}

func (p *authProcess) authorizeBatchAuthZEN() error {
	var results []models.Response

	p.controller.GetPIP().ReportDynamicData(func() {
		results, p.err = p.controller.Batch(p.reqUID, p.batch)
	})

	if p.err != nil {
		p.msg = "AuthZEN evaluations failed"
		return server.SendMessageResponse(p.fc, p.status, p.msg)
	}

	out := &oas.EvaluationsResponse{Evaluations: make([]oas.EvaluationDecision, 0, len(results))}

	for i := range results {
		resp := &results[i]

		allowed, msg := resp.Allowed, resp.Message
		if msg == "" {
			if allowed {
				msg = "ok"
			} else {
				msg = "not authorized"
			}
		}

		out.Evaluations = append(out.Evaluations, oas.EvaluationDecision{
			Decision: resp.Allowed,
			Context:  oas.ReasonObject{Id: "0", ReasonUser: oas.ReasonField{"en": msg}},
		})
	}

	p.authResp = out
	return p.fc.JSON(p.authResp)
}
