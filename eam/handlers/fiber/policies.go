package fiber

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"
	"unicode/utf8"

	"github.com/gofiber/fiber/v2"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	auth "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/identity"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	server "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server/fiber"
	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/policies"
)

// PoliciesVersion is the full semantic API version for the policy endpoints.
const PoliciesVersion = "1.7.1" // check against oas/policies/openapi.yaml!

// Maximum lengths enforced by the policy field checks below.
const (
	maxLanguageLength  = 40
	maxRvvaIDLength    = 80
	maxPolicyURLLength = 400
	maxPolicyKeyLength = 40
)

// policyFetchTimeout bounds how long a policy uri may take to respond.
const policyFetchTimeout = 10 * time.Second

var (
	errPolNotFound     = errors.New("policy not found")
	errPolExists       = errors.New("policy already exists")
	errPolKeyError     = fmt.Errorf("policy id must be filled and less or equal %d characters", maxPolicyKeyLength)
	errPolVersionError = errors.New("policy version must be filled and positive")
	errPolUrlContent   = errors.New("policy data or url required")

	errLanguageRequired  = checkIssue{code: "E02005", msg: "language must be filled"}
	errLanguageTooLong   = checkIssue{code: "E02010", msg: fmt.Sprintf("language too long (max %d characters)", maxLanguageLength)}
	errRvvaIDTooLong     = checkIssue{code: "E02015", msg: fmt.Sprintf("RvVA identifier too long (max %d characters)", maxRvvaIDLength)}
	errPolicyDataMissing = checkIssue{code: "E02020", msg: "either uri or data must be filled"}
	errPolicyDataBothSet = checkIssue{code: "E02025", msg: "only one of uri and data can be filled"}
	errPolicyURLTooLong  = checkIssue{code: "E02030", msg: fmt.Sprintf("uri too long (max %d characters)", maxPolicyURLLength)}
	errPolicyIDMismatch  = checkIssue{code: "E02035", msg: "policy id mismatch"}
)

func (f *fieldChecker) checkLanguage(language string) *fieldChecker {
	switch {
	case language == "":
		f.addIssue(errLanguageRequired)
	case utf8.RuneCountInString(language) > maxLanguageLength:
		f.addIssue(errLanguageTooLong)
	}

	return f
}

func (f *fieldChecker) checkRvvaID(id string) *fieldChecker {
	if utf8.RuneCountInString(id) > maxRvvaIDLength {
		f.addIssue(errRvvaIDTooLong)
	}

	return f
}

func (f *fieldChecker) checkPolicyData(url, data string) *fieldChecker {
	switch {
	case url == "" && data == "":
		f.addIssue(errPolicyDataMissing)
	case url != "" && data != "":
		f.addIssue(errPolicyDataBothSet)
	case utf8.RuneCountInString(url) > maxPolicyURLLength:
		f.addIssue(errPolicyURLTooLong)
	}

	return f
}

// PoliciesHandler represents the interface for handling requests about policies.
type PoliciesHandler interface {
	GetPolicies(req *fiber.Ctx) error       // retrieve all policies.
	GetPolicy(req *fiber.Ctx) error         // retrieve a single policy.
	GetPolicyVersions(req *fiber.Ctx) error // retrieve all versions of a policy.
	GetPolicyVersion(req *fiber.Ctx) error  // retrieve a specific version of a policy.
	PostPolicy(req *fiber.Ctx) error        // create a new policy.
	PutPolicy(req *fiber.Ctx) error         // update an existing policy.
	PatchPolicyStatus(req *fiber.Ctx) error // update the status of an existing policy.
	PostPolicyRestore(req *fiber.Ctx) error // restore an old version of a policy.
	DeletePolicy(req *fiber.Ctx) error      // remove an existing policy.
}

type policiesHandler struct {
	logger     *slog.Logger
	cache      *pap.PAP
	authorizer authorization.Authorizer
}

// NewPoliciesHandler instantiates a policy handler.
func NewPoliciesHandler(logger *slog.Logger, cache *pap.PAP, authorizer authorization.Authorizer) PoliciesHandler {
	return &policiesHandler{logger: logger, cache: cache, authorizer: authorizer}
}

// GetPolicies implements the PoliciesHandler interface.
func (h *policiesHandler) GetPolicies(req *fiber.Ctx) error {
	req.Set(HeaderVersion, PoliciesVersion)

	if _, err := h.authorize(req); err != nil {
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

	if _, err := h.authorize(req); err != nil {
		return err
	}

	id, err := h.checkKey(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	pol, _, err2 := h.cache.Read(id)
	if err2 != nil {
		return h.error(req, fiber.StatusInternalServerError, err2)
	}

	if pol == nil {
		return h.error(req, fiber.StatusNotFound, errPolNotFound)
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

// GetPolicyVersions implements the PoliciesHandler interface.
func (h *policiesHandler) GetPolicyVersions(req *fiber.Ctx) error {
	req.Set(HeaderVersion, PoliciesVersion)

	if _, err := h.authorize(req); err != nil {
		return err
	}

	id, err := h.checkKey(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	list, err2 := h.cache.ReadVersions(id)
	if err2 != nil {
		return h.error(req, fiber.StatusInternalServerError, err2)
	}

	return req.JSON(list)
}

// GetPolicyVersion implements the PoliciesHandler interface.
func (h *policiesHandler) GetPolicyVersion(req *fiber.Ctx) error {
	req.Set(HeaderVersion, PoliciesVersion)

	if _, err := h.authorize(req); err != nil {
		return err
	}

	id, err := h.checkKey(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	version, err := h.checkVersion(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	pol, err2 := h.cache.ReadVersion(id, version)
	if err2 != nil {
		return h.error(req, fiber.StatusInternalServerError, err2)
	}

	if pol == nil {
		return h.error(req, fiber.StatusNotFound, errPolNotFound)
	}

	return req.JSON(pol)
}

// PostPolicy implements the PoliciesHandler interface.
func (h *policiesHandler) PostPolicy(req *fiber.Ctx) error {
	req.Set(HeaderVersion, PoliciesVersion)

	user, err := h.authorize(req)
	if err != nil {
		return err
	}

	id, err := h.checkKey(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	p, code, err := h.checkBody(req, id)
	if err != nil {
		return h.badRequest(req, code, err)
	}

	p2, err := h.buildPolicy(p)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	prev, lastIndex, err2 := h.cache.Read(p.Id)
	switch {
	case err2 != nil:
		// no-op
	case prev != nil && !req.QueryBool("forceUpsert"):
		return h.error(req, fiber.StatusConflict, errPolExists)
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

	id, err := h.checkKey(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	// Load the stored target object BEFORE authorizing, so the PDP can
	// evaluate fine-grained policies against its attributes (e.g. status).
	prev, _, _ := h.cache.Read(id)

	user, err := h.authorizeResource(req, id, prev)
	if err != nil {
		return err
	}

	p, code, err := h.checkBody(req, id)
	if err != nil {
		return h.badRequest(req, code, err)
	}

	p2, err := h.buildPolicy(p)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	prev, lastIndex, err2 := h.cache.Read(p.Id)
	switch {
	case err2 != nil:
		// no-op
	case prev == nil && !req.QueryBool("forceUpsert"):
		return h.error(req, fiber.StatusNotFound, errPolNotFound)
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

// PatchPolicyStatus implements the PoliciesHandler interface.
func (h *policiesHandler) PatchPolicyStatus(req *fiber.Ctx) error {
	req.Set(HeaderVersion, PoliciesVersion)

	user, err := h.authorize(req)
	if err != nil {
		return err
	}

	id, err := h.checkKey(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	p, code, err := h.checkBodyStatus(req, id)
	if err != nil {
		return h.badRequest(req, code, err)
	}

	prev, lastIndex, err2 := h.cache.Read(p.Id)
	if err2 != nil {
		return h.error(req, fiber.StatusNotFound, err2)
	}

	p2, err3 := h.cache.UpdateStatus(prev, lastIndex, models.StatusFromString(p.Status), user)
	if err3 != nil {
		return h.error(req, fiber.StatusBadRequest, err3)
	}
	return req.JSON(p2.ToOAS(true))
}

// PostPolicyRestore implements the PoliciesHandler interface.
func (h *policiesHandler) PostPolicyRestore(req *fiber.Ctx) error {
	req.Set(HeaderVersion, PoliciesVersion)

	user, err := h.authorize(req)
	if err != nil {
		return err
	}

	id, err := h.checkKey(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	version, err := h.checkVersion(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	pol, err2 := h.cache.RestoreVersion(id, version, user)
	if err2 != nil {
		return h.error(req, fiber.StatusInternalServerError, err2)
	}

	return req.JSON(pol)
}

// DeletePolicy implements the PoliciesHandler interface.
func (h *policiesHandler) DeletePolicy(req *fiber.Ctx) error {
	req.Set(HeaderVersion, PoliciesVersion)

	user, err := h.authorize(req)
	if err != nil {
		return err
	}

	id, err := h.checkKey(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	var p2 *models.Policy

	prev, lastIndex, err2 := h.cache.Read(id)
	switch {
	case err2 != nil:
		// no-op
	case prev == nil && !req.QueryBool("ignoreMissing"):
		return h.error(req, fiber.StatusNotFound, errPolNotFound)
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

func (h *policiesHandler) checkKey(req *fiber.Ctx) (string, error) {
	id := req.Params("id")
	if id == "" || len(id) > maxPolicyKeyLength {
		return "", errPolKeyError
	}

	return id, nil
}

func (h *policiesHandler) checkVersion(req *fiber.Ctx) (int, error) {
	v, err := req.ParamsInt("version")
	if err != nil {
		return 0, err
	}

	if v <= 0 {
		return 0, errPolVersionError
	}

	return v, nil
}

// checkBody parses and validates the request body, returning the Code and error to report via
// badRequest on failure. A nil error means the returned Policy is valid.
func (h *policiesHandler) checkBody(req *fiber.Ctx, id string) (*oas.Policy, string, error) {
	var p oas.Policy
	if err := req.BodyParser(&p); err != nil {
		return nil, codeBadRequest, err
	}

	chk := newFieldChecker().
		checkIdentifiers(id, &p.Id, errPolicyIDMismatch).
		checkLanguage(p.Language).
		checkStatus(p.Status).
		checkTitle(p.Metadata.Title).
		checkRvvaID(p.Metadata.RvvaId).
		checkPolicyData(p.Metadata.Url, p.Data).
		checkTags(p.Metadata.Tags)

	if chk.checkFailed() {
		return nil, chk.firstCode(), chk.error()
	}

	return &p, "", nil
}

// checkBodyStatus is like checkBody but for the status-only request body.
func (h *policiesHandler) checkBodyStatus(req *fiber.Ctx, id string) (*oas.PolicyStatus, string, error) {
	var p oas.PolicyStatus
	if err := req.BodyParser(&p); err != nil {
		return nil, codeBadRequest, err
	}

	chk := newFieldChecker().
		checkIdentifiers(id, &p.Id, errPolicyIDMismatch).
		checkStatus(p.Status)

	if chk.checkFailed() {
		return nil, chk.firstCode(), chk.error()
	}

	return &p, "", nil
}

// buildPolicy resolves a policy from either its inline data or the uri it points at.
func (h *policiesHandler) buildPolicy(p *oas.Policy) (*models.Policy, error) {
	if p.Metadata.Url == "" {
		if p.Data == "" {
			return nil, errPolUrlContent
		}

		return models.NewPolicyFromOAS(p, bytes.NewBufferString(p.Data))
	}

	ctx, cancel := context.WithTimeout(context.Background(), policyFetchTimeout)
	defer cancel()

	fetch, err := http.NewRequestWithContext(ctx, fiber.MethodGet, p.Metadata.Url, http.NoBody)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(fetch)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	return models.NewPolicyFromOAS(p, resp.Body)
}

func (h *policiesHandler) authorize(req *fiber.Ctx) (identity.Principal, error) {
	return authorizeRequest(h.authorizer, req, h.logger)
}

// authorizeResource authorizes a request like authorize, but additionally
// passes the stored target object (with its attributes, e.g. status) into
// authorization so the PDP can evaluate fine-grained, resource-attribute
// policies. When prev is nil (no stored object), no resource is attached.
func (h *policiesHandler) authorizeResource(req *fiber.Ctx, id string, prev *models.Policy) (identity.Principal, error) {
	if h.authorizer == nil {
		return identity.NewSystemPrincipal(), nil
	}

	var res *models.Entity
	if prev != nil {
		attrs := models.NewAttributeSet()
		attrs.AddAttributeKV("status", prev.StatusName())
		res = models.NewEntity(models.EntityTypeService, id, attrs)
	}

	resp, principal, err := h.authorizer.Authorize(auth.FormatRequestWithResource(req, res))

	return auth.Check(req, resp, principal, err, h.logger)
}

func (h *policiesHandler) error(req *fiber.Ctx, status int, err error) error {
	h.logger.Error("request error", "path", req.Path(), "err", err, "status", status)
	return server.SendMessageResponse(req, status, err.Error())
}

// badRequest logs a warning and returns a 400 with the given validation error's code and message.
func (h *policiesHandler) badRequest(req *fiber.Ctx, code string, err error) error {
	h.logger.Warn("policy request rejected", "path", req.Path(), "err", err, "status", fiber.StatusBadRequest)
	return server.SendProblemResponse(req, fiber.StatusBadRequest, code, err.Error())
}
