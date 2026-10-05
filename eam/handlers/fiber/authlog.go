package fiber

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	auth "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/decisions"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/search"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/management"
	server "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/convert"
)

// ADLHandler represents the interface for handling search requests on an Authorisation Decision Log.
type ADLHandler interface {
	Search(req *fiber.Ctx) error // summary for a resource.
}

// NewADLHandler instantiates an authorisation log handler. Every search is decided by the
// decision service through the authorizer; a nil authorizer permits everything.
func NewADLHandler(logger *slog.Logger, s search.Searcher, authorizer authorization.RequestAuthorizer) ADLHandler {
	return &adlHandler{logger: logger, svc: management.NewDecisionService(s, authorizer), secured: authorizer != nil}
}

// Search implements the ADLHandler interface.
func (h *adlHandler) Search(fc *fiber.Ctx) error {
	fc.Set(HeaderVersion, AttributesVersion)

	caller := callerOf(fc, h.secured, h.logger)

	c, err := h.criteria(fc)
	if err != nil {
		// malformed criteria are still answered in decision order: the denial comes first.
		if derr := h.svc.DecideSearch(fc.UserContext(), caller); derr != nil {
			return h.fail(fc, derr)
		}

		h.logger.Warn("invalid search criteria", "path", fc.Path(), "err", err)

		return server.SendMessageResponse(fc, fiber.StatusBadRequest, err.Error())
	}

	resp, err := h.svc.Search(fc.UserContext(), caller, c)
	if err != nil {
		return h.fail(fc, err)
	}

	return fc.JSON(resp)
}

// criteria reads the search criteria from the query parameters.
func (h *adlHandler) criteria(fc *fiber.Ctx) (*search.Criteria, error) {
	requestTypes, err := h.getRequestTypes(fc)
	if err != nil {
		return nil, err
	}

	bundles, err := h.getBundles(fc)
	if err != nil {
		return nil, err
	}

	var recent time.Duration
	if s := convert.AnyToString(fc.Query("recent")); s != "" {
		if recent, err = time.ParseDuration(s); err != nil {
			return nil, err
		}
	}

	return &search.Criteria{
		ID:           int64(fc.QueryInt("id")),
		From:         convert.AnyToDateTime(fc.Query("start")),
		To:           convert.AnyToDateTime(fc.Query("end")),
		Recent:       recent,
		RequestTypes: requestTypes,
		Bundles:      bundles,
		TraceId:      fc.Query("traceId"),
		SpanId:       fc.Query("spanId"),
		SubjectType:  fc.Query("subjectType"),
		SubjectId:    fc.Query("subjectId"),
		ActionName:   fc.Query("actionName"),
		ResourceType: fc.Query("resourceType"),
		ResourceId:   fc.Query("resourceId"),
		Limit:        fc.QueryInt("limit"),
	}, nil
}

// fail answers a failed service call: a denial is 403, criteria the search rejects 400, and
// anything else 500 without the cause.
func (h *adlHandler) fail(fc *fiber.Ctx, err error) error {
	var invalid *search.ParameterError

	switch {
	case errors.Is(err, authorization.ErrForbidden):
		return auth.Forbidden(fc, err, h.logger)
	case errors.As(err, &invalid):
		h.logger.Warn("invalid search criteria", "path", fc.Path(), "err", err)
		return server.SendMessageResponse(fc, fiber.StatusBadRequest, err.Error())
	default:
		return internalError(fc, h.logger, err)
	}
}

func (h *adlHandler) getRequestTypes(fc *fiber.Ctx) ([]decisions.AuthRequestType, error) {
	s := fc.Query("types")
	if s == "" {
		return nil, nil
	}

	list := strings.Split(s, ",")
	out := make([]decisions.AuthRequestType, len(list))

	var rt decisions.AuthRequestType
	for i := range list {
		if rt.UnmarshalJSON([]byte(`"`+list[i]+`"`)) != nil {
			return nil, fmt.Errorf("invalid request type: %s", list[i])
		}
		out[i] = rt
	}

	return out, nil
}

func (h *adlHandler) getBundles(fc *fiber.Ctx) ([]int64, error) {
	s := fc.Query("policies")
	if s == "" {
		return nil, nil
	}

	list := strings.Split(s, ",")
	out := make([]int64, len(list))

	for i := range list {
		if v, err := strconv.ParseInt(list[i], 10, 64); err != nil {
			return nil, fmt.Errorf("invalid policy bundle: %s", list[i])
		} else {
			out[i] = v
		}
	}

	return out, nil
}

type adlHandler struct {
	logger  *slog.Logger
	svc     *management.DecisionService
	secured bool // an authorizer decides; the routes must carry the Identify middleware.
}
