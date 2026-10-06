package fiber

import (
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

	batch, single := p.verifyBatchAuthZEN()
	if p.err != nil {
		p.logger.Error("AuthZEN batch authorization handler failed", "request", p.authReq, "error", p.err)
		return server.SendMessageResponse(fc, p.status, p.msg)
	}

	if single != nil {
		p.createRequestAuthZEN(single)
		p.logger.Debug("AuthZEN evaluation request", "request", p.parc)

		return p.authorizeRequestAuthZEN()
	}

	p.createBatchAuthZEN(batch)
	p.logger.Debug("AuthZEN evaluations request", "request", p.batch)

	return p.authorizeBatchAuthZEN()
}

// verifyBatchAuthZEN returns either the batch, or, when it holds no evaluations, the single evaluation
// its top-level keys describe: AuthZEN 1.0 treats such a request as an Access Evaluation request.
func (p *authProcess) verifyBatchAuthZEN() (batch *oas.EvaluationsRequest, single *oas.EvaluationRequest) {
	p.initVerification()

	req := new(oas.EvaluationsRequest)
	if p.err = p.fc.BodyParser(req); p.err != nil {
		p.msg = "invalid input data"
		return nil, nil
	}

	if len(req.Evaluations) == 0 {
		single := new(oas.EvaluationRequest)
		fixDefaults(single, req)

		return nil, p.acceptRequestAuthZEN(single)
	}

	// From here on the body is a coherent AuthZEN object, even if it fails the checks below,
	// then log it (Logius ADL §3.3.6 lists missing attributes as an Error case).
	p.authReq = req
	applyTraceFallback(p.fc, models.NewAttributeSet(req.Context))

	for i := range req.Evaluations {
		e := &req.Evaluations[i]
		fixDefaults(e, req)

		if !p.checkEvaluationObject(e) {
			p.msg, p.err = "invalid batch request", fmt.Errorf("evaluation #%d invalid: %w", i, p.err)
			return nil, nil
		}
	}

	p.status = fiber.StatusOK

	return req, nil
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

	if len(entity.Context) == 0 {
		entity.Context = req.Context
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

	switch strings.ToLower(req.Options.EvaluationsSemantic) {
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
	p.decided = time.Now()

	if p.err != nil {
		p.status = fiber.StatusInternalServerError
		p.msg = "AuthZEN evaluations failed"
		p.logger.Error(p.msg, "request", p.batch, "error", p.err)
		return server.SendMessageResponse(p.fc, p.status, p.msg)
	}

	out := &oas.EvaluationsResponse{Evaluations: make([]oas.EvaluationDecision, 0, len(results))}

	for i := range results {
		resp := &results[i]

		out.Evaluations = append(out.Evaluations, oas.EvaluationDecision{
			Decision: resp.Allowed,
			Context:  resp.Context,
		})
	}

	p.authResp = out
	return p.fc.JSON(p.authResp)
}
