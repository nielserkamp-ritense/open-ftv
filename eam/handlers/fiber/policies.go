package fiber

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	auth "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	server "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server/fiber"
	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/policies"
)

// PoliciesVersion is the full semantic API version for the policy endpoints.
const PoliciesVersion = "1.6.1" // check against oas/policies/openapi.yaml!

// PoliciesHandler represents the interface for handling requests about policies.
type PoliciesHandler interface {
	GetPolicies(req *fiber.Ctx) error  // retrieve all policies.
	GetPolicy(req *fiber.Ctx) error    // retrieve a single policy.
	PostPolicy(req *fiber.Ctx) error   // create a new policy.
	PutPolicy(req *fiber.Ctx) error    // update an existing policy.
	DeletePolicy(req *fiber.Ctx) error // remove an existing policy.
}

// NewPoliciesHandler instantiates a policy handler.
func NewPoliciesHandler(logger *slog.Logger, cache *pap.PAP, authorizer authorization.Authorizer) PoliciesHandler {
	return &policiesHandler{logger: logger, cache: cache, authorizer: authorizer}
}

// GetPolicies implements the PoliciesHandler interface.
func (h *policiesHandler) GetPolicies(req *fiber.Ctx) error {
	req.Set(HeaderVersion, PoliciesVersion)

	_, ok, err := h.authorize(req)
	if !ok {
		return err
	}

	list, err2 := h.cache.List("")
	if err2 != nil {
		return h.error(req, fiber.StatusInternalServerError, err2)
	}

	list2 := make([]*oas.Policy, len(list))
	for i := range list {
		list2[i] = list[i].ToOAS(false)
	}
	return req.JSON(list2)
}

// GetPolicy implements the PoliciesHandler interface.
func (h *policiesHandler) GetPolicy(req *fiber.Ctx) error {
	req.Set(HeaderVersion, PoliciesVersion)

	_, ok, err := h.authorize(req)
	if !ok {
		return err
	}

	var id string
	if id, ok, err = h.checkKey(req); !ok {
		return err
	}

	pol, _, err2 := h.cache.Read(id)
	if err2 != nil {
		return h.error(req, fiber.StatusInternalServerError, err2)
	}

	if pol == nil {
		return h.error(req, fiber.StatusNotFound, polNotFound)
	}

	out := pol.ToOAS(true)

	if audit, err3 := h.cache.ReadAudit(id); err3 != nil {
		h.logger.Warn("failed to read policy audit", "id", id, "err", err3)
	} else {
		out.AuditLog = audit
	}

	if usage, err3 := h.cache.ReadDeployments(id); err3 != nil {
		h.logger.Warn("failed to read policy deployments", "id", id, "err", err3)
	} else {
		out.UsageData = usage
	}

	return req.JSON(out)
}

// PostPolicy implements the PoliciesHandler interface.
func (h *policiesHandler) PostPolicy(req *fiber.Ctx) error {
	req.Set(HeaderVersion, PoliciesVersion)

	user, ok, err := h.authorize(req)
	if !ok {
		return err
	}

	var id string
	if id, ok, err = h.checkKey(req); !ok {
		return err
	}

	var p *oas.Policy
	if p, ok, err = h.checkBody(req, id); !ok {
		return err
	}

	var p2 *models.Policy
	if p2, ok, err = h.buildPolicy(req, p); !ok {
		return err
	}

	prev, lastIndex, err2 := h.cache.Read(p.Id)
	switch {
	case err2 != nil:
		// no-op
	case prev != nil && !req.QueryBool("forceUpsert"):
		return h.error(req, fiber.StatusConflict, polExists)
	case prev != nil:
		p2, err2 = h.cache.Update(prev, lastIndex, p2, user)
	default:
		p2, err2 = h.cache.Create(p2, user)
	}

	if err2 != nil {
		return h.error(req, fiber.StatusInternalServerError, err2)
	}
	return req.Status(fiber.StatusCreated).JSON(p2.ToOAS(true))
}

// PutPolicy implements the PoliciesHandler interface.
func (h *policiesHandler) PutPolicy(req *fiber.Ctx) error {
	req.Set(HeaderVersion, PoliciesVersion)

	user, ok, err := h.authorize(req)
	if !ok {
		return err
	}

	var id string
	if id, ok, err = h.checkKey(req); !ok {
		return err
	}

	var p *oas.Policy
	if p, ok, err = h.checkBody(req, id); !ok {
		return err
	}

	var p2 *models.Policy
	if p2, ok, err = h.buildPolicy(req, p); !ok {
		return err
	}

	prev, lastIndex, err2 := h.cache.Read(p.Id)
	switch {
	case err2 != nil:
		// no-op
	case prev == nil && !req.QueryBool("forceUpsert"):
		return h.error(req, fiber.StatusNotFound, polNotFound)
	case prev == nil:
		p2, err2 = h.cache.Create(p2, user)
	default:
		p2, err2 = h.cache.Update(prev, lastIndex, p2, user)
	}

	if err2 != nil {
		return h.error(req, fiber.StatusInternalServerError, err2)
	}
	return req.JSON(p2.ToOAS(true))
}

// DeletePolicy implements the PoliciesHandler interface.
func (h *policiesHandler) DeletePolicy(req *fiber.Ctx) error {
	req.Set(HeaderVersion, PoliciesVersion)

	user, ok, err := h.authorize(req)
	if !ok {
		return err
	}

	_ = user

	var id string
	if id, ok, err = h.checkKey(req); !ok {
		return err
	}

	var p2 *models.Policy

	prev, lastIndex, err2 := h.cache.Read(id)
	switch {
	case err2 != nil:
		// no-op
	case prev == nil && !req.QueryBool("ignoreMissing"):
		return h.error(req, fiber.StatusNotFound, polNotFound)
	case prev == nil:
		p2, err2 = models.NewPolicyFromData(id, "", "", "", &bytes.Buffer{})
	default:
		p2, err2 = h.cache.Delete(prev, lastIndex, user)
	}

	if err2 != nil {
		return h.error(req, fiber.StatusInternalServerError, err2)
	}
	return req.JSON(p2.ToOAS(true))
}

func (h *policiesHandler) checkKey(req *fiber.Ctx) (string, bool, error) {
	id := req.Params("id")
	if id == "" || len(id) > 40 {
		return "", false, h.error(req, fiber.StatusBadRequest, polKeyError)
	}

	return id, true, nil
}

func (h *policiesHandler) checkBody(req *fiber.Ctx, id string) (*oas.Policy, bool, error) {
	var p oas.Policy
	if err := req.BodyParser(&p); err != nil {
		return nil, false, h.error(req, fiber.StatusBadRequest, err)
	}

	chk := newFieldChecker().
		checkIdentifiers(id, &p.Id, "policy id mismatch").
		checkLanguage(p.Language).
		checkTitle(p.Metadata.Title).
		checkRvvaID(p.Metadata.RvvaId).
		checkPolicyData(p.Metadata.Url, p.Data).
		checkTags(p.Metadata.Tags)

	if chk.checkFailed() {
		return nil, false, h.error(req, fiber.StatusBadRequest, chk.error())
	}
	return &p, true, nil
}

func (h *policiesHandler) buildPolicy(req *fiber.Ctx, p *oas.Policy) (*models.Policy, bool, error) {
	if p.Metadata.Url == "" {
		if p.Data == "" {
			return nil, false, h.error(req, fiber.StatusBadRequest, polUrlContent)
		}

		pol, err3 := models.NewPolicyFromOAS(p, bytes.NewBufferString(p.Data))
		if err3 != nil {
			return nil, false, h.error(req, fiber.StatusBadRequest, err3)
		}
		return pol, true, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req2, err := http.NewRequestWithContext(ctx, fiber.MethodGet, p.Metadata.Url, nil)
	if err != nil {
		return nil, false, h.error(req, fiber.StatusBadRequest, err)
	}

	resp, err2 := http.DefaultClient.Do(req2)
	if err2 != nil {
		return nil, false, h.error(req, fiber.StatusBadRequest, err2)
	}

	defer resp.Body.Close()

	pol, err3 := models.NewPolicyFromOAS(p, resp.Body)
	if err3 != nil {
		return nil, false, h.error(req, fiber.StatusBadRequest, err3)
	}
	return pol, true, nil
}

func (h *policiesHandler) authorize(req *fiber.Ctx) (string, bool, error) {
	if h.authorizer == nil {
		return auth.SystemUser, true, nil
	}

	resp, err := h.authorizer.Authorize(auth.FormatRequest(req))
	return auth.Check(req, resp, err, h.logger)
}

func (h *policiesHandler) error(req *fiber.Ctx, status int, err error) error {
	h.logger.Error("request error", "path", req.Path(), "err", err, "status", status)
	return server.SendMessageResponse(req, status, err.Error())
}

type policiesHandler struct {
	logger     *slog.Logger
	cache      *pap.PAP
	authorizer authorization.Authorizer
}

var (
	polNotFound   = errors.New("policy not found")
	polExists     = errors.New("policy already exists")
	polKeyError   = errors.New("policy id must be filled and less or equal 40 characters")
	polUrlContent = errors.New("policy data or url required")
)
