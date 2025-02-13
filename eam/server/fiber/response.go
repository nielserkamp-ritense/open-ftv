// Package fiber contains HTTP request handlers for integration with Fiber/FastHTTP based services.
package fiber

import (
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/oas/authzen"
)

// SendBasicResponse sends a basic response corresponding with the given status code.
// If the given status code is not in the supported statuses list, an empty body will be sent.
func SendBasicResponse(req *fiber.Ctx, status int) error {
	req.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	return req.Status(status).Send(ResponseBody[status])
}

// SendMessageResponse sends a basic response corresponding with the given status code.
func SendMessageResponse(req *fiber.Ctx, status int, msg string) error {
	return req.Status(status).JSON(&authzen.ErrorResponse{Title: msg})
}

var (
	supportedStatus = [...]int{
		fiber.StatusOK,
		fiber.StatusCreated,
		fiber.StatusNoContent,
		fiber.StatusBadRequest,
		fiber.StatusUnauthorized,
		fiber.StatusForbidden,
		fiber.StatusNotFound,
		fiber.StatusConflict,
		fiber.StatusInternalServerError,
	}

	// ResponseBody contains preformatted response bodies for supported status codes.
	ResponseBody = make(map[int][]byte, len(supportedStatus))
)

func init() {
	for _, status := range supportedStatus {
		ResponseBody[status], _ = json.Marshal(&authzen.ErrorResponse{Title: utils.StatusMessage(status)})
	}
}
