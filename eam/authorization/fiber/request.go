package fiber

import (
	"errors"
	"log/slog"
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authentication"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/identity"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

// FormatRequest formats an authorization request from a Fiber/FastHTTP request.
func FormatRequest(req *fiber.Ctx) *authorization.Request {
	uid := uuid.New()
	u, _ := url.Parse(string(req.Request().RequestURI()))

	return &authorization.Request{
		UID:     &uid,
		URL:     u,
		Method:  req.Method(),
		Headers: req.GetReqHeaders(),
		Body:    req.Body(),
	}
}

// FormatRequestWithResource formats an authorization request like FormatRequest but additionally attaches the stored
// target object so the PDP can evaluate fine-grained, resource-attribute policies.
func FormatRequestWithResource(req *fiber.Ctx, resource *models.Entity) *authorization.Request {
	r := FormatRequest(req)
	r.Resource = resource
	return r
}

// Check verifies an authorization attempt, returning the principal to attribute the caller's action to,
// whether access was allowed, and an error response written to req on failure.
func Check(req *fiber.Ctx, resp *models.Response, principal identity.Principal, err error, log *slog.Logger) (identity.Principal, bool, error) {
	if principal == identity.NewUnknownPrincipal() {
		principal = identity.NewSystemPrincipal()
	}

	switch {
	case err != nil && authenticationError(err):
		if !strings.Contains(err.Error(), "api-key") {
			req.Set(fiber.HeaderWWWAuthenticate, "Basic realm=OpenFTV")
		}

		msg := "authentication failed" // 401
		log.Error(msg, "path", req.Path(), "err", err)

		return principal, false, fiber.NewError(fiber.StatusUnauthorized, msg)

	case err != nil || resp == nil || !resp.Allowed:
		msg := "authorization failed" // 403
		log.Error(msg, "path", req.Path(), "authResponse", resp, "err", err)

		return principal, false, fiber.NewError(fiber.StatusForbidden, msg)

	default:
		return principal, true, nil
	}
}

func authenticationError(err error) bool {
	var e *authentication.ErrUnauthenticated
	return errors.As(err, &e)
}
