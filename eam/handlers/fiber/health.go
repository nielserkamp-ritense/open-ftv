// Package fiber contains functionality for handling HTTP requests using Fiber/FastHTTP.
package fiber

import (
	"github.com/gofiber/fiber/v2"

	fiber2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/server/fiber"
)

// HealthZ is the endpoint for health checks.
//
// There are no internal service checks, so for now it will always respond with a static OK message.
func HealthZ(fc *fiber.Ctx) error {
	return fiber2.SendBasicResponse(fc, fiber.StatusOK)
}
