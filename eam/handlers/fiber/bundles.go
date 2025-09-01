package fiber

import (
	"bytes"
	"errors"
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v2"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	auth "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/bundles"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	server "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server/fiber"
	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/bundles"
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

	_, ok, err := h.authorize(req)
	if !ok {
		return err
	}

	resp := make(oas.Statuses, 0, bundles.StatusCount)
	for s := bundles.StatusMIN; s <= bundles.StatusMAX; s++ {
		resp = append(resp, oas.Status{Code: int(s), Name: s.String()})
	}
	return req.JSON(resp)
}

// GetCompressTypes is the endpoint for retrieving the list of compression types.
func (h *BundlesHandler) GetCompressTypes(req *fiber.Ctx) error {
	// TODO: log request/response to audit log.
	req.Set(HeaderVersion, BundlesVersion)

	_, ok, err := h.authorize(req)
	if !ok {
		return err
	}

	resp := make(oas.CompressTypes, 0, bundles.CompressCount)
	for s := bundles.CompressMIN; s <= bundles.CompressMAX; s++ {
		resp = append(resp, oas.CompressType{Code: int(s), Name: s.String()})
	}
	return req.JSON(resp)
}

// GetConfigs is the endpoint for retrieving the bundle configurations.
func (h *BundlesHandler) GetConfigs(req *fiber.Ctx) error {
	// TODO: log request/response to audit log.
	req.Set(HeaderVersion, BundlesVersion)

	_, ok, err := h.authorize(req)
	if !ok {
		return err
	}

	resp := make(oas.BundleConfigs, len(h.cfg))

	for i := range h.cfg {
		cfg := h.cfg[i]
		cfg2 := oas.BundleConfig{
			Id:       cfg.ID,
			Version:  cfg.Version,
			Policies: cfg.Policies,
			Data:     cfg.Data,
			Tags:     cfg.Tags,
			Targets:  make([]oas.Target, 0, len(cfg.Targets)),
		}

		for j := range cfg.Targets {
			target := cfg.Targets[j]
			cfg2.Targets = append(cfg2.Targets, oas.Target{
				Uri:         target.URI,
				Tls:         target.CA != "" || target.Cert != "" || target.Key != "",
				Apikey:      target.APIKey != "",
				Compression: int(bundles.CompressionTypeFromString(target.Encoding)),
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

	_, ok, err := h.authorize(req)
	if !ok {
		return err
	}

	var list []*bundles.Deployment
	if list, err = h.pap.ListDeployments(); err != nil {
		return h.error(req, fiber.StatusInternalServerError, err)
	}

	resp := make([]*oas.Deployment, 0, len(list))
	for i := range list {
		resp = append(resp, list[i].ToOAS())
	}

	return req.JSON(resp)
}

// GetDeployment is the endpoint for retrieving a specific bundle deployment.
func (h *BundlesHandler) GetDeployment(req *fiber.Ctx) error {
	// TODO: log request/response to audit log.
	req.Set(HeaderVersion, BundlesVersion)

	_, ok, err := h.authorize(req)
	if !ok {
		return err
	}

	var version int
	if version, err = req.ParamsInt("key"); err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	var resp *bundles.Deployment
	if resp, err = h.pap.ReadDeployment(uint64(version)); err != nil {
		return h.error(req, fiber.StatusNotFound, err)
	}
	return req.JSON(resp.ToOAS())
}

// GetLastDeployment is the endpoint for retrieving the last bundle deployment.
func (h *BundlesHandler) GetLastDeployment(req *fiber.Ctx) error {
	// TODO: log request/response to audit log.
	req.Set(HeaderVersion, BundlesVersion)

	_, ok, err := h.authorize(req)
	if !ok {
		return err
	}

	var last *bundles.Deployment
	if last, err = h.pap.LastDeployment(); err != nil {
		return h.error(req, fiber.StatusInternalServerError, err)
	}

	if last == nil || last.Version() <= 0 {
		return h.error(req, fiber.StatusNotFound, errors.New("last deployment not found"))
	}

	var resp *bundles.Deployment
	if resp, err = h.pap.ReadDeployment(last.Version()); err != nil {
		return h.error(req, fiber.StatusInternalServerError, err)
	}
	return req.JSON(resp.ToOAS())
}

// PostDeployment is the endpoint for starting a new bundle deployment.
func (h *BundlesHandler) PostDeployment(req *fiber.Ctx) error {
	// TODO: log request/response to audit log.
	req.Set(HeaderVersion, BundlesVersion)

	user, ok, err := h.authorize(req)
	if !ok {
		return err
	}

	var body oas.NewDeploymentBody
	if err = req.BodyParser(&body); err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	var resp *bundles.Deployment
	if resp, err = h.pap.NewDeployment(body.Description, h.manager, user); err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}
	return req.JSON(resp.ToOAS())
}

// GetBundle is the endpoint for retrieving the last deployment bundle.
func (h *BundlesHandler) GetBundle(req *fiber.Ctx) error {
	// TODO: log request/response to audit log.
	req.Set(HeaderVersion, BundlesVersion)

	_, ok, err := h.authorize(req)
	if !ok {
		return err
	}

	id := req.Params("id")
	if id == "" {
		return h.error(req, fiber.StatusBadRequest, errors.New("bundle id required"))
	}

	var last *bundles.Deployment
	if last, err = h.pap.LastDeployment(); err != nil {
		return h.error(req, fiber.StatusInternalServerError, err)
	}

	if last == nil || last.Version() <= 0 {
		return h.error(req, fiber.StatusNotFound, errors.New("last deployment not found"))
	}

	var b *bundles.Bundle
	if b, err = h.manager.Get(last, id); err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	enc := req.AcceptsEncodings("gz", "gzip", "bzip", "bz2", "bzip2")
	if enc == "" {
		enc = "gz"
	}
	ct := bundles.CompressionTypeFromString(enc)

	var buf bytes.Buffer
	if err = b.Compress(ct, &buf); err != nil {
		return h.error(req, fiber.StatusInternalServerError, err)
	}

	req.Set(fiber.HeaderContentEncoding, strings.ToLower(ct.String()))
	req.Set(fiber.HeaderContentType, fiber.MIMEOctetStream)
	return req.Send(buf.Bytes())
}

func (h *BundlesHandler) authorize(req *fiber.Ctx) (string, bool, error) {
	if h.authorizer == nil {
		return auth.SystemUser, true, nil
	}

	resp, err := h.authorizer.Authorize(auth.FormatRequest(req))

	// TODO: log authorization decision to auth-decision log.

	return auth.Check(req, resp, err, h.logger)
}

func (h *BundlesHandler) error(req *fiber.Ctx, status int, err error) error {
	h.logger.Error("request error", "path", req.Path(), "err", err, "status", status)
	return server.SendMessageResponse(req, status, err.Error())
}
