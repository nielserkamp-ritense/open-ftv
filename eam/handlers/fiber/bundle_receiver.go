package fiber

import (
	"bytes"
	"compress/gzip"
	"io"
	"log/slog"

	"github.com/dsnet/compress/bzip2"
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	auth "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/bundles"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller"
	server "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server/fiber"
	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/bundles"
)

// BundleReceiverHandler is used to handle receiving new bundles.
type BundleReceiverHandler struct {
	logger     *slog.Logger
	ctl        pdp.Controller
	authorizer authorization.Authorizer
}

// NewBundleReceiverHandler instantiates a bundle receiver handler.
func NewBundleReceiverHandler(logger *slog.Logger, ctl pdp.Controller, authorizer authorization.Authorizer) *BundleReceiverHandler {
	return &BundleReceiverHandler{logger: logger, ctl: ctl, authorizer: authorizer}
}

// PostBundle is the endpoint for receiving a new bundle.
func (h *BundleReceiverHandler) PostBundle(req *fiber.Ctx) error {
	// TODO: log request/response to audit log.
	req.Set(HeaderVersion, BundlesVersion)

	_, ok, err := h.authorize(req)
	if !ok {
		return err
	}

	var r io.Reader

	ct := bundles.CompressionTypeFromString(req.Get(fiber.HeaderContentEncoding))
	switch ct {
	case bundles.CompressBZ2:
		r, err = bzip2.NewReader(bytes.NewReader(req.BodyRaw()), &bzip2.ReaderConfig{})
	default:
		r, err = gzip.NewReader(bytes.NewReader(req.BodyRaw()))
	}

	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	bundle := new(bundles.Bundle)
	if err = json.NewDecoder(r).Decode(bundle); err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	oldVersion, err2 := h.ctl.NewBundle(bundle)
	if err2 != nil {
		return h.error(req, fiber.StatusInternalServerError, err2)
	}

	resp := &oas.BundleActivated{PreviousVersion: int(oldVersion)}
	return req.JSON(resp)
}

func (h *BundleReceiverHandler) authorize(req *fiber.Ctx) (string, bool, error) {
	if h.authorizer == nil {
		return auth.SystemUser, true, nil
	}

	resp, err := h.authorizer.Authorize(auth.FormatRequest(req))

	// TODO: log authorization decision to auth-decision log.

	return auth.Check(req, resp, err, h.logger)
}

func (h *BundleReceiverHandler) error(req *fiber.Ctx, status int, err error) error {
	h.logger.Error("request error", "path", req.Path(), "err", err, "status", status)
	return server.SendMessageResponse(req, status, err.Error())
}
