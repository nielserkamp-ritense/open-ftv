package handlers

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/oas/policies"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/components/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/components/pdp"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/fsc/plugin/generic/config"
)

// PoliciesHandler represents the interface for handling requests about policies.
type PoliciesHandler interface {
	GetPolicies(req *fiber.Ctx) error
	GetPolicy(req *fiber.Ctx) error
	PutPolicy(req *fiber.Ctx) error
	PostPolicy(req *fiber.Ctx) error
	DeletePolicy(req *fiber.Ctx) error
}

// NewPoliciesHandler instantiates a policy handler.
func NewPoliciesHandler(cfg *config.Config, logger *slog.Logger, c pdp.Controller) PoliciesHandler {
	return &policiesHandler{cfg: cfg, logger: logger, c: c}
}

// GetPolicies implements the PoliciesHandler interface.
func (h *policiesHandler) GetPolicies(req *fiber.Ctx) error {
	list := h.c.PAP().ListAllKeys()
	resp := make(policies.PoliciesResponse, 0, len(list))

	for i := range list {
		key := list[i]
		if pol, err := h.c.PAP().Get(key); err == nil {
			resp = append(resp, h.convertPolicy(pol))
		}
	}

	return req.JSON(&resp)
}

// GetPolicy implements the PoliciesHandler interface.
func (h *policiesHandler) GetPolicy(req *fiber.Ctx) error {
	key := req.Params("id")
	if key == "" || len(key) > 500 {
		return SendMessageResponse(req, fiber.StatusBadRequest, "id must be filled")
	}

	pol, err := h.c.PAP().Get(key)
	if err != nil {
		return SendMessageResponse(req, fiber.StatusNotFound, err.Error())
	}
	return req.JSON(h.convertPolicy(pol))
}

// PutPolicy implements the PoliciesHandler interface.
func (h *policiesHandler) PutPolicy(req *fiber.Ctx) error {
	key := req.Params("id")
	if key == "" || len(key) > 500 {
		return SendMessageResponse(req, fiber.StatusBadRequest, "id must be filled")
	}

	var p policies.Policy
	if err := req.BodyParser(&p); err != nil {
		return SendMessageResponse(req, fiber.StatusBadRequest, err.Error())
	}

	return SendMessageResponse(req, fiber.StatusNotImplemented, "not implemented")
}

// PostPolicy implements the PoliciesHandler interface.
func (h *policiesHandler) PostPolicy(req *fiber.Ctx) error {
	key := req.Params("id")
	if key == "" || len(key) > 500 {
		return SendMessageResponse(req, fiber.StatusBadRequest, "id must be filled")
	}

	var p policies.Policy
	if err := req.BodyParser(&p); err != nil {
		return SendMessageResponse(req, fiber.StatusBadRequest, err.Error())
	}

	return SendMessageResponse(req, fiber.StatusNotImplemented, "not implemented")
}

// DeletePolicy implements the PoliciesHandler interface.
func (h *policiesHandler) DeletePolicy(req *fiber.Ctx) error {
	key := req.Params("id")
	if key == "" || len(key) > 500 {
		return SendMessageResponse(req, fiber.StatusBadRequest, "id must be filled")
	}

	ignore := req.QueryBool("ignoreMissing")

	if _, err := h.c.PAP().Get(key); err != nil {
		if ignore {
			return req.JSON(&policies.Policy{Id: key})
		} else {
			return SendMessageResponse(req, fiber.StatusNotFound, err.Error())
		}
	}

	pol, err := h.c.PAP().Remove(key)
	if err != nil {
		return SendMessageResponse(req, fiber.StatusInternalServerError, err.Error())
	}
	return req.JSON(h.convertPolicy(pol))
}

func (h *policiesHandler) convertPolicy(pol pap.Policy) policies.Policy {
	return policies.Policy{
		Id:       pol.ID(),
		Language: h.c.String(),
		RvvaID:   pol.RvvaID(),
		Source:   pol.Source(),
		Target:   pol.Target(),
		Url:      pol.URI(),
	}
}

type policiesHandler struct {
	cfg    *config.Config
	logger *slog.Logger
	c      pdp.Controller
}
