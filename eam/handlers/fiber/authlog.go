package fiber

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"

	"github.com/gofiber/fiber/v2"

	fiber2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/server/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/oas/authlog"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities-no-ci/opensearch"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/convert"
)

// AuthlogHandler represents the interface for handling requests for the authorisation log.
type AuthlogHandler interface {
	GetAuthlogResource(req *fiber.Ctx) error // summary for a resource.
}

// NewAuthlogHandler instantiates an authorisation log handler.
func NewAuthlogHandler(logger *slog.Logger, index string, os opensearch.Searcher) AuthlogHandler {
	return &authlogHandler{logger: logger, index: index, os: os}
}

// GetAuthlogResource implements the AuthlogHandler interface.
func (h *authlogHandler) GetAuthlogResource(req *fiber.Ctx) error {
	var resp authlog.AuthlogResourceResponse

	resource := req.Params("resource")
	unescaped, err := url.QueryUnescape(resource)
	if err != nil {
		return fiber2.SendMessageResponse(req, fiber.StatusBadRequest, "invalid resource id")
	}

	q := fmt.Sprintf(`{"size":0,"query":{"term":{"resource.id.keyword":{"value":"%s"}}},"aggs":{"principal.id":{"terms":{"field":"principal.id.keyword","size":999999}}}}`, unescaped)
	m, err2 := h.os.SearchBySQL(context.Background(), h.index, q, 0)

	if err2 != nil {
		h.logger.Error("failed to query index", "index", h.index, "q", q, "err", err2)
		return fiber2.SendMessageResponse(req, fiber.StatusInternalServerError, "failed to query authlog")
	}

	if m2, ok := m.Aggregations["principal.id"].(map[string]any); ok {
		if list, ok2 := m2["buckets"].([]any); ok2 {
			for i := range list {
				if m3, ok3 := list[i].(map[string]any); ok3 {
					rvvaId := convert.AnyToString(m3["key"])
					count := int(convert.AnyToInt64(m3["doc_count"]))

					resp.Total += count

					resp.Rvva = append(resp.Rvva, struct {
						Count  int    `json:"count,omitempty"`
						RvvaId string `json:"rvvaId,omitempty"`
					}{
						Count:  count,
						RvvaId: rvvaId,
					})
				}
			}
		}
	}

	if len(resp.Rvva) == 0 {
		return fiber2.SendMessageResponse(req, fiber.StatusNotFound, authlogResourceNotFound)
	}
	return req.JSON(resp)
}

type authlogHandler struct {
	logger *slog.Logger
	index  string
	os     opensearch.Searcher
}

const (
	authlogResourceNotFound = "authlog resource not found"
)
