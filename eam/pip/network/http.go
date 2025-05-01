package network

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"k8s.io/client-go/transport"
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
	if r.httpResp, r.err = r.httpClient.Do(r.httpReq); r.err != nil {
		r.logger.Error("failed to execute http request", "error", r.err)
		return
	}

	defer r.httpResp.Body.Close()
	r.processResponse()
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
	err        error
}

const (
	msgStarted = "processing scheduled job"
	msgFailed  = "scheduled job processing failed"
	msgOK      = "scheduled job processed successfully"
)
