package network

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"k8s.io/client-go/transport"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/warc"
)

func (m *manager) execute(logger *slog.Logger, req *Request) error {
	req.prepare()

	ctx, cancel := context.WithTimeout(m.ctx, req.Timeout)
	defer cancel()

	r := runner{manager: m, ctx: ctx, logger: logger, req: req}
	r.run()

	return r.err
}

func (r *runner) run() {
	if r.logger.Enabled(nil, slog.LevelDebug) {
		r.logger.Debug(msgStarted)
		r.msg = msgFailed
		defer func() { r.logger.Debug(r.msg) }()
	}

	if r.initClient() && r.initRequest() {
		r.doRequest()
	}
}

func (r *runner) initClient() bool {
	trans := http.DefaultTransport

	if cfg := r.req.TLSConfig(); cfg != nil {
		if trans, r.err = transport.New(&transport.Config{TLS: *cfg}); r.err != nil {
			r.logger.Error("failed to setup TLS for https request", "error", r.err)
			return false
		}
	}

	r.httpClient = &http.Client{
		Transport:     trans,
		CheckRedirect: http.DefaultClient.CheckRedirect,
		Jar:           http.DefaultClient.Jar,
		Timeout:       r.req.Timeout,
	}
	return true
}

func (r *runner) initRequest() bool {
	if r.httpReq, r.err = r.req.HTTPRequest(r.ctx, r.manager.attributes); r.err != nil {
		r.logger.Error("failed to build http request", "error", r.err)
		return false
	}
	return true
}

func (r *runner) doRequest() {
	// Capture the outgoing request bytes before the body is consumed by the transport.
	reqDump := warc.DumpRequest(r.httpReq, true)

	if r.httpResp, r.err = r.httpClient.Do(r.httpReq); r.err != nil {
		r.logger.Error("failed to execute http request", "error", r.err)
		r.logWARC(reqDump)
		return
	}

	defer r.httpResp.Body.Close()

	// DumpResponse restores the response body so decoding can still read it afterwards.
	r.logWARC(reqDump)
	r.processResponse()
}

// logWARC writes the request/response pair to the WARC log (if configured) and remembers
// the trace/span identifiers and file so decoded values can register a source reference.
func (r *runner) logWARC(reqDump []byte) {
	if r.manager.warc == nil {
		return
	}

	r.traceID = warc.NewTraceID()
	r.spanID = warc.NewSpanID()

	respDump := warc.DumpResponse(r.httpResp, true)

	file, err := r.manager.warc.WritePair(r.traceID, r.spanID, r.httpReq.URL.String(), reqDump, respDump, nil)
	if err != nil {
		r.logger.Warn("failed to write WARC record", "error", err)
		return
	}
	r.warcFile = file
}

// attributeSet returns the attribute set decoded values should be written to.
//
// When a source recorder is configured, mutations are wrapped so each updated key is
// registered as a "Logged" source reference pointing at the WARC exchange.
func (r *runner) attributeSet() models.AttributeSet {
	if r.manager.recorder == nil {
		return r.manager.attributes
	}
	return &recordingAttributes{AttributeSet: r.manager.attributes, rec: func(key string) {
		r.manager.recorder("attribute", key, r.traceID, r.spanID, r.warcFile, "")
	}}
}

// entitySet returns the entity set decoded values should be written to (see attributeSet).
//
// Besides the entity UID, each attribute of the entity is registered under the
// composite key "<entity-uid>/<attr>". A PDP request delivers an entity's attributes
// under exactly that key shape, so registering them here lets the ADL level-3 hook
// resolve the source reference for the attributes the policy actually consulted -
// not only for the bare entity.
func (r *runner) entitySet() models.EntitySet {
	if r.manager.recorder == nil {
		return r.manager.entities
	}
	return &recordingEntities{EntitySet: r.manager.entities, rec: func(entity models.Entity) {
		uid := entity.UID()
		r.manager.recorder("entity", uid, r.traceID, r.spanID, r.warcFile, "")
		if attrs := entity.Attributes(); attrs != nil {
			attrs.IterateAttributes(func(a models.Attribute) {
				r.manager.recorder("attribute", uid+"/"+a.Key(), r.traceID, r.spanID, r.warcFile, "")
			})
		}
	}}
}

func (r *runner) processResponse() {
	if r.req.Mapping != nil && len(r.req.Mapping.StatusCodes) > 0 {
		r.decodeResponse()
		return
	}

	switch r.httpResp.StatusCode {
	case http.StatusOK:
		// process the returned data.
		r.decodeResponse()

	case http.StatusNotModified:
		// nothing changed, so we're all good.
		r.msg = msgOK

	default:
		r.logger.Error("unexpected response status code", "code", r.httpResp.StatusCode, "status", r.httpResp.Status, "headers", r.httpResp.Header)
		r.err = fmt.Errorf("unexpected response status code %d", r.httpResp.StatusCode)
	}
}

type runner struct {
	msg        string
	manager    *manager
	ctx        context.Context
	logger     *slog.Logger
	req        *Request
	httpClient *http.Client
	httpReq    *http.Request
	httpResp   *http.Response
	data       any
	traceID    string
	spanID     string
	warcFile   string
	err        error
}

const (
	msgStarted = "processing scheduled job"
	msgFailed  = "scheduled job processing failed"
	msgOK      = "scheduled job processed successfully"
)
