package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/oas/policies"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/components/pap"
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
func NewPoliciesHandler(logger *slog.Logger, cache pap.PAP) PoliciesHandler {
	return &policiesHandler{logger: logger, cache: cache}
}

// GetPolicies implements the PoliciesHandler interface.
func (h *policiesHandler) GetPolicies(req *fiber.Ctx) error {
	list := h.cache.ListAllKeys()
	resp := make([]*policies.Policy, 0, len(list))

	for i := range list {
		key := list[i]
		if pol, err := h.cache.Get(key); err == nil {
			// we ignore policies that got deleted after we retrieved the list of keys.
			resp = append(resp, h.convertPolicy(pol))
		}
	}

	return req.JSON(&resp)
}

// GetPolicy implements the PoliciesHandler interface.
func (h *policiesHandler) GetPolicy(req *fiber.Ctx) error {
	id, ok, err := h.checkID(req)
	if !ok {
		return err
	}

	pol, err2 := h.cache.Get(id)
	if err2 != nil {
		return SendMessageResponse(req, fiber.StatusNotFound, err2.Error())
	}
	return req.JSON(h.convertPolicy(pol))
}

// PutPolicy implements the PoliciesHandler interface.
func (h *policiesHandler) PutPolicy(req *fiber.Ctx) error {
	id, ok, err := h.checkID(req)
	if !ok {
		return err
	}

	upsert := req.QueryBool("forceUpsert")

	var p *policies.Policy
	if p, ok, err = h.checkBody(req, id); !ok {
		return err
	}

	var pol pap.Policy
	if pol, ok, err = h.buildPolicy(req, p); !ok {
		return err
	}

	if upsert {
		// for upsert we check if the policy exists.
		// if it exists, we replace it, otherwise we add it.
		if _, err = h.cache.Get(p.Id); err == nil {
			pol2, err2 := h.cache.Replace(pol)
			if err2 != nil {
				return SendMessageResponse(req, fiber.StatusNotFound, err2.Error())
			}
			return req.JSON(h.convertPolicy(pol2))
		}
	}

	pol2, err2 := h.cache.Add(pol)
	if err2 != nil {
		return SendMessageResponse(req, fiber.StatusConflict, err2.Error())
	}
	return req.JSON(h.convertPolicy(pol2))
}

// PostPolicy implements the PoliciesHandler interface.
func (h *policiesHandler) PostPolicy(req *fiber.Ctx) error {
	id, ok, err := h.checkID(req)
	if !ok {
		return err
	}

	upsert := req.QueryBool("forceUpsert")

	var p *policies.Policy
	if p, ok, err = h.checkBody(req, id); !ok {
		return err
	}

	var pol pap.Policy
	if pol, ok, err = h.buildPolicy(req, p); !ok {
		return err
	}

	if upsert {
		// for upsert we check if the policy exists.
		// if it doesn't exist, we add it, otherwise we replace it.
		if _, err = h.cache.Get(p.Id); err != nil {
			pol2, err2 := h.cache.Add(pol)
			if err2 != nil {
				return SendMessageResponse(req, fiber.StatusConflict, err2.Error())
			}
			return req.JSON(h.convertPolicy(pol2))
		}
	}

	pol2, err2 := h.cache.Replace(pol)
	if err2 != nil {
		return SendMessageResponse(req, fiber.StatusNotFound, err2.Error())
	}
	return req.JSON(h.convertPolicy(pol2))
}

// DeletePolicy implements the PoliciesHandler interface.
func (h *policiesHandler) DeletePolicy(req *fiber.Ctx) error {
	id, ok, err := h.checkID(req)
	if !ok {
		return err
	}

	ignore := req.QueryBool("ignoreMissing")

	if _, err = h.cache.Get(id); err != nil {
		if ignore {
			return req.JSON(&policies.Policy{Id: id})
		}
		return SendMessageResponse(req, fiber.StatusNotFound, err.Error())
	}

	pol, err2 := h.cache.Remove(id)
	if err2 != nil {
		return SendMessageResponse(req, fiber.StatusNotFound, err2.Error())
	}
	return req.JSON(h.convertPolicy(pol))
}

func (h *policiesHandler) checkID(req *fiber.Ctx) (string, bool, error) {
	id := req.Params("id")
	if id == "" || len(id) > 500 {
		return "", false, SendMessageResponse(req, fiber.StatusBadRequest, "id must be filled or nit more than 500 characters")
	}
	return id, true, nil
}

func (h *policiesHandler) checkBody(req *fiber.Ctx, id string) (*policies.Policy, bool, error) {
	var p policies.Policy
	if err := req.BodyParser(&p); err != nil {
		return nil, false, SendMessageResponse(req, fiber.StatusBadRequest, err.Error())
	}

	if p.Id != id {
		if p.Id != "" {
			return nil, false, SendMessageResponse(req, fiber.StatusBadRequest, "mismatched policy id")
		}
		p.Id = id
	}

	return &p, true, nil
}

func (h *policiesHandler) buildPolicy(req *fiber.Ctx, p *policies.Policy) (pap.Policy, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req2, err := http.NewRequestWithContext(ctx, fiber.MethodGet, p.Url, nil)
	if err != nil {
		return nil, false, SendMessageResponse(req, fiber.StatusBadRequest, err.Error())
	}

	resp, err2 := http.DefaultClient.Do(req2)
	if err2 != nil {
		return nil, false, SendMessageResponse(req, fiber.StatusBadRequest, err2.Error())
	}

	defer resp.Body.Close()

	pol, err3 := pap.NewPolicy(p, resp.Body)
	if err3 != nil {
		return nil, false, SendMessageResponse(req, fiber.StatusBadRequest, err3.Error())
	}
	return pol, true, nil
}

func (h *policiesHandler) convertPolicy(pol pap.Policy) *policies.Policy {
	return &policies.Policy{
		Id:       pol.ID(),
		Language: pol.Language(),
		RvvaID:   pol.RvvaID(),
		Source:   pol.Source(),
		Target:   pol.Target(),
		Url:      pol.URI(),
	}
}

type policiesHandler struct {
	logger *slog.Logger
	cache  pap.PAP
}
