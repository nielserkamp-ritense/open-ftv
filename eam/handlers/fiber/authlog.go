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
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/identity"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/decisions"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/search"
	server "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/authlog"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/convert"
)

// ADLHandler represents the interface for handling search requests on an Authorisation Decision Log.
type ADLHandler interface {
	Search(req *fiber.Ctx) error // summary for a resource.
}

// NewADLHandler instantiates an authorisation log handler.
func NewADLHandler(logger *slog.Logger, s search.Searcher, authorizer authorization.Authorizer) ADLHandler {
	return &adlHandler{logger: logger, search: s, authorizer: authorizer}
}

// Search implements the ADLHandler interface.
func (h *adlHandler) Search(fc *fiber.Ctx) error {
	fc.Set(HeaderVersion, AttributesVersion)

	_, ok, err := h.authorize(fc)
	if !ok {
		return err
	}

	var requestTypes []decisions.AuthRequestType
	if requestTypes, err = h.getRequestTypes(fc); err != nil {
		return server.SendMessageResponse(fc, fiber.StatusBadRequest, err.Error())
	}

	var bundles []int64
	if bundles, err = h.getBundles(fc); err != nil {
		return server.SendMessageResponse(fc, fiber.StatusBadRequest, err.Error())
	}

	var recent time.Duration
	if s := convert.AnyToString(fc.Query("recent")); s != "" {
		if recent, err = time.ParseDuration(s); err != nil {
			return server.SendMessageResponse(fc, fiber.StatusBadRequest, err.Error())
		}
	}

	c := &search.Criteria{
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
	}

	var resp authlog.AuthlogEntries
	if resp, err = h.search.Search(fc.UserContext(), c); err != nil {
		if errors.As(err, &pErr) {
			return server.SendMessageResponse(fc, fiber.StatusBadRequest, err.Error())
		}
		return server.SendMessageResponse(fc, fiber.StatusInternalServerError, err.Error())
	}

	return fc.JSON(resp)
}

var pErr = &search.ParameterError{}

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

func (h *adlHandler) authorize(req *fiber.Ctx) (identity.Principal, bool, error) {
	return authorizeRequest(h.authorizer, req, h.logger)
}

type adlHandler struct {
	logger     *slog.Logger
	search     search.Searcher
	authorizer authorization.Authorizer
}
