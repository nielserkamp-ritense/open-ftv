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
		// if the API version was not set previously, we set a generic version number.
		if s := req.Get("API-Version"); s == "" {
			req.Set("API-Version", "1.0.0")
		}

		status := fiber.StatusInternalServerError

		var message string

		if fe, ok := errors.AsType[*fiber.Error](err); ok {
			status, message = fe.Code, fe.Message
		} else {
			message = "An unexpected error occurred."
		}

		if status >= fiber.StatusInternalServerError {
			logger.Error("internal server error", "status", status, "error", err)
		}

		if message != "" {
			return SendMessageResponse(req, status, message)
		}

		return SendBasicResponse(req, status)
	}
}
