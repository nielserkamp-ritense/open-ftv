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
	server "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/convert"
)

// ADLStrict controls the fail-closed behaviour of the durable authorization log
// (write-ahead log). When true (the default), a synchronous WAL-append failure is
// treated as fatal: the authorization decision is NOT returned, and the handler
// responds with a generic 500 instead. Accountability is a precondition for the
// decision - a decision whose durable record could not be written must not be
// acted upon. Set it to false to restore best-effort logging (the earlier, unsafe
// behaviour where a WAL failure was swallowed and the decision returned anyway).
//
// It is only consulted when an authLogger is configured (i.e. when ADL is enabled),
// so strict mode is effectively "default on whenever ADL is on".
var ADLStrict = true

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
	case p.walFailed:
		// A durable-log failure overrode the decision: report the true outcome.
		msg = "authorization failed closed: decision log (WAL) append failed"
		args = append(args, "status", p.status)
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
	if p.parc == nil {
		return // the request never reached the evaluation stage.
	}

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
		DecisionContext: models.NewAttributeSet(),
		TraceParent:     convert.AnyToString(p.parc.Context.GetAttributeValue(models.AttrTraceParent)),
		TraceState:      convert.AnyToString(p.parc.Context.GetAttributeValue(models.AttrTraceState)),
		RequestContext:  p.parc.Context,
		// F18: empty on the AuthZEN path (no FSC crossing), the header value on the FSC path.
		FSCTransactionID: p.fscTransactionID,
	}

	if p.resp != nil {
		rec.Decision = p.resp.Allowed
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
		if obs := obligationsFromAttributes(p.resp.Attributes["obligations"]); len(obs) > 0 {
			rec.DecisionContext.AddAttribute("obligations", obs)
		}
		if cs := p.resp.Attributes["consultedSources"]; cs != nil {
			// ADL level 3: the engine's exactly-consulted PIP sources for this decision.
			rec.DecisionContext.AddAttribute("consultedSources", cs)
		}
	} else {
		// The PDP produced no decision; per ADL the evaluation still yields one record,
		// with status Error when the process failed.
		rec.Errored = p.err != nil
	}

	if err := p.authLogger.Log(context.Background(), false, rec); err != nil {
		p.logger.Error("failed to write authlog", "record", rec, "error", err)

		// Fail-closed on accountability: the synchronous write-ahead log append
		// failed, so the decision was never durably recorded. In strict mode we must
		// not return that (now unverifiable) decision. authLog runs as a deferred
		// step before the fiber handler returns, so overriding the response here
		// replaces whatever body/status the decision path had already set with a
		// generic 500 - the caller learns nothing about the would-be decision.
		if ADLStrict && p.fc != nil {
			p.status = fiber.StatusInternalServerError
			p.walFailed = true
			_ = server.SendMessageResponse(p.fc, fiber.StatusInternalServerError, "internal server error")
		}
	}
}

type authProcess struct {
	status           int
	fc               *fiber.Ctx
	reqUID           string
	parc             *models.PARC
	resp             *models.Response
	logger           *slog.Logger
	authLogger       authlog.Logger
	controller       pdp.Controller
	started          time.Time
	err              error
	msg              string
	walFailed        bool   // set when a strict-mode WAL failure forced the request closed.
	fscTransactionID string // FSC Fsc-Transaction-Id header, set only on the FSC path.
}
