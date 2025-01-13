package handlers

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/components"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/components/pdp"
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

type authProcess struct {
	status     int
	fc         *fiber.Ctx
	req        *components.Request
	resp       *components.Response
	logger     *slog.Logger
	controller pdp.Controller
	started    time.Time
	err        error
	msg        string
}
