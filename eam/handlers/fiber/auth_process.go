package fiber

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller/adl"
	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/authzen"
)

func initAuthProcess(fc *fiber.Ctx, logger *slog.Logger, decisionLog *adl.ADL, controller pdp.Controller) (*authProcess, func()) {
	a := &authProcess{
		logger:     logger,
		adl:        decisionLog,
		controller: controller,
		fc:         fc,
		status:     fiber.StatusInternalServerError,
		started:    time.Now(),
	}
	return a, a.finish
}

func (p *authProcess) finish() {
	if p.logger.Enabled(nil, slog.LevelInfo) {
		p.log()
	}
	if p.adl != nil && p.authReq != nil && p.authResp != nil {
		p.logDecision(p.fc.UserContext())
	}
}

func (p *authProcess) log() {
	args := make([]any, 0, 24)

	args = append(args, "request-uid", p.reqUID)

	switch {
	case p.batch != nil:
		args = append(args, "#items", len(p.batch.Items))
		args = append(args, "semantics", p.batch.Semantics.String())
	case p.parc != nil:
		args = append(args, "subject", p.parc.Principal.Type())
		args = append(args, "action", p.parc.Action.Type())
		args = append(args, "resource", p.parc.Resource.Type())
	}

	var allowed bool
	if p.resp != nil {
		allowed = p.resp.Allowed
		args = append(args, "allowed", allowed)

		if p.resp.Message != "" {
			args = append(args, "message", p.resp.Message)
		}
		if p.resp.PolicyKey != "" {
			args = append(args, "policy-key", p.resp.PolicyKey)
		}
		if p.resp.PolicyHash != "" {
			args = append(args, "policy-hash", p.resp.PolicyHash)
		}
	}

	var msg string
	switch {
	case p.err != nil:
		msg = "authorization process failed"
		args = append(args, "status", p.status, "error", p.err)
	case p.resp == nil:
		msg = "authorization processed"
	case !allowed:
		msg = "authorization denied"
	default:
		msg = "authorization granted"
	}

	args = append(args, "elapsed time", time.Since(p.started).String())
	p.logger.Info(msg, args...)
}

func (p *authProcess) logDecision(ctx context.Context) {
	var err error

	switch t := p.authReq.(type) {
	case *oas.EvaluationRequest:
		err = p.adl.Evaluation(ctx, p.started, t, p.authResp.(*oas.EvaluationResponse))
	case *oas.EvaluationsRequest:
		err = p.adl.Evaluations(ctx, p.started, t, p.authResp.(*oas.EvaluationsResponse))
	case *oas.SearchRequest:
		if p.search == searchSubject {
			err = p.adl.SearchSubject(ctx, p.started, t, p.authResp.(*oas.SearchResponse))
		} else {
			err = p.adl.SearchResource(ctx, p.started, t, p.authResp.(*oas.SearchResponse))
		}
	case *oas.SearchActionRequest:
		err = p.adl.SearchAction(ctx, p.started, t, p.authResp.(*oas.SearchActionResponse))
	default:
		err = fmt.Errorf("invalid request type: %T", t)
	}

	if err != nil {
		p.logger.Error("failed to write authorization decision log", "error", err)
	}
}

type authProcess struct {
	status     int
	ctx        context.Context
	fc         *fiber.Ctx
	reqUID     string
	parc       *models.PARC
	batch      *models.Batch
	search     searchType
	resp       *models.Response
	logger     *slog.Logger
	adl        *adl.ADL
	controller pdp.Controller
	started    time.Time
	err        error
	msg        string
	authReq    any
	authResp   any
	offset     int
	limit      int
}

type searchType uint8

const (
	searchSubject searchType = iota + 1
	searchAction
	searchResource
)

// String implements the Stringer interface.
func (s searchType) String() string {
	switch s {
	case searchSubject:
		return "subject"
	case searchAction:
		return "action"
	case searchResource:
		return "resource"
	default:
		return "???"
	}
}
