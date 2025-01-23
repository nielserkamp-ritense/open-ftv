package network

import (
	"context"
	"log/slog"
	"net/http"

	"k8s.io/client-go/transport"
)

func (m *manager) execute(logger *slog.Logger, req *Request) {
	ctx, cancel := context.WithTimeout(m.ctx, req.Timeout)
	defer cancel()

	r := runner{manager: m, ctx: ctx, logger: logger, req: req}
	r.run()
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
	trans := http.DefaultClient.Transport

	if cfg := r.req.TLSConfig(); cfg != nil {
		var err error
		if trans, err = transport.New(&transport.Config{TLS: *cfg, Transport: trans}); err != nil {
			r.logger.Error("failed to setup TLS for http request", "error", err)
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
	var err error
	r.httpReq, err = r.req.HTTPRequest(r.ctx, r.manager.get)

	if err != nil {
		r.logger.Error("failed to build http request", "error", err)
		return false
	}

	return true
}

func (r *runner) doRequest() {
	var err error
	r.httpResp, err = r.httpClient.Do(r.httpReq)

	if err != nil {
		r.logger.Error("failed to execute http request", "error", err)
		return
	}

	defer r.httpResp.Body.Close()
	r.processResponse()
}

func (r *runner) processResponse() {
	switch r.httpResp.StatusCode {
	case http.StatusOK:
		// precess the returned data.
		r.decodeResponse()

	case http.StatusNotModified:
		// nothing changed, so we're all good.
		r.msg = msgOK

	default:
		r.logger.Error("unexpected status code in http response", "code", r.httpResp.StatusCode, "status", r.httpResp.Status, "headers", r.httpResp.Header)
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
}

const (
	msgStarted = "processing scheduled job"
	msgFailed  = "scheduled job processing failed"
	msgOK      = "scheduled job processed successfully"
)
