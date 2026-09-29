package fiber

import (
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v2"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	auth "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/identity"
	server "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server/fiber"
)

// authorizeRequest authorizes req and resolves the caller's Principal,
// falling back to the system sentinel when no authorizer is configured.
func authorizeRequest(authorizer authorization.RequestAuthorizer, req *fiber.Ctx, logger *slog.Logger) (identity.Principal, error) {
	if authorizer == nil {
		return identity.NewSystemPrincipal(), nil
	}

	resp, principal, err := authorizer.AuthorizeRequest(auth.FormatRequest(req))

	return auth.Check(req, resp, principal, err, logger)
}

var (
	errInternal             = errors.New("internal error")
	errPrincipalNotRecorded = errors.New("failed to record principal")
)

// internalError answers a failure the caller cannot act on with 500. The cause is logged and
// never sent: a store or PDP error names hosts, tables and constraints. A permitted caller
// that could not be recorded (ADR 0004) is the one cause named, so an operator can act on it.
func internalError(req *fiber.Ctx, logger *slog.Logger, err error) error {
	logger.Error("request error", "path", req.Path(), "err", err, "status", fiber.StatusInternalServerError)

	var notRecorded *authorization.ErrPrincipalNotRecorded
	if errors.As(err, &notRecorded) {
		return server.SendMessageResponse(req, fiber.StatusInternalServerError, errPrincipalNotRecorded.Error())
	}

	return server.SendMessageResponse(req, fiber.StatusInternalServerError, errInternal.Error())
}

// callerOf returns the request-scoped Principal the Identify middleware attached.
//
// Without it the handler runs as the system principal: right for a route without an
// authorizer, and a wiring mistake for a secured one, which is logged so it is not silent.
func callerOf(req *fiber.Ctx, secured bool, logger *slog.Logger) *authorization.RequestPrincipal {
	if p, ok := authorization.RequestPrincipalFromContext(req.UserContext()); ok {
		return p
	}

	if secured {
		logger.Error("no principal in the request context: the route is missing the Identify middleware; running as the system principal", "path", req.Path())
	}

	return authorization.SystemRequestPrincipal()
}
