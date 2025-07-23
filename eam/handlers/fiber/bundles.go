package fiber

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	authRequest "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/bundles"
	server "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server/fiber"
	bundles2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/bundles"
)

// BundlesVersion is the full semantic API version for the bundle endpoints.
const BundlesVersion = "1.0.0" // check against oas/bundles/openapi.yaml!

// BundlesHandler represents the interface for handling requests about bundles.
type BundlesHandler interface {
	GetStatuses(req *fiber.Ctx) error
	GetCompressTypes(req *fiber.Ctx) error
	GetConfigs(req *fiber.Ctx) error
	GetDeployments(req *fiber.Ctx) error
	GetDeployment(req *fiber.Ctx) error
	PostDeployment(req *fiber.Ctx) error
}

// NewBundlesHandler instantiates a bundle deployment handler (no push function!).
func NewBundlesHandler(logger *slog.Logger, manager *bundles.Manager, authorizer authorization.Authorizer) BundlesHandler {
	return &bundlesHandler{logger: logger, manager: manager, cfg: manager.Bundles(), authorizer: authorizer}
}

// NewBundlePushHandler instantiates a bundle push endpoint handler (only push function!).
func NewBundlePushHandler(logger *slog.Logger, authorizer authorization.Authorizer) BundlesHandler {
	return &bundlesHandler{logger: logger, authorizer: authorizer, push: true}
}

// GetStatuses implements the BundlesHandler interface.
func (h *bundlesHandler) GetStatuses(req *fiber.Ctx) error {
	// TODO: log request/response to audit log.
	req.Set(HeaderVersion, BundlesVersion)

	if h.push {
		return server.SendMessageResponse(req, fiber.StatusBadRequest, msgOnlyPushEndpoint)
	}

	if ok, err := h.authorize(req); !ok || err != nil {
		return err
	}

	resp := make(bundles2.Statuses, 0, bundles.StatusCount)
	for s := bundles.StatusMIN; s <= bundles.StatusMAX; s++ {
		resp = append(resp, bundles2.Status{Code: int(s), Name: s.String()})
	}
	return req.JSON(resp)
}

// GetCompressTypes implements the BundlesHandler interface.
func (h *bundlesHandler) GetCompressTypes(req *fiber.Ctx) error {
	// TODO: log request/response to audit log.
	req.Set(HeaderVersion, BundlesVersion)

	if h.push {
		return server.SendMessageResponse(req, fiber.StatusBadRequest, msgOnlyPushEndpoint)
	}

	if ok, err := h.authorize(req); !ok || err != nil {
		return err
	}

	resp := make(bundles2.CompressTypes, 0, bundles.CompressCount)
	for s := bundles.CompressMIN; s <= bundles.CompressMAX; s++ {
		resp = append(resp, bundles2.CompressType{Code: int(s), Name: s.String()})
	}
	return req.JSON(resp)
}

// GetConfigs implements the BundlesHandler interface.
func (h *bundlesHandler) GetConfigs(req *fiber.Ctx) error {
	// TODO: log request/response to audit log.
	req.Set(HeaderVersion, BundlesVersion)

	if h.push {
		return server.SendMessageResponse(req, fiber.StatusBadRequest, msgOnlyPushEndpoint)
	}

	if ok, err := h.authorize(req); !ok || err != nil {
		return err
	}

	resp := make(bundles2.BundleConfigs, len(h.cfg))

	for i := range h.cfg {
		cfg := h.cfg[i]
		cfg2 := bundles2.BundleConfig{
			Tag:      cfg.Tag,
			Version:  cfg.Version,
			Policies: cfg.Policies,
			Data:     cfg.Data,
			Targets:  make([]bundles2.Target, 0, len(cfg.Targets)),
		}

		for j := range cfg.Targets {
			target := cfg.Targets[j]
			cfg2.Targets = append(cfg2.Targets, bundles2.Target{
				Uri:         target.URI,
				Tls:         target.CA != "" || target.Cert != "" || target.Key != "",
				Apikey:      target.APIKey != "",
				Compression: int(bundles.CompressionTypeFromString(target.Compress)),
				Headers:     target.Headers,
			})
		}

		resp[i] = cfg2
	}

	return req.JSON(resp)
}

// GetDeployments implements the BundlesHandler interface.
func (h *bundlesHandler) GetDeployments(req *fiber.Ctx) error {
	// TODO: log request/response to audit log.
	req.Set(HeaderVersion, BundlesVersion)

	if h.push {
		return server.SendMessageResponse(req, fiber.StatusBadRequest, msgOnlyPushEndpoint)
	}

	if ok, err := h.authorize(req); !ok || err != nil {
		return err
	}

	resp := make(bundles2.Deployments, 0, 8)

	// TODO: retrieve deployments

	return req.JSON(resp)
}

// GetDeployment implements the BundlesHandler interface.
func (h *bundlesHandler) GetDeployment(req *fiber.Ctx) error {
	// TODO: log request/response to audit log.
	req.Set(HeaderVersion, BundlesVersion)

	if h.push {
		return server.SendMessageResponse(req, fiber.StatusBadRequest, msgOnlyPushEndpoint)
	}

	if ok, err := h.authorize(req); !ok || err != nil {
		return err
	}

	// TODO: start deployment

	resp := &bundles2.Deployment{
		Version:     1,
		Description: "new deployment #1",
		Status:      int(bundles.Creating),
		Created:     time.Now().UTC().Format(time.RFC3339),
		Updated:     time.Now().UTC().Format(time.RFC3339),
	}

	return req.JSON(resp)
}

// PostDeployment implements the BundlesHandler interface.
func (h *bundlesHandler) PostDeployment(req *fiber.Ctx) error {
	// TODO: log request/response to audit log.
	req.Set(HeaderVersion, BundlesVersion)

	if !h.push {
		return server.SendMessageResponse(req, fiber.StatusBadRequest, msgNoPushEndpoint)
	}

	if ok, err := h.authorize(req); !ok || err != nil {
		return err
	}

	// TODO: retrieve bundle from body

	// TODO: decompress bundle

	// TODO: deploy bundle

	resp := &bundles2.BundleActivated{PreviousVersion: 0}
	return req.JSON(resp)
}

func (h *bundlesHandler) authorize(req *fiber.Ctx) (bool, error) {
	if h.authorizer == nil {
		return true, nil
	}

	resp, err := h.authorizer.Authorize(authRequest.FormatRequest(req))

	// TODO: log authorization decision to audit log.

	return authRequest.Check(req, resp, err)
}

type bundlesHandler struct {
	push       bool
	logger     *slog.Logger
	manager    *bundles.Manager
	cfg        []*bundles.Config
	authorizer authorization.Authorizer
}

const (
	msgOnlyPushEndpoint = "only bundle push endpoint configured"
	msgNoPushEndpoint   = "no bundle push endpoint configured"
)
