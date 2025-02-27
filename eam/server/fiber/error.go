package fiber

import (
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v2"
)

// ErrorHandler is the default error handler for things gone awry in fiber.
// E.g. invalid paths, bad parameters, code panics, etc.
// The actual error gets logged while the error response body is based on the embedded status code.
// If the error does not embed a status code, InternalServerError will be used as the status.
func ErrorHandler(logger *slog.Logger) fiber.ErrorHandler {
	return func(req *fiber.Ctx, err error) error {
		// we can't be sure of the requested route, so a generic version number will have to do.
		req.Set("API-Version", "1.0.0")

		status := fiber.StatusInternalServerError

		var e *fiber.Error
		if errors.As(err, &e) {
			status = e.Code
		}

		logger.Error("internal server error", "status", status, "error", err)
		return SendBasicResponse(req, status)
	}
}
