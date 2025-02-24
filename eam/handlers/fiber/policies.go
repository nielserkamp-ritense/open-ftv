package fiber

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pap"
	fiber2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/server/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/oas/policies"
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
	list, err := h.cache.List("")
	if err != nil {
		return fiber2.SendMessageResponse(req, fiber.StatusInternalServerError, err.Error())
	}

	if len(list) == 0 {
		return fiber2.SendMessageResponse(req, fiber.StatusNotFound, "no policies found")
	}

	list2 := make([]*policies.Policy, len(list))
	for i := range list {
		list2[i] = h.convertPolicy(list[i])
	}

	return req.JSON(list2)
}

// GetPolicy implements the PoliciesHandler interface.
func (h *policiesHandler) GetPolicy(req *fiber.Ctx) error {
	language, id, ok, err := h.checkKey(req)
	if !ok {
		return err
	}

	pol, _, err2 := h.cache.Read(language, id)
	if err2 != nil {
		return fiber2.SendMessageResponse(req, fiber.StatusNotFound, err2.Error())
	}
	return req.JSON(h.convertPolicy(pol))
}

// PutPolicy implements the PoliciesHandler interface.
func (h *policiesHandler) PutPolicy(req *fiber.Ctx) error {
	language, id, ok, err := h.checkKey(req)
	if !ok {
		return err
	}

	upsert := req.QueryBool("forceUpsert")

	var p *policies.Policy
	if p, ok, err = h.checkBody(req, language, id); !ok {
		return err
	}

	var pol pap.Policy
	if pol, ok, err = h.buildPolicy(req, p); !ok {
		return err
	}

	if upsert {
		// for upsert we check if the policy exists.
		// if it exists, we replace it, otherwise we add it.
		if prev, lastIndex, err2 := h.cache.Read(p.Language, p.Id); err2 == nil && prev != nil {
			pol2, err3 := h.cache.Update(prev, lastIndex, pol)
			if err3 != nil {
				return fiber2.SendMessageResponse(req, fiber.StatusNotFound, err3.Error())
			}
			return req.JSON(h.convertPolicy(pol2))
		}
	}

	pol2, err2 := h.cache.Create(pol)
	if err2 != nil {
		return fiber2.SendMessageResponse(req, fiber.StatusConflict, err2.Error())
	}
	return req.JSON(h.convertPolicy(pol2))
}

// PostPolicy implements the PoliciesHandler interface.
func (h *policiesHandler) PostPolicy(req *fiber.Ctx) error {
	language, id, ok, err := h.checkKey(req)
	if !ok {
		return err
	}

	upsert := req.QueryBool("forceUpsert")

	var p *policies.Policy
	if p, ok, err = h.checkBody(req, language, id); !ok {
		return err
	}

	var pol pap.Policy
	if pol, ok, err = h.buildPolicy(req, p); !ok {
		return err
	}

	prev, lastIndex, err2 := h.cache.Read(p.Language, p.Id)
	if err2 != nil || prev == nil {
		if upsert {
			// for upsert we check if the policy exists.
			// if it doesn't exist, we add it, otherwise we replace it.
			pol2, err3 := h.cache.Create(pol)
			if err3 != nil {
				return fiber2.SendMessageResponse(req, fiber.StatusConflict, err3.Error())
			}
			return req.JSON(h.convertPolicy(pol2))
		}
		return fiber2.SendMessageResponse(req, fiber.StatusNotFound, err2.Error())
	}

	pol2, err3 := h.cache.Update(prev, lastIndex, pol)
	if err3 != nil {
		return fiber2.SendMessageResponse(req, fiber.StatusNotFound, err3.Error())
	}
	return req.JSON(h.convertPolicy(pol2))
}

// DeletePolicy implements the PoliciesHandler interface.
func (h *policiesHandler) DeletePolicy(req *fiber.Ctx) error {
	language, id, ok, err := h.checkKey(req)
	if !ok {
		return err
	}

	ignore := req.QueryBool("ignoreMissing")

	prev, lastIndex, err2 := h.cache.Read(language, id)
	if err2 != nil || prev == nil {
		if ignore {
			return req.JSON(&policies.Policy{Language: language, Id: id})
		}
		return fiber2.SendMessageResponse(req, fiber.StatusNotFound, err2.Error())
	}

	pol, err3 := h.cache.Delete(prev, lastIndex)
	if err3 != nil {
		return fiber2.SendMessageResponse(req, fiber.StatusNotFound, err3.Error())
	}
	return req.JSON(h.convertPolicy(pol))
}

func (h *policiesHandler) checkKey(req *fiber.Ctx) (string, string, bool, error) {
	language := req.Params("language")
	if language == "" || len(language) > 100 {
		return "", "", false, fiber2.SendMessageResponse(req, fiber.StatusBadRequest, "language must be filled and not more than 100 characters")
	}

	id := req.Params("id")
	if id == "" || len(id) > 500 {
		return "", "", false, fiber2.SendMessageResponse(req, fiber.StatusBadRequest, "id must be filled and not more than 500 characters")
	}

	return language, id, true, nil
}

func (h *policiesHandler) checkBody(req *fiber.Ctx, language, id string) (*policies.Policy, bool, error) {
	var p policies.Policy
	if err := req.BodyParser(&p); err != nil {
		return nil, false, fiber2.SendMessageResponse(req, fiber.StatusBadRequest, err.Error())
	}

	if p.Language != language {
		if p.Language != "" {
			return nil, false, fiber2.SendMessageResponse(req, fiber.StatusBadRequest, "mismatched policy language")
		}
		p.Language = language
	}

	if p.Id != id {
		if p.Id != "" {
			return nil, false, fiber2.SendMessageResponse(req, fiber.StatusBadRequest, "mismatched policy id")
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
		return nil, false, fiber2.SendMessageResponse(req, fiber.StatusBadRequest, err.Error())
	}

	resp, err2 := http.DefaultClient.Do(req2)
	if err2 != nil {
		return nil, false, fiber2.SendMessageResponse(req, fiber.StatusBadRequest, err2.Error())
	}

	defer resp.Body.Close()

	pol, err3 := pap.NewPolicy(p, resp.Body)
	if err3 != nil {
		return nil, false, fiber2.SendMessageResponse(req, fiber.StatusBadRequest, err3.Error())
	}
	return pol, true, nil
}

func (h *policiesHandler) convertPolicy(pol pap.Policy) *policies.Policy {
	return &policies.Policy{
		Id:       pol.ID(),
		Language: pol.Language(),
		RvvaId:   pol.RvvaID(),
		Url:      pol.URI(),
	}
}

type policiesHandler struct {
	logger *slog.Logger
	cache  pap.PAP
}
