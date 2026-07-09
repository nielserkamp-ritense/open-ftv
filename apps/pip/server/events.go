package server

import (
	"log/slog"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	authRequest "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/handlers"
	handle "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/handlers/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	pip2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip/cloudevents"
	server "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/attributes"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/warc"
)

// EventsVersion is the full semantic API version for the CloudEvents ingest endpoint.
const EventsVersion = "1.0.0"

// eventPayload is the expected shape of the CloudEvents data: attribute and/or entity
// updates in the same (OAS attributes) format as the pull/CRUD side of the PIP.
type eventPayload struct {
	Attributes []attributes.Attribute `json:"attributes,omitempty"`
	Entities   []attributes.Entity    `json:"entities,omitempty"`
}

// eventsHandler handles CloudEvents push-ingest for the PIP.
type eventsHandler struct {
	logger     *slog.Logger
	pip        pip2.PIP
	authorizer authorization.Authorizer
	warc       *warc.Writer
	dedup      *cloudevents.Dedup
}

func newEventsHandler(logger *slog.Logger, p pip2.PIP, authorizer authorization.Authorizer, w *warc.Writer) *eventsHandler {
	return &eventsHandler{logger: logger, pip: p, authorizer: authorizer, warc: w, dedup: cloudevents.NewDedup(8192)}
}

// PostEvent handles POST /v1/events: a CloudEvents 1.0 event in structured or binary mode.
func (h *eventsHandler) PostEvent(req *fiber.Ctx) error {
	req.Set(handle.HeaderVersion, EventsVersion)

	if ok, err := h.authorize(req); !ok || err != nil {
		return err
	}

	// determine the trace context: keep an inbound trace id, always create a new span.
	traceID, parentID, _ := cloudevents.ParseTraceParent(req.Get("traceparent"))
	if traceID == "" {
		traceID = warc.NewTraceID()
	}
	spanID := warc.NewSpanID()

	event, err := cloudevents.Parse(req.Get(fiber.HeaderContentType), headersMap(req), req.Body())
	if err != nil {
		h.logWARC(req, traceID, spanID, parentID, nil)
		return server.SendMessageResponse(req, fiber.StatusBadRequest, err.Error())
	}

	// log the incoming push event to the WARC before applying it.
	file := h.logWARC(req, traceID, spanID, parentID, event)

	// idempotent ingest on (source, id): acknowledge duplicates without reapplying.
	if h.dedup.Seen(event) {
		return req.Status(fiber.StatusOK).JSON(fiber.Map{"status": "duplicate", "id": event.ID, "source": event.Source})
	}

	var payload eventPayload
	if err = json.Unmarshal(event.Data, &payload); err != nil {
		return server.SendMessageResponse(req, fiber.StatusBadRequest, "invalid event data: "+err.Error())
	}
	if len(payload.Attributes) == 0 && len(payload.Entities) == 0 {
		return server.SendMessageResponse(req, fiber.StatusBadRequest, "event data must contain attributes and/or entities")
	}

	refs, _ := h.pip.(pip2.SourceReferencer)
	record := func(kind, key string) {
		if refs != nil {
			refs.RecordSourceRef(pip2.SourceRef{
				Kind: kind, Key: key,
				TraceID: traceID, SpanID: spanID, ParentSpanID: parentID,
				WARCFile: file,
				Version:  event.DataVersion(), Sequence: event.Sequence(),
				Time: time.Now().UTC(),
			})
		}
	}

	// apply the updates with the same OAS decoding as the CRUD endpoints.
	for i := range payload.Attributes {
		a := handlers.AttributeFromOAS(&payload.Attributes[i])
		h.pip.AddOriginalAttribute(a.Key(), a.Value(), a.Original(), a.Type())
		record("attribute", a.Key())
	}

	for i := range payload.Entities {
		e := handlers.EntityFromOAS(&payload.Entities[i], h.pip.NewAttributeSet())
		h.pip.AddEntity(e)
		record("entity", e.UID())
	}

	h.logger.Info("cloud event ingested",
		"id", event.ID, "source", event.Source, "type", event.Type,
		"dataversion", event.DataVersion(), "sequence", event.Sequence(),
		"attributes", len(payload.Attributes), "entities", len(payload.Entities),
		"trace", traceID, "span", spanID, "warc", file)

	return req.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":     "accepted",
		"id":         event.ID,
		"source":     event.Source,
		"attributes": len(payload.Attributes),
		"entities":   len(payload.Entities),
		"traceId":    traceID,
		"spanId":     spanID,
	})
}

// GetSourceRefs handles GET /v1/sourcerefs: all known source references.
func (h *eventsHandler) GetSourceRefs(req *fiber.Ctx) error {
	req.Set(handle.HeaderVersion, EventsVersion)

	if ok, err := h.authorize(req); !ok || err != nil {
		return err
	}

	refs, ok := h.pip.(pip2.SourceReferencer)
	if !ok {
		return server.SendMessageResponse(req, fiber.StatusNotFound, "source references not supported")
	}
	return req.JSON(refs.ListSourceRefs())
}

// GetSourceRef handles GET /v1/sourceref?key=...: the source reference for one
// attribute key or entity UID.
func (h *eventsHandler) GetSourceRef(req *fiber.Ctx) error {
	req.Set(handle.HeaderVersion, EventsVersion)

	if ok, err := h.authorize(req); !ok || err != nil {
		return err
	}

	key := req.Query("key")
	if key == "" {
		return server.SendMessageResponse(req, fiber.StatusBadRequest, "query parameter 'key' is required")
	}

	refs, ok := h.pip.(pip2.SourceReferencer)
	if !ok {
		return server.SendMessageResponse(req, fiber.StatusNotFound, "source references not supported")
	}

	ref, found := refs.SourceRef(key)
	if !found {
		return server.SendMessageResponse(req, fiber.StatusNotFound, "no source reference for key")
	}
	return req.JSON(ref)
}

// logWARC appends the incoming push exchange to the WARC log as a request/resource
// record pair carrying the trace/span identifiers, and returns the WARC file name.
func (h *eventsHandler) logWARC(req *fiber.Ctx, traceID, spanID, parentID string, event *cloudevents.Event) string {
	if h.warc == nil {
		return ""
	}

	custom := map[string]string{
		warc.HeaderTraceID:  traceID,
		warc.HeaderSpanID:   spanID,
		warc.HeaderParentID: parentID,
	}
	if event != nil {
		custom[warc.HeaderVersion] = event.DataVersion()
		custom[warc.HeaderSequence] = event.Sequence()
	}

	target := req.BaseURL() + req.OriginalURL()
	reqID, resID := warc.NewUUID(), warc.NewUUID()

	// the raw HTTP request as received.
	reqRec := &warc.Record{
		Type:         warc.TypeRequest,
		RecordID:     reqID,
		TargetURI:    target,
		ConcurrentTo: resID,
		ContentType:  warc.CTRequest,
		Custom:       custom,
		Block:        []byte(req.Request().String()),
	}

	// the decoded event as a resource record (canonical JSON), when available.
	var block []byte
	if event != nil {
		block, _ = json.Marshal(fiber.Map{
			"specversion":     event.SpecVersion,
			"id":              event.ID,
			"source":          event.Source,
			"type":            event.Type,
			"subject":         event.Subject,
			"time":            event.Time,
			"datacontenttype": event.DataContentType,
			"extensions":      event.Extensions,
			"data":            json.RawMessage(event.Data),
		})
	}
	resRec := &warc.Record{
		Type:         warc.TypeResource,
		RecordID:     resID,
		TargetURI:    target,
		ConcurrentTo: reqID,
		ContentType:  "application/cloudevents+json",
		Custom:       custom,
		Block:        block,
	}

	file, err := h.warc.Write(reqRec)
	if err != nil {
		h.logger.Warn("failed to write WARC request record", "error", err)
		return ""
	}
	if _, err = h.warc.Write(resRec); err != nil {
		h.logger.Warn("failed to write WARC resource record", "error", err)
	}
	return file
}

func (h *eventsHandler) authorize(req *fiber.Ctx) (bool, error) {
	if h.authorizer == nil {
		return true, nil
	}

	resp, err := h.authorizer.Authorize(authRequest.FormatRequest(req))
	return authRequest.Check(req, resp, err)
}

// headersMap collects all request headers into a plain map (last value wins).
func headersMap(req *fiber.Ctx) map[string]string {
	out := make(map[string]string, 16)
	req.Request().Header.VisitAll(func(key, value []byte) {
		out[string(key)] = string(value)
	})
	return out
}

// slogEventSink implements models.EventSink by logging every PIP mutation.
type slogEventSink struct {
	logger *slog.Logger
}

// Handle implements the models.EventSink interface.
func (s *slogEventSink) Handle(t models.EventType, key string) {
	s.logger.Debug("pip event", "event", t.String(), "key", key)
}
