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
