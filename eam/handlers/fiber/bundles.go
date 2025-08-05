package fiber

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	authRequest "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/bundles"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	server "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server/fiber"
	bundles2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/bundles"
)

// BundlesVersion is the full semantic API version for the bundle endpoints.
const BundlesVersion = "1.0.0" // check against oas/bundles/openapi.yaml!

// BundlesHandler contains the details for handling requests about bundles and deployments.
type BundlesHandler struct {
	logger     *slog.Logger
	pap        *pap.PAP
	manager    *bundles.Manager
	cfg        []*bundles.Config
	authorizer authorization.Authorizer
}

// NewBundlesHandler instantiates a bundle deployment handler.
func NewBundlesHandler(logger *slog.Logger, pap *pap.PAP, manager *bundles.Manager, authorizer authorization.Authorizer) *BundlesHandler {
	return &BundlesHandler{logger: logger, pap: pap, manager: manager, cfg: manager.Bundles(), authorizer: authorizer}
}

// GetStatuses is the endpoint for retrieving the list of bundle status codes.
func (h *BundlesHandler) GetStatuses(req *fiber.Ctx) error {
	// TODO: log request/response to audit log.
	req.Set(HeaderVersion, BundlesVersion)

	if ok, err := h.authorize(req); !ok || err != nil {
		return err
	}

	resp := make(bundles2.Statuses, 0, bundles.StatusCount)
	for s := bundles.StatusMIN; s <= bundles.StatusMAX; s++ {
		resp = append(resp, bundles2.Status{Code: int(s), Name: s.String()})
	}
	return req.JSON(resp)
}

// GetCompressTypes is the endpoint for retrieving the list of compression types.
func (h *BundlesHandler) GetCompressTypes(req *fiber.Ctx) error {
	// TODO: log request/response to audit log.
	req.Set(HeaderVersion, BundlesVersion)

	if ok, err := h.authorize(req); !ok || err != nil {
		return err
	}

	resp := make(bundles2.CompressTypes, 0, bundles.CompressCount)
	for s := bundles.CompressMIN; s <= bundles.CompressMAX; s++ {
		resp = append(resp, bundles2.CompressType{Code: int(s), Name: s.String()})
	}
	return req.JSON(resp)
}

// GetConfigs is the endpoint for retrieving the bundle configurations.
func (h *BundlesHandler) GetConfigs(req *fiber.Ctx) error {
	// TODO: log request/response to audit log.
	req.Set(HeaderVersion, BundlesVersion)

	if ok, err := h.authorize(req); !ok || err != nil {
		return err
	}

	resp := make(bundles2.BundleConfigs, len(h.cfg))

	for i := range h.cfg {
		cfg := h.cfg[i]
		cfg2 := bundles2.BundleConfig{
			Id:       cfg.ID,
			Version:  cfg.Version,
			Policies: cfg.Policies,
			Data:     cfg.Data,
			Tags:     cfg.Tags,
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

// GetDeployments is the endpoint for retrieving the list of all bundle deployments.
func (h *BundlesHandler) GetDeployments(req *fiber.Ctx) error {
	// TODO: log request/response to audit log.
	req.Set(HeaderVersion, BundlesVersion)

	if ok, err := h.authorize(req); !ok || err != nil {
		return err
	}

	resp, err := h.pap.ListDeployments()
	if err != nil {
		return server.SendMessageResponse(req, fiber.StatusInternalServerError, err.Error())
	}
	return req.JSON(resp)
}

// GetDeployment is the endpoint for retrieving a specific bundle deployment.
func (h *BundlesHandler) GetDeployment(req *fiber.Ctx) error {
	// TODO: log request/response to audit log.
	req.Set(HeaderVersion, BundlesVersion)

	if ok, err := h.authorize(req); !ok || err != nil {
		return err
	}

	version, err := req.ParamsInt("key")
	if err != nil {
		return server.SendMessageResponse(req, fiber.StatusBadRequest, err.Error())
	}

	resp, err2 := h.pap.ReadDeployment(uint64(version))
	if err2 != nil {
		return server.SendMessageResponse(req, fiber.StatusNotFound, err2.Error())
	}
	return req.JSON(resp)
}

// PostDeployment is the endpoint for starting a new bundle deployment.
func (h *BundlesHandler) PostDeployment(req *fiber.Ctx) error {
	// TODO: log request/response to audit log.
	req.Set(HeaderVersion, BundlesVersion)

	if ok, err := h.authorize(req); !ok || err != nil {
		return err
	}

	var body bundles2.NewDeploymentBody
	if err := req.BodyParser(&body); err != nil {
		return server.SendMessageResponse(req, fiber.StatusBadRequest, err.Error())
	}

	resp, err := h.pap.NewDeployment(body.Description, h.manager)
	if err != nil {
		return server.SendMessageResponse(req, fiber.StatusBadRequest, err.Error())
	}
	return req.JSON(resp)
}

func (h *BundlesHandler) authorize(req *fiber.Ctx) (bool, error) {
	if h.authorizer == nil {
		return true, nil
	}

	resp, err := h.authorizer.Authorize(authRequest.FormatRequest(req))

	// TODO: log authorization decision to auth-decision log.

	return authRequest.Check(req, resp, err)
}
