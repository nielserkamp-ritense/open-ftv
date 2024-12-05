package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/components"
)

func (p *authProcess) log() {
	args := make([]any, 0, 16)

	if p.req != nil {
		args = append(args, "method", p.req.Method)
		args = append(args, "request-uid", p.req.UID)
	}

	if p.resp != nil {
		args = append(args, "allowed", p.resp.Allowed, "policy", p.resp.PolicyKey)
	}

	var msg string
	if p.err != nil {
		msg = "authorization process failed"
		args = append(args, "status", p.status, "error", p.err)
	} else {
		msg = "authorization process successful"
	}

	args = append(args, "elapsed time", time.Since(p.started).String())
	p.logger.Info(msg, args...)
}

type authProcess struct {
	status  int
	fc      *fiber.Ctx
	req     *components.Request
	resp    *components.Response
	started time.Time
	err     error
	msg     string
	authHandler
}
