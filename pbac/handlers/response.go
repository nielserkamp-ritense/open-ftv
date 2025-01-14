// Package handlers contains generic HTTP request handlers.
package handlers

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
	return req.Status(status).Send(responseBody[status])
}

// SendMessageResponse sends a basic response corresponding with the given status code.
func SendMessageResponse(req *fiber.Ctx, status int, msg string) error {
	return req.Status(status).JSON(&authzen.ErrorResponse{Message: msg})
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
	responseBody = make(map[int][]byte, len(supportedStatus))
)

func init() {
	for _, status := range supportedStatus {
		responseBody[status], _ = json.Marshal(&authzen.ErrorResponse{Message: utils.StatusMessage(status)})
	}
}
