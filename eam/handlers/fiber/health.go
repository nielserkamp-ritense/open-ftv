package fiber

import (
	"github.com/gofiber/fiber/v2"
)

// HealthZ is the endpoint for health checks.
//
// There are no internal service checks, so for now it will always respond with a static OK message.
func HealthZ(fc *fiber.Ctx) error {
	return SendBasicResponse(fc, fiber.StatusOK)
}
