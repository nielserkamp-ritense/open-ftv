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
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/management"
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
	logger  *slog.Logger
	svc     *management.PolicyService
	secured bool // an authorizer decides; the routes must carry the Identify middleware.
	principalResolver
}

// NewPoliciesHandler instantiates a policy handler. Every operation is decided by the policy
// service through the authorizer; a nil authorizer permits everything.
func NewPoliciesHandler(logger *slog.Logger, cache *pap.PAP, authorizer authorization.Authorizer, opts ...HandlerOption) PoliciesHandler {
	return &policiesHandler{
		logger:            logger,
		svc:               management.NewPolicyService(logger, cache, authorizer),
		secured:           authorizer != nil,
		principalResolver: newPrincipalResolver(logger, opts),
	}
}

// caller is the request-scoped Principal this request runs as.
func (h *policiesHandler) caller(req *fiber.Ctx) *authorization.RequestPrincipal {
	return callerOf(req, h.secured, h.logger)
}

// GetPolicies implements the PoliciesHandler interface.
func (h *policiesHandler) GetPolicies(req *fiber.Ctx) error {
	req.Set(HeaderVersion, PoliciesVersion)

	list, err := h.svc.List(req.UserContext(), h.caller(req))
	if err != nil {
		return h.fail(req, err)
	}

	list2 := make([]*oas.Policy, len(list))
	for i := range list {
		list2[i] = list[i].ToOAS(false)
	}

	return h.respond(req, list2)
}

// GetPolicy implements the PoliciesHandler interface.
func (h *policiesHandler) GetPolicy(req *fiber.Ctx) error {
	req.Set(HeaderVersion, PoliciesVersion)

	id, err := h.checkKey(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	details, err := h.svc.Get(req.UserContext(), h.caller(req), id)
	if err != nil {
		return h.fail(req, err)
	}

	out := details.Policy.ToOAS(true)
	out.AuditLog = details.AuditLog
	out.UsageData = details.UsageData

	return h.respond(req, out)
}

// GetPolicyVersions implements the PoliciesHandler interface.
func (h *policiesHandler) GetPolicyVersions(req *fiber.Ctx) error {
	req.Set(HeaderVersion, PoliciesVersion)

	id, err := h.checkKey(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	list, err := h.svc.Versions(req.UserContext(), h.caller(req), id)
	if err != nil {
		return h.fail(req, err)
	}

	return h.respond(req, list)
}

// GetPolicyVersion implements the PoliciesHandler interface.
func (h *policiesHandler) GetPolicyVersion(req *fiber.Ctx) error {
	req.Set(HeaderVersion, PoliciesVersion)

	id, err := h.checkKey(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	version, err := h.checkVersion(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	pol, err := h.svc.Version(req.UserContext(), h.caller(req), id, version)
	if err != nil {
		return h.fail(req, err)
	}

	return h.respond(req, pol)
}

// PostPolicy implements the PoliciesHandler interface.
func (h *policiesHandler) PostPolicy(req *fiber.Ctx) error {
	req.Set(HeaderVersion, PoliciesVersion)

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

	p2, err = h.svc.Create(req.UserContext(), h.caller(req), p2, req.QueryBool("forceUpsert"))
	if err != nil {
		return h.fail(req, err)
	}

	return h.respond(req.Status(fiber.StatusCreated), p2.ToOAS(true))
}

// PutPolicy implements the PoliciesHandler interface.
func (h *policiesHandler) PutPolicy(req *fiber.Ctx) error {
	req.Set(HeaderVersion, PoliciesVersion)

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

	p2, err = h.svc.Update(req.UserContext(), h.caller(req), p2, req.QueryBool("forceUpsert"))
	if err != nil {
		return h.fail(req, err)
	}

	return h.respond(req, p2.ToOAS(true))
}

// PatchPolicyStatus implements the PoliciesHandler interface.
func (h *policiesHandler) PatchPolicyStatus(req *fiber.Ctx) error {
	req.Set(HeaderVersion, PoliciesVersion)

	id, err := h.checkKey(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	p, code, err := h.checkBodyStatus(req, id)
	if err != nil {
		return h.badRequest(req, code, err)
	}

	p2, err := h.svc.SetStatus(req.UserContext(), h.caller(req), id, models.StatusFromString(p.Status))
	if err != nil {
		return h.fail(req, err)
	}

	return h.respond(req, p2.ToOAS(true))
}

// PostPolicyRestore implements the PoliciesHandler interface.
func (h *policiesHandler) PostPolicyRestore(req *fiber.Ctx) error {
	req.Set(HeaderVersion, PoliciesVersion)

	id, err := h.checkKey(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	version, err := h.checkVersion(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	pol, err := h.svc.Restore(req.UserContext(), h.caller(req), id, version)
	if err != nil {
		return h.fail(req, err)
	}

	return h.respond(req, pol)
}

// DeletePolicy implements the PoliciesHandler interface.
func (h *policiesHandler) DeletePolicy(req *fiber.Ctx) error {
	req.Set(HeaderVersion, PoliciesVersion)

	id, err := h.checkKey(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	p2, err := h.svc.Delete(req.UserContext(), h.caller(req), id, req.QueryBool("ignoreMissing"))
	if err != nil {
		return h.fail(req, err)
	}

	return h.respond(req, p2.ToOAS(true))
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

// fail answers a failed service call: a denial is 403 regardless of whether the object
// exists, the life cycle's refusal 400, the store's own outcomes 404 or 409, and anything
// else (including a permitted caller that could not be recorded) 500.
func (h *policiesHandler) fail(req *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, authorization.ErrForbidden):
		return auth.Forbidden(req, err, h.logger)
	case errors.Is(err, management.ErrInvalidTransition):
		return h.error(req, fiber.StatusBadRequest, err)
	case errors.Is(err, management.ErrNotFound):
		return h.error(req, fiber.StatusNotFound, errPolNotFound)
	case errors.Is(err, management.ErrExists):
		return h.error(req, fiber.StatusConflict, errPolExists)
	default:
		return h.error(req, fiber.StatusInternalServerError, err)
	}
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
