// Package fiber contains functionality for handling HTTP requests using Fiber/FastHTTP.
package fiber

import (
	"sync"

	"github.com/gofiber/fiber/v2"

	server "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server/fiber"
)

// HealthVersion is the full semantic API version for the health endpoints.
const HealthVersion = "1.0.0"

// NewChecks instantiates new health, liveliness and readiness endpoints.
func NewChecks() *Checks {
	return &Checks{}
}

// HealthZ is the endpoint for health checks.
//
// There are no internal service checks, so for now it will always respond with a static OK message.
func (h *Checks) HealthZ(fc *fiber.Ctx) error {
	fc.Set(HeaderVersion, HealthVersion)

	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if h.healthy {
		return server.SendBasicResponse(fc, fiber.StatusOK)
	}
	return server.SendMessageResponse(fc, fiber.StatusServiceUnavailable, "not healthy yet")
}

// LiveZ is the endpoint for liveliness checks.
//
// There are no internal service checks, so for now it will always respond with a static OK message.
func (h *Checks) LiveZ(fc *fiber.Ctx) error {
	fc.Set(HeaderVersion, HealthVersion)

	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if h.lively {
		return server.SendBasicResponse(fc, fiber.StatusOK)
	}
	return server.SendMessageResponse(fc, fiber.StatusServiceUnavailable, "not alive yet")
}

// ReadyZ is the endpoint for readiness checks.
//
// There are no internal service checks, so for now it will always respond with a static OK message.
func (h *Checks) ReadyZ(fc *fiber.Ctx) error {
	fc.Set(HeaderVersion, HealthVersion)

	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if h.ready {
		return server.SendBasicResponse(fc, fiber.StatusOK)
	}
	return server.SendMessageResponse(fc, fiber.StatusServiceUnavailable, "not ready yet")
}

// SetHealth updates the health status.
func (h *Checks) SetHealth(in bool) {
	h.mutex.Lock()
	h.healthy = in
	h.mutex.Unlock()
}

// SetAlive updates the live status.
func (h *Checks) SetAlive(in bool) {
	h.mutex.Lock()
	h.lively = in
	h.mutex.Unlock()
}

// SetReady updates the ready status.
func (h *Checks) SetReady(in bool) {
	h.mutex.Lock()
	h.ready = in
	h.mutex.Unlock()
}

// Checks represents all health, liveliness and readiness endpoints.
type Checks struct {
	healthy bool
	lively  bool
	ready   bool
	mutex   sync.RWMutex
}
