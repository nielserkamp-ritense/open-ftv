package fiber

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

// Identify is the middleware that validates the caller once per request and hands the
// request-scoped Principal, with the request's trace context, to the handlers behind it.
// The handlers decide through the management-plane services (ADR 0006).
//
// A caller that fails authentication is answered with 401 here; a request without any
// credentials passes as the system principal and is denied by the PDP later, so that
// fail-open deployments keep working.
func Identify(a authorization.Authorizer, log *slog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		req := FormatRequest(c)
		req.Body = nil // identity never depends on the body.

		caller, err := a.Identify(req)
		if err != nil {
			return unauthenticated(c, err, log)
		}

		ctx := authorization.WithRequestPrincipal(c.UserContext(), caller)
		ctx = authorization.WithTrace(ctx, authorization.Trace{
			Parent: c.Get(models.HeaderTraceParent),
			State:  c.Get(models.HeaderTraceState),
		})
		c.SetUserContext(ctx)

		return c.Next()
	}
}
