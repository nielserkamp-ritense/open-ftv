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
	authRequest "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/bundles"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller"
	server "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server/fiber"
	bundles2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/bundles"
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

	if ok, err := h.authorize(req); !ok || err != nil {
		return err
	}

	var r io.Reader
	var err error

	ct := bundles.CompressionTypeFromString(req.Get(fiber.HeaderContentEncoding))
	switch ct {
	case bundles.CompressBZ2:
		r, err = bzip2.NewReader(bytes.NewReader(req.BodyRaw()), &bzip2.ReaderConfig{})
	default:
		r, err = gzip.NewReader(bytes.NewReader(req.BodyRaw()))
	}

	if err != nil {
		return server.SendMessageResponse(req, fiber.StatusBadRequest, err.Error())
	}

	bundle := new(bundles.Bundle)
	if err = json.NewDecoder(r).Decode(bundle); err != nil {
		return server.SendMessageResponse(req, fiber.StatusBadRequest, err.Error())
	}

	oldVersion, err2 := h.ctl.NewBundle(bundle)
	if err2 != nil {
		return server.SendMessageResponse(req, fiber.StatusInternalServerError, err2.Error())
	}

	resp := &bundles2.BundleActivated{PreviousVersion: int(oldVersion)}
	return req.JSON(resp)
}

func (h *BundleReceiverHandler) authorize(req *fiber.Ctx) (bool, error) {
	if h.authorizer == nil {
		return true, nil
	}

	resp, err := h.authorizer.Authorize(authRequest.FormatRequest(req))

	// TODO: log authorization decision to auth-decision log.

	return authRequest.Check(req, resp, err)
}
