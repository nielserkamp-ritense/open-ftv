package fiber

import (
	"context"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/log/authlog"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pdp"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/convert"
)

func (p *authProcess) log() {
	args := make([]any, 0, 16)

	if p.req != nil {
		args = append(args, "method", p.req.Method)
		args = append(args, "request-uid", p.req.UID)
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
	clientIP, _ := p.req.Attributes[models.AttrClientIP].(string)

	rvvaID, _ := p.req.Attributes[models.AttrRvvaID].(string)
	if rvvaID == "" && p.req.Principal != nil && p.req.Principal.Type() == models.PrincipalRVVA {
		rvvaID = p.req.Principal.ID()
	}

	var t time.Time
	if p.req.RequestTime != nil {
		t = *p.req.RequestTime
	} else {
		t = time.Now().UTC()
	}

	rec := &authlog.AuthRecord{
		ClientIP:        clientIP,
		RequestTime:     &t,
		RvvaID:          rvvaID,
		Principal:       p.req.Principal,
		Action:          p.req.Action,
		Resource:        p.req.Resource,
		Decision:        p.resp.Allowed,
		DecisionContext: models.NewAttributeSet(),
		TraceParent:     convert.AnyToString(p.req.Attributes[models.AttrTraceParent]),
		TraceState:      convert.AnyToString(p.req.Attributes[models.AttrTraceState]),
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
	req        *components.Request
	resp       *components.Response
	logger     *slog.Logger
	authLogger authlog.Logger
	controller pdp.Controller
	started    time.Time
	err        error
	msg        string
}
