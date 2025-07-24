package fiber

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	authRequest "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	server "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/policies"
)

// PoliciesVersion is the full semantic API version for the policy endpoints.
const PoliciesVersion = "1.3.0" // check against oas/policies/openapi.yaml!

// PoliciesHandler represents the interface for handling requests about policies.
type PoliciesHandler interface {
	GetPolicies(req *fiber.Ctx) error
	GetPolicy(req *fiber.Ctx) error
	PostPolicy(req *fiber.Ctx) error
	PutPolicy(req *fiber.Ctx) error
	DeletePolicy(req *fiber.Ctx) error
}

// NewPoliciesHandler instantiates a policy handler.
func NewPoliciesHandler(logger *slog.Logger, cache *pap.PAP, authorizer authorization.Authorizer) PoliciesHandler {
	return &policiesHandler{logger: logger, cache: cache, authorizer: authorizer}
}

// GetPolicies implements the PoliciesHandler interface.
func (h *policiesHandler) GetPolicies(req *fiber.Ctx) error {
	// TODO: log request/response to audit log.
	req.Set(HeaderVersion, PoliciesVersion)

	if ok, err := h.authorize(req); !ok || err != nil {
		return err
	}

	list, err := h.cache.List("")
	if err != nil {
		return server.SendMessageResponse(req, fiber.StatusInternalServerError, err.Error())
	}

	list2 := make([]*policies.Policy, len(list))
	for i := range list {
		list2[i] = h.convertPolicy(list[i], false)
	}
	return req.JSON(list2)
}

// GetPolicy implements the PoliciesHandler interface.
func (h *policiesHandler) GetPolicy(req *fiber.Ctx) error {
	// TODO: log request/response to audit log.
	req.Set(HeaderVersion, PoliciesVersion)

	if ok, err := h.authorize(req); !ok || err != nil {
		return err
	}

	language, id, ok, err := h.checkKey(req)
	if !ok {
		return err
	}

	pol, _, err2 := h.cache.Read(language, id)
	if err2 != nil {
		return server.SendMessageResponse(req, fiber.StatusNotFound, err2.Error())
	}
	return req.JSON(h.convertPolicy(pol, true))
}

// PostPolicy implements the PoliciesHandler interface.
func (h *policiesHandler) PostPolicy(req *fiber.Ctx) error {
	// TODO: log request/response to audit log.
	req.Set(HeaderVersion, PoliciesVersion)

	if ok, err := h.authorize(req); !ok || err != nil {
		return err
	}

	language, id, ok, err := h.checkKey(req)
	if !ok {
		return err
	}

	upsert := req.QueryBool("forceUpsert")

	var p *policies.Policy
	if p, ok, err = h.checkBody(req, language, id); !ok {
		return err
	}

	var pol *models.Policy
	if pol, ok, err = h.buildPolicy(req, p); !ok {
		return err
	}

	if upsert {
		// for upsert we check if the policy exists.
		// if it exists, we replace it, otherwise we add it.
		if prev, lastIndex, err2 := h.cache.Read(p.Language, p.Id); err2 == nil && prev != nil {
			pol2, err3 := h.cache.Update(prev, lastIndex, pol)
			if err3 != nil {
				return server.SendMessageResponse(req, fiber.StatusNotFound, err3.Error())
			}
			return req.JSON(h.convertPolicy(pol2, true))
		}
	}

	pol2, err2 := h.cache.Create(pol)
	if err2 != nil {
		return server.SendMessageResponse(req, fiber.StatusConflict, err2.Error())
	}

	return req.Status(fiber.StatusCreated).JSON(h.convertPolicy(pol2, true))
}

// PutPolicy implements the PoliciesHandler interface.
func (h *policiesHandler) PutPolicy(req *fiber.Ctx) error {
	// TODO: log request/response to audit log.
	req.Set(HeaderVersion, PoliciesVersion)

	if ok, err := h.authorize(req); !ok || err != nil {
		return err
	}

	language, id, ok, err := h.checkKey(req)
	if !ok {
		return err
	}

	upsert := req.QueryBool("forceUpsert")

	var p *policies.Policy
	if p, ok, err = h.checkBody(req, language, id); !ok {
		return err
	}

	var pol *models.Policy
	if pol, ok, err = h.buildPolicy(req, p); !ok {
		return err
	}

	prev, lastIndex, err2 := h.cache.Read(p.Language, p.Id)
	if err2 != nil || prev == nil {
		switch {
		case upsert:
			// for upsert we check if the policy exists.
			// if it doesn't exist, we add it, otherwise we replace it.
			pol2, err3 := h.cache.Create(pol)
			if err3 != nil {
				return server.SendMessageResponse(req, fiber.StatusConflict, err3.Error())
			}
			return req.JSON(h.convertPolicy(pol2, true))

		case err2 != nil:
			return server.SendMessageResponse(req, fiber.StatusNotFound, err2.Error())

		default:
			return server.SendMessageResponse(req, fiber.StatusNotFound, "not found")
		}
	}

	pol2, err3 := h.cache.Update(prev, lastIndex, pol)
	if err3 != nil {
		return server.SendMessageResponse(req, fiber.StatusConflict, err3.Error())
	}
	return req.JSON(h.convertPolicy(pol2, true))
}

// DeletePolicy implements the PoliciesHandler interface.
func (h *policiesHandler) DeletePolicy(req *fiber.Ctx) error {
	// TODO: log request/response to audit log.
	req.Set(HeaderVersion, PoliciesVersion)

	if ok, err := h.authorize(req); !ok || err != nil {
		return err
	}

	language, id, ok, err := h.checkKey(req)
	if !ok {
		return err
	}

	ignore := req.QueryBool("ignoreMissing")

	prev, lastIndex, err2 := h.cache.Read(language, id)
	if err2 != nil || prev == nil {
		switch {
		case ignore:
			return req.JSON(&policies.Policy{Language: language, Id: id})
		case err2 != nil:
			return server.SendMessageResponse(req, fiber.StatusNotFound, err2.Error())
		default:
			return server.SendMessageResponse(req, fiber.StatusNotFound, "not found")
		}
	}

	pol, err3 := h.cache.Delete(prev, lastIndex)
	if err3 != nil {
		return server.SendMessageResponse(req, fiber.StatusNotFound, err3.Error())
	}
	return req.JSON(h.convertPolicy(pol, true))
}

func (h *policiesHandler) checkKey(req *fiber.Ctx) (string, string, bool, error) {
	language := req.Params("language")
	if language == "" || len(language) > 100 {
		return "", "", false, server.SendMessageResponse(req, fiber.StatusBadRequest, "language must be filled and not more than 100 characters")
	}

	id := req.Params("id")
	if id == "" || len(id) > 500 {
		return "", "", false, server.SendMessageResponse(req, fiber.StatusBadRequest, "id must be filled and not more than 500 characters")
	}

	return language, id, true, nil
}

func (h *policiesHandler) checkBody(req *fiber.Ctx, language, id string) (*policies.Policy, bool, error) {
	var p policies.Policy
	if err := req.BodyParser(&p); err != nil {
		return nil, false, server.SendMessageResponse(req, fiber.StatusBadRequest, err.Error())
	}

	if p.Language != language {
		if p.Language != "" {
			return nil, false, server.SendMessageResponse(req, fiber.StatusBadRequest, "mismatched policy language")
		}
		p.Language = language
	}

	if p.Id != id {
		if p.Id != "" {
			return nil, false, server.SendMessageResponse(req, fiber.StatusBadRequest, "mismatched policy id")
		}
		p.Id = id
	}

	return &p, true, nil
}

func (h *policiesHandler) buildPolicy(req *fiber.Ctx, p *policies.Policy) (*models.Policy, bool, error) {
	if p.Url == "" {
		if p.Data == "" {
			return nil, false, server.SendMessageResponse(req, fiber.StatusBadRequest, "policy data or url required")
		}

		pol, err3 := models.NewPolicy(p, bytes.NewBufferString(p.Data))
		if err3 != nil {
			return nil, false, server.SendMessageResponse(req, fiber.StatusBadRequest, err3.Error())
		}
		return pol, true, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req2, err := http.NewRequestWithContext(ctx, fiber.MethodGet, p.Url, nil)
	if err != nil {
		return nil, false, server.SendMessageResponse(req, fiber.StatusBadRequest, err.Error())
	}

	resp, err2 := http.DefaultClient.Do(req2)
	if err2 != nil {
		return nil, false, server.SendMessageResponse(req, fiber.StatusBadRequest, err2.Error())
	}

	defer resp.Body.Close()

	pol, err3 := models.NewPolicy(p, resp.Body)
	if err3 != nil {
		return nil, false, server.SendMessageResponse(req, fiber.StatusBadRequest, err3.Error())
	}
	return pol, true, nil
}

func (h *policiesHandler) convertPolicy(pol *models.Policy, withData bool) *policies.Policy {
	if !withData || pol.URI() != "" {
		return &policies.Policy{
			Id:       pol.ID(),
			Language: pol.Language(),
			RvvaId:   pol.RvvaID(),
			Url:      pol.URI(),
		}
	}

	data, _ := io.ReadAll(pol.Content())
	return &policies.Policy{
		Id:       pol.ID(),
		Language: pol.Language(),
		RvvaId:   pol.RvvaID(),
		Data:     string(data),
	}
}

func (h *policiesHandler) authorize(req *fiber.Ctx) (bool, error) {
	if h.authorizer == nil {
		return true, nil
	}

	resp, err := h.authorizer.Authorize(authRequest.FormatRequest(req))

	// TODO: log authorization decision to auth-decision log.

	return authRequest.Check(req, resp, err)
}

type policiesHandler struct {
	logger     *slog.Logger
	cache      *pap.PAP
	authorizer authorization.Authorizer
}
