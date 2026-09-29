package fiber

import (
	"bytes"
	"errors"
	"log/slog"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	auth "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/bundles"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/management"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	server "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server/fiber"
	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/bundles"
)

// BundlesVersion is the full semantic API version for the bundle endpoints.
const BundlesVersion = "1.0.1" // check against oas/bundles/openapi.yaml!

var (
	errDeploymentNotFound = errors.New("deployment not found")
	errDeploymentInternal = errors.New("internal error")
)

// BundlesHandler contains the details for handling requests about bundles and deployments.
type BundlesHandler struct {
	logger  *slog.Logger
	svc     *management.DeploymentService
	secured bool // an authorizer decides; the routes must carry the Identify middleware.
	principalResolver
}

// NewBundlesHandler instantiates a bundle deployment handler. Every operation is decided by
// the deployment service through the authorizer; a nil authorizer permits everything.
func NewBundlesHandler(logger *slog.Logger, store *pap.PAP, manager *bundles.Manager, authorizer authorization.RequestAuthorizer, opts ...HandlerOption) *BundlesHandler {
	return &BundlesHandler{
		logger:            logger,
		svc:               management.NewDeploymentService(logger, store, manager, authorizer),
		secured:           authorizer != nil,
		principalResolver: newPrincipalResolver(logger, opts),
	}
}

// caller is the request-scoped Principal this request runs as.
func (h *BundlesHandler) caller(req *fiber.Ctx) *authorization.RequestPrincipal {
	return callerOf(req, h.secured, h.logger)
}

// GetStatuses is the endpoint for retrieving the list of bundle status codes.
func (h *BundlesHandler) GetStatuses(req *fiber.Ctx) error {
	req.Set(HeaderVersion, BundlesVersion)

	if err := h.svc.Describe(req.UserContext(), h.caller(req)); err != nil {
		return h.fail(req, err)
	}

	resp := make(oas.Statuses, 0, bundles.StatusCount)
	for s := bundles.StatusMIN; s <= bundles.StatusMAX; s++ {
		resp = append(resp, oas.Status{Code: int(s), Name: s.String()})
	}

	return h.respond(req, resp)
}

// GetCompressTypes is the endpoint for retrieving the list of compression types.
func (h *BundlesHandler) GetCompressTypes(req *fiber.Ctx) error {
	req.Set(HeaderVersion, BundlesVersion)

	if err := h.svc.Describe(req.UserContext(), h.caller(req)); err != nil {
		return h.fail(req, err)
	}

	resp := make(oas.CompressTypes, 0, bundles.CompressCount)
	for s := bundles.CompressMIN; s <= bundles.CompressMAX; s++ {
		resp = append(resp, oas.CompressType{Code: int(s), Name: s.String()})
	}

	return h.respond(req, resp)
}

// GetConfigs is the endpoint for retrieving the bundle configurations.
func (h *BundlesHandler) GetConfigs(req *fiber.Ctx) error {
	req.Set(HeaderVersion, BundlesVersion)

	cfgs, err := h.svc.Configs(req.UserContext(), h.caller(req))
	if err != nil {
		return h.fail(req, err)
	}

	resp := make(oas.BundleConfigs, len(cfgs))

	for i := range cfgs {
		cfg := cfgs[i]
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

	return h.respond(req, resp)
}

// GetDeployments is the endpoint for retrieving the list of all bundle deployments.
func (h *BundlesHandler) GetDeployments(req *fiber.Ctx) error {
	req.Set(HeaderVersion, BundlesVersion)

	list, err := h.svc.List(req.UserContext(), h.caller(req))
	if err != nil {
		return h.fail(req, err)
	}

	resp := make([]*oas.Deployment, 0, len(list))
	for i := range list {
		resp = append(resp, list[i].ToOAS())
	}

	return h.respond(req, resp)
}

// GetDeployment is the endpoint for retrieving a specific bundle deployment.
func (h *BundlesHandler) GetDeployment(req *fiber.Ctx) error {
	req.Set(HeaderVersion, BundlesVersion)

	version, err := strconv.ParseUint(req.Params("key"), 10, 64)
	if err != nil {
		err = errors.New("deployment version must be a positive number")

		// a malformed version is still answered in decision order: the denial comes first.
		if derr := h.svc.Describe(req.UserContext(), h.caller(req)); derr != nil {
			return h.fail(req, derr)
		}

		return h.error(req, fiber.StatusBadRequest, err)
	}

	resp, err := h.svc.Get(req.UserContext(), h.caller(req), version)
	if err != nil {
		return h.fail(req, err)
	}

	return h.respond(req, resp.ToOAS())
}

// GetLastDeployment is the endpoint for retrieving the last bundle deployment.
func (h *BundlesHandler) GetLastDeployment(req *fiber.Ctx) error {
	req.Set(HeaderVersion, BundlesVersion)

	resp, err := h.svc.Last(req.UserContext(), h.caller(req))
	if err != nil {
		return h.fail(req, err)
	}

	return h.respond(req, resp.ToOAS())
}

// PostDeployment is the endpoint for starting a new bundle deployment.
func (h *BundlesHandler) PostDeployment(req *fiber.Ctx) error {
	req.Set(HeaderVersion, BundlesVersion)

	var body oas.NewDeploymentBody
	if err := req.BodyParser(&body); err != nil {
		// a malformed body is still answered in decision order: the denial comes first.
		if derr := h.svc.DecideDeploy(req.UserContext(), h.caller(req)); derr != nil {
			return h.fail(req, derr)
		}

		return server.SendProblemResponse(req, fiber.StatusBadRequest, codeBadRequest, err.Error())
	}

	resp, err := h.svc.Deploy(req.UserContext(), h.caller(req), body.Title, body.Description)
	if err != nil {
		return h.fail(req, err)
	}

	return h.respond(req, resp.ToOAS())
}

// GetBundle is the endpoint for retrieving the last deployment bundle.
func (h *BundlesHandler) GetBundle(req *fiber.Ctx) error {
	req.Set(HeaderVersion, BundlesVersion)

	id := req.Params("id")
	if id == "" {
		return h.error(req, fiber.StatusBadRequest, errors.New("bundle id required"))
	}

	b, err := h.svc.Bundle(req.UserContext(), h.caller(req), id)
	if err != nil {
		return h.fail(req, err)
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

// fail answers a failed service call: a denial is 403 regardless of whether the object
// exists, a missing deployment 404, an unknown bundle or a deployment refused because the
// last one still runs 400, and anything else (a store failure, or a permitted caller that
// could not be recorded) 500 without the cause.
func (h *BundlesHandler) fail(req *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, authorization.ErrForbidden):
		return auth.Forbidden(req, err, h.logger)
	case errors.Is(err, management.ErrNotFound):
		return h.error(req, fiber.StatusNotFound, errDeploymentNotFound)
	case errors.Is(err, management.ErrBundle):
		return h.error(req, fiber.StatusBadRequest, err)
	default:
		h.logger.Error("request error", "path", req.Path(), "err", err, "status", fiber.StatusInternalServerError)
		return server.SendMessageResponse(req, fiber.StatusInternalServerError, errDeploymentInternal.Error())
	}
}

func (h *BundlesHandler) error(req *fiber.Ctx, status int, err error) error {
	h.logger.Error("request error", "path", req.Path(), "err", err, "status", status)
	return server.SendMessageResponse(req, status, err.Error())
}
