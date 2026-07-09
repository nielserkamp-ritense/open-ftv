// Package handler exposes the ODRL im-/export of a PAP as HTTP endpoints
// (Fiber), following the handler pattern of eam/handlers/fiber:
//
//	GET  /v1/odrl/export                     the whole PAP as ODRL-AP-NL
//	GET  /v1/odrl/export/{language}/{id...}  one policy as ODRL-AP-NL
//	POST /v1/odrl/import                     import an ODRL document (or {"url": ...})
//	POST /v1/odrl/events                     import via CloudEvents 1.0 push
package handler

import (
	"bytes"
	"log/slog"
	"strings"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"

	pap2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/odrl/export"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/odrl/importer"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/odrl/model"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/cloudevents"
	mime "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/io"
)

// Version is the semantic API version for the ODRL endpoints.
const Version = "1.0.0"

// Handler handles the ODRL im-/export endpoints.
type Handler struct {
	logger   *slog.Logger
	pap      pap2.PAP
	exporter *export.Exporter
	importer *importer.Importer
	dedup    *cloudevents.Dedup
}

// New creates a Handler around the given exporter/importer.
func New(logger *slog.Logger, p pap2.PAP, exp *export.Exporter, imp *importer.Importer) *Handler {
	if logger == nil {
		logger = slog.Default()
	}
	return &Handler{
		logger:   logger,
		pap:      p,
		exporter: exp,
		importer: imp,
		dedup:    cloudevents.NewDedup(0),
	}
}

// Register adds the ODRL routes to the given router group (typically /v1).
func (h *Handler) Register(group fiber.Router) {
	group.Get("/odrl/export", h.Export)
	group.Get("/odrl/export/:language/+", h.ExportPolicy)
	group.Post("/odrl/import", h.Import)
	group.Post("/odrl/events", h.Events)
}

// Export serves the whole PAP as one ODRL-AP-NL document.
// The response syntax follows the Accept header: text/turtle (default) or
// application/ld+json.
func (h *Handler) Export(req *fiber.Ctx) error {
	req.Set("API-Version", Version)

	doc, err := h.exporter.ExportAll()
	if err != nil {
		return sendMessage(req, fiber.StatusInternalServerError, err.Error())
	}

	return h.serialize(req, doc)
}

// ExportPolicy serves one policy key as an ODRL-AP-NL document.
func (h *Handler) ExportPolicy(req *fiber.Ctx) error {
	req.Set("API-Version", Version)

	language := req.Params("language")
	id := req.Params("+")
	if language == "" || id == "" {
		return sendMessage(req, fiber.StatusBadRequest, "language and id must be filled")
	}

	doc, err := h.exporter.ExportPolicy(language, id)
	if err != nil {
		return sendMessage(req, fiber.StatusNotFound, err.Error())
	}

	return h.serialize(req, doc)
}

// Import ingests an ODRL-AP-NL document. The request body is either the
// document itself (Content-Type text/turtle or application/ld+json) or a JSON
// reference: {"url": "https://..."}.
func (h *Handler) Import(req *fiber.Ctx) error {
	req.Set("API-Version", Version)

	body := req.Body()
	if len(body) == 0 {
		return sendMessage(req, fiber.StatusBadRequest, "empty request body")
	}

	ct := contentType(req)

	var (
		res *importer.Result
		err error
	)

	switch ct {
	case mime.MimeTypeTurtle, mime.MimeTypeJSONLD:
		res, err = h.importer.Import(req.Context(), body, ct)
	case mime.MimeTypeJSON, "":
		ref := &reference{}
		if err2 := json.Unmarshal(body, ref); err2 != nil || ref.URL == "" {
			return sendMessage(req, fiber.StatusBadRequest, "body must be text/turtle, application/ld+json or {\"url\": ...}")
		}
		res, err = h.importer.ImportURL(req.Context(), ref.URL)
	default:
		return sendMessage(req, fiber.StatusUnsupportedMediaType, "unsupported content type "+ct)
	}

	if err != nil {
		return sendMessage(req, fiber.StatusBadRequest, err.Error())
	}

	h.logger.Info("odrl import", "policies", len(res.Policies), "artifacts", len(res.Artifacts), "warnings", len(res.Warnings))
	return req.JSON(res)
}

// Events ingests an ODRL policy pushed as a CloudEvents 1.0 event (structured
// or binary mode). The event data is either an ODRL document (datacontenttype
// text/turtle or application/ld+json) or a JSON reference {"url": ...}.
// Delivery is idempotent on the (source, id) tuple.
func (h *Handler) Events(req *fiber.Ctx) error {
	req.Set("API-Version", Version)

	event, err := cloudevents.Parse(contentType(req), headerMap(req), req.Body())
	if err != nil {
		return sendMessage(req, fiber.StatusBadRequest, err.Error())
	}

	if h.dedup.Seen(event) {
		return req.Status(fiber.StatusOK).JSON(fiber.Map{"status": "duplicate", "id": event.ID})
	}

	if len(event.Data) == 0 {
		// event without payload (e.g. a removal notification): acknowledged.
		return req.Status(fiber.StatusAccepted).JSON(fiber.Map{"status": "ignored", "id": event.ID})
	}

	var res *importer.Result

	dct := strings.ToLower(strings.TrimSpace(strings.Split(event.DataContentType, ";")[0]))
	switch dct {
	case mime.MimeTypeTurtle, mime.MimeTypeJSONLD:
		res, err = h.importer.Import(req.Context(), event.Data, dct)
	default:
		ref := &reference{}
		if err2 := json.Unmarshal(event.Data, ref); err2 == nil && ref.URL != "" {
			res, err = h.importer.ImportURL(req.Context(), ref.URL)
		} else {
			return sendMessage(req, fiber.StatusBadRequest, "unsupported event data content type "+event.DataContentType)
		}
	}

	if err != nil {
		return sendMessage(req, fiber.StatusBadRequest, err.Error())
	}

	h.logger.Info("odrl event ingested", "id", event.ID, "source", event.Source, "type", event.Type,
		"policies", len(res.Policies), "artifacts", len(res.Artifacts))
	return req.JSON(res)
}

// serialize writes an ODRL document in the syntax asked for by the Accept
// header: application/ld+json when requested, text/turtle otherwise.
func (h *Handler) serialize(req *fiber.Ctx, doc *model.Document) error {
	out := mime.MimeTypeTurtle
	if accept := strings.ToLower(string(req.Request().Header.Peek(fiber.HeaderAccept))); strings.Contains(accept, mime.MimeTypeJSONLD) {
		out = mime.MimeTypeJSONLD
	}

	var buf bytes.Buffer
	if err := doc.Serialize(&buf, out); err != nil {
		return sendMessage(req, fiber.StatusInternalServerError, err.Error())
	}

	req.Set(fiber.HeaderContentType, out)
	return req.Send(buf.Bytes())
}

func sendMessage(req *fiber.Ctx, status int, message string) error {
	return req.Status(status).JSON(fiber.Map{"message": message})
}

func contentType(req *fiber.Ctx) string {
	ct := string(req.Request().Header.ContentType())
	if idx := strings.IndexByte(ct, ';'); idx >= 0 {
		ct = ct[:idx]
	}
	return strings.ToLower(strings.TrimSpace(ct))
}

func headerMap(req *fiber.Ctx) map[string]string {
	headers := make(map[string]string)
	req.Request().Header.VisitAll(func(key, value []byte) {
		headers[string(key)] = string(value)
	})
	return headers
}

type reference struct {
	URL string `json:"url"`
}
