package fiber

import (
	"bytes"
	"io"
	"log/slog"

	"github.com/gofiber/fiber/v2"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/bundles"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/identity"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller"
	server "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server/fiber"
	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/bundles"
)

// BundleReceiverHandler is used to handle receiving new bundles.
type BundleReceiverHandler struct {
	logger     *slog.Logger
	ctl        pdp.Controller
	authorizer authorization.RequestAuthorizer
}

// NewBundleReceiverHandler instantiates a bundle receiver handler.
func NewBundleReceiverHandler(logger *slog.Logger, ctl pdp.Controller, authorizer authorization.RequestAuthorizer) *BundleReceiverHandler {
	return &BundleReceiverHandler{logger: logger, ctl: ctl, authorizer: authorizer}
}

// PostBundle is the endpoint for receiving a new bundle.
func (h *BundleReceiverHandler) PostBundle(req *fiber.Ctx) error {
	req.Set(HeaderVersion, BundlesVersion)

	if _, err := h.authorize(req); err != nil {
		return err
	}

	ct := req.Get(fiber.HeaderContentEncoding)
	body := bytes.NewReader(req.BodyRaw())

	oldVersion, err2 := h.ProcessBundle(ct, body)
	if err2 != nil {
		return h.error(req, fiber.StatusBadRequest, err2)
	}

	resp := &oas.BundleActivated{PreviousVersion: int(oldVersion)}
	return req.Status(fiber.StatusCreated).JSON(resp)
}

// ProcessBundle processes a received bundle and replaces all policies/data.
func (h *BundleReceiverHandler) ProcessBundle(ct string, body io.Reader) (old uint64, err error) {
	bundle, err := bundles.Decompress(bundles.CompressionTypeFromString(ct), body)
	if err != nil {
		h.logger.Error("failed to decode new bundle", "compression", ct, "err", err)
		return
	}

	if old, err = h.ctl.NewBundle(bundle); err != nil {
		h.logger.Error("failed to activate new bundle", "version", bundle.Version, "err", err)
		return
	}

	h.logger.Info("new bundle activated", "version", bundle.Version)
	return
}

func (h *BundleReceiverHandler) authorize(req *fiber.Ctx) (identity.Principal, error) {
	return authorizeRequest(h.authorizer, req, h.logger)
}

func (h *BundleReceiverHandler) error(req *fiber.Ctx, status int, err error) error {
	h.logger.Error("request error", "path", req.Path(), "err", err, "status", status)
	return server.SendMessageResponse(req, status, err.Error())
}
