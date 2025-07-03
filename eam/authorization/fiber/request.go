package fiber

import (
	"errors"
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authentication"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	server "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server/fiber"
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

// Check verifies the response from an authorization test.
//
// The function returns true if authorization was successful, otherwise it returns false.
//
// It will automatically encode an appropriate error message on the request if authorization failed.
func Check(req *fiber.Ctx, resp *models.Response, err error) (bool, error) {
	switch {
	case err != nil && authenticationError(err):
		if !strings.Contains(err.Error(), "api-key") {
			req.Set(fiber.HeaderWWWAuthenticate, "Basic realm=OpenFTV")
		}
		return false, server.SendMessageResponse(req, fiber.StatusUnauthorized, "authentication failed")

	case err != nil || resp == nil || !resp.Allowed:
		return false, server.SendMessageResponse(req, fiber.StatusForbidden, "authorization failed")

	default:
		return true, nil
	}
}

func authenticationError(err error) bool {
	var e *authentication.ErrUnauthenticated
	return errors.As(err, &e)
}
