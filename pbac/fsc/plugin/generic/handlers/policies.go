package handlers

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/oas/policies"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/components/pdp"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/fsc/plugin/generic/config"
)

// PoliciesHandler represents the interface for handling requests about policies.
type PoliciesHandler interface {
	Policies(req *fiber.Ctx) error
}

// NewPoliciesHandler instantiates a policy handler.
func NewPoliciesHandler(cfg *config.Config, logger *slog.Logger, c pdp.Controller) PoliciesHandler {
	return &policiesHandler{cfg: cfg, logger: logger, c: c}
}

// Policies implements the PoliciesHandler interface.
func (h *policiesHandler) Policies(req *fiber.Ctx) error {
	list := h.c.PAP().ListAllKeys()
	resp := make(policies.PoliciesResponse, 0, len(list))

	for i := range list {
		key := list[i]
		if _, err := h.c.PAP().Get(key); err == nil {
			key = strings.TrimSuffix(key, filepath.Ext(key))
			p := policies.Policy{Id: key, Language: h.c.Name()}

			if e := h.c.PIP().GetEntity(fmt.Sprintf("doelbinding::%s", key)); e != nil {
				if s, ok := e.Attributes().GetAttribute("rvvaID").(string); ok {
					p.RvvaID = &s
				}
				if s, ok := e.Attributes().GetAttribute("source").(string); ok {
					p.Source = &s
				}
				if s, ok := e.Attributes().GetAttribute("target").(string); ok {
					p.Target = &s
				}
				if s, ok := e.Attributes().GetAttribute("url").(string); ok {
					p.Url = s
				}
			}

			resp = append(resp, p)
		}
	}

	return req.JSON(&resp)
}

type policiesHandler struct {
	cfg    *config.Config
	logger *slog.Logger
	c      pdp.Controller
}
