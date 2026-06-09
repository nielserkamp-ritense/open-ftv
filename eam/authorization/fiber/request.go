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
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	server "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/convert"
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

// FormatRequestWithResource formats an authorization request like FormatRequest
// but additionally attaches the stored target object so the PDP can evaluate
// fine-grained, resource-attribute policies.
func FormatRequestWithResource(req *fiber.Ctx, resource *models.Entity) *authorization.Request {
	r := FormatRequest(req)
	r.Resource = resource
	return r
}

// Check verifies the response from an authorization test.
//
// The function returns true if authorization was successful, otherwise it returns false.
//
// It will automatically encode an appropriate error message on the request if authorization failed.
func Check(req *fiber.Ctx, resp *models.Response, err error, log *slog.Logger) (string, bool, error) {
	user := SystemUser
	if resp != nil {
		if u, ok := resp.Attributes["user"]; ok {
			user = convert.AnyToString(u)
		}
	}

	switch {
	case err != nil && authenticationError(err):
		if !strings.Contains(err.Error(), "api-key") {
			req.Set(fiber.HeaderWWWAuthenticate, "Basic realm=OpenFTV")
		}

		msg := "authentication failed" // 401
		log.Error(msg, "path", req.Path(), "err", err)
		return user, false, server.SendMessageResponse(req, fiber.StatusUnauthorized, msg)

	case err != nil || resp == nil || !resp.Allowed:
		msg := "authorization failed" // 403
		log.Error(msg, "path", req.Path(), "authResponse", resp, "err", err)
		return user, false, server.SendMessageResponse(req, fiber.StatusForbidden, msg)

	default:
		return user, true, nil
	}
}

func authenticationError(err error) bool {
	var e *authentication.ErrUnauthenticated
	return errors.As(err, &e)
}

// SystemUser represents the default code for an unknown user.
const SystemUser = "*SYSTEM*"
