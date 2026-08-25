package fiber

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	auth "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/identity"
)

// authorizeRequest authorizes req and resolves the caller's Principal,
// falling back to the system sentinel when no authorizer is configured.
func authorizeRequest(authorizer authorization.Authorizer, req *fiber.Ctx, logger *slog.Logger) (identity.Principal, error) {
	if authorizer == nil {
		return identity.NewSystemPrincipal(), nil
	}

	resp, principal, err := authorizer.Authorize(auth.FormatRequest(req))

	return auth.Check(req, resp, principal, err, logger)
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
