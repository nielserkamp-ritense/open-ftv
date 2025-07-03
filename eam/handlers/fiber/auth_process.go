package fiber

import (
	"context"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/authlog"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pep"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/convert"
)

func (p *authProcess) log() {
	args := make([]any, 0, 16)

	if p.parc != nil {
		args = append(args, "method", convert.AnyToString(p.parc.Action.Attributes().GetAttribute(models.AttrMethod)))
		args = append(args, "request-uid", p.reqUID)
	}

	var allowed bool
	if p.resp != nil {
		allowed = p.resp.Allowed
		args = append(args, "allowed", allowed, "message", p.resp.Message, "policy", p.resp.PolicyKey)
	}

	var msg string
	switch {
	case p.err != nil:
		msg = "authorization process failed"
		args = append(args, "status", p.status, "error", p.err)
	case !allowed:
		msg = "authorization denied"
	default:
		msg = "authorization granted"
	}

	args = append(args, "elapsed time", time.Since(p.started).String())
	p.logger.Info(msg, args...)
}

func (p *authProcess) authLog() {
	clientIP := convert.AnyToString(p.parc.Context.GetAttributeValue(models.AttrClientIP))

	rvvaID := convert.AnyToString(p.parc.Context.GetAttributeValue(models.AttrRvvaID))
	if rvvaID == "" && p.parc.Principal != nil && p.parc.Principal.Type() == pep.PrincipalRVVA {
		rvvaID = p.parc.Principal.ID()
	}

	t := convert.AnyToDateTime(p.parc.Context.GetAttributeValue(models.AttrTime))
	if t.IsZero() {
		t = time.Now().UTC()
	}

	rec := &authlog.AuthRecord{
		ClientIP:        clientIP,
		RequestTime:     &t,
		RvvaID:          rvvaID,
		Principal:       p.parc.Principal,
		Action:          p.parc.Action,
		Resource:        p.parc.Resource,
		Decision:        p.resp.Allowed,
		DecisionContext: models.NewAttributeSet(),
		TraceParent:     convert.AnyToString(p.parc.Context.GetAttributeValue(models.AttrTraceParent)),
		TraceState:      convert.AnyToString(p.parc.Context.GetAttributeValue(models.AttrTraceState)),
	}

	if p.resp.Message != "" {
		rec.DecisionContext.AddAttribute("message", p.resp.Message)
	}
	if p.resp.PolicyKey != "" {
		rec.DecisionContext.AddAttribute("policy", p.resp.PolicyKey)
	}
	if p.resp.PolicyHash != "" {
		rec.DecisionContext.AddAttribute("policyHash", p.resp.PolicyHash)
	}
	if diag := p.resp.Attributes["diagnostic"]; diag != nil {
		rec.DecisionContext.AddAttribute("diagnostic", diag)
	}

	if err := p.authLogger.Log(context.Background(), false, rec); err != nil {
		p.logger.Error("failed to write authlog", "record", rec, "error", err)
	}
}

type authProcess struct {
	status     int
	fc         *fiber.Ctx
	reqUID     string
	parc       *models.PARC
	resp       *models.Response
	logger     *slog.Logger
	authLogger authlog.Logger
	controller pdp.Controller
	started    time.Time
	err        error
	msg        string
}
