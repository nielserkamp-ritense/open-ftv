package main

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	auth "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	types "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/goccy/go-json"
	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/codes"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pep"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/authzen"
)

func newAuthZEN(c *cfg, logger *slog.Logger) *authServer {
	return &authServer{
		logger:     logger,
		pdp:        c.PDP,
		timeout:    c.Timeout,
		httpClient: &http.Client{Timeout: c.Timeout},
		pep:        pep.New(context.Background(), logger),
	}
}

// Check implements an Envoy authorization check.
func (server *authServer) Check(ctx context.Context, request *auth.CheckRequest) (*auth.CheckResponse, error) {
	httpReq := request.Attributes.Request.Http

	headers := make(map[string][]string, len(httpReq.Headers))
	for k, v := range httpReq.Headers {
		headers[k] = []string{v}
	}

	// Resolved once so every outgoing hop for this request (the PDP call, and the forwarded call to the
	// upstream service) shares the same trace_id — resolving independently per hop would otherwise mint a
	// separate, unrelated trace_id on each hop whenever the incoming request carried none.
	incomingTraceParent := models.FirstHeader(headers, models.HeaderTraceParent)
	baseTraceParent := models.ResolveTraceParent(incomingTraceParent)
	traceState := models.ResolveTraceState(incomingTraceParent, models.FirstHeader(headers, models.HeaderTraceState))
	fscTransactionID := models.FirstHeader(headers, models.HeaderFSCTransactionID)

	upstream := upstreamTrace{
		traceParent:      models.ResolveOutgoingTraceParent(baseTraceParent),
		traceState:       traceState,
		fscTransactionID: fscTransactionID,
	}

	if httpReq.Method == "OPTIONS" {
		return allowed(upstream), nil
	}

	if err := server.authorizeRequest(ctx, headers, baseTraceParent, traceState, fscTransactionID, request); err != nil {
		server.logger.Error("authorization failed", "request", request, "error", err)
		return denied(http.StatusUnauthorized, "unauthorized"), nil
	}

	return allowed(upstream), nil
}

// upstreamTrace is the trace context applied, via OkHttpResponse.Headers, to the request Envoy forwards to
// the protected service.
type upstreamTrace struct {
	traceParent      string
	traceState       string
	fscTransactionID string
}

func (server *authServer) authorizeRequest(ctx context.Context, headers map[string][]string, baseTraceParent, traceState, fscTransactionID string, request *auth.CheckRequest) error {
	httpReq := request.Attributes.Request.Http

	uri := &url.URL{
		Scheme:   httpReq.Scheme,
		Host:     httpReq.Host,
		Path:     httpReq.Path,
		RawQuery: httpReq.Query,
	}

	now := time.Now().UTC()
	req := models.Request{
		RequestTime: &now,
		Method:      httpReq.Method,
		URL:         uri,
		Headers:     headers,
		Body:        httpReq.RawBody,
	}

	traceParent := models.ResolveOutgoingTraceParent(baseTraceParent)
	models.SetHeader(req.Headers, models.HeaderTraceParent, traceParent)

	parc := server.pep.PARCFromRequest(&req, func(uid string) (*models.Entity, uint64, error) { return nil, 0, nil })

	authBody := authzen.EvaluationRequest{
		Subject: authzen.Entity{
			Type:       parc.Principal.Type(),
			Id:         parc.Principal.ID(),
			Properties: models.MapFromAttributes(parc.Principal.Attributes()),
		},
		Action: authzen.Action{
			Name:       parc.Action.ID(),
			Properties: models.MapFromAttributes(parc.Action.Attributes()),
		},
		Resource: authzen.Entity{
			Type:       parc.Resource.Type(),
			Id:         parc.Resource.ID(),
			Properties: models.MapFromAttributes(parc.Resource.Attributes()),
		},
		Context: models.MapFromAttributes(parc.Context),
	}

	b, err := json.Marshal(authBody)
	if err != nil {
		return fmt.Errorf("error encoding request body: %w", err)
	}

	ctx2, cancel := context.WithTimeout(ctx, server.timeout)
	defer cancel()

	authReq, err2 := http.NewRequestWithContext(ctx2, http.MethodPost, server.pdp, bytes.NewBuffer(b))
	if err2 != nil {
		return fmt.Errorf("error creating authorization request: %w", err2)
	}

	authReq.Header.Set(models.HeaderTraceParent, traceParent)

	if traceState != "" {
		authReq.Header.Set(models.HeaderTraceState, traceState)
	}

	if fscTransactionID != "" {
		authReq.Header.Set(models.HeaderFSCTransactionID, fscTransactionID)
	}

	resp, err3 := http.DefaultClient.Do(authReq)
	if err3 != nil {
		return fmt.Errorf("error executing authorization request: %w", err3)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("authorization request failed with status %s", resp.Status)
	}

	authResp := authzen.EvaluationResponse{}
	if err = json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return fmt.Errorf("error decoding authorization response: %w", err)
	}

	if !authResp.Decision {
		return fmt.Errorf("authorization failed: access denied: %v", authResp)
	}
	return nil
}

func denied(code int32, body string) *auth.CheckResponse {
	return &auth.CheckResponse{
		Status: &status.Status{Code: code},
		HttpResponse: &auth.CheckResponse_DeniedResponse{
			DeniedResponse: &auth.DeniedHttpResponse{
				Status: &types.HttpStatus{
					Code: types.StatusCode(code),
				},
				Body: body,
			},
		},
	}
}

func allowed(tc upstreamTrace) *auth.CheckResponse {
	headers := []*corev3.HeaderValueOption{headerOption(models.HeaderTraceParent, tc.traceParent)}

	var headersToRemove []string

	// If the request had no valid traceparent, we can't tell which trace the tracestate belongs
	// to, so we don't forward it (§3.2.2/W3C §4.2-§4.3). But Envoy forwards headers by default: if we
	// just leave tracestate out of our response, Envoy sends the client's original value anyway.
	// So we have to explicitly tell Envoy to remove it, via HeadersToRemove.
	if tc.traceState != "" {
		headers = append(headers, headerOption(models.HeaderTraceState, tc.traceState))
	} else {
		headersToRemove = append(headersToRemove, models.HeaderTraceState)
	}

	if tc.fscTransactionID != "" {
		headers = append(headers, headerOption(models.HeaderFSCTransactionID, tc.fscTransactionID))
	}

	return &auth.CheckResponse{
		Status: &status.Status{Code: int32(codes.OK)},
		HttpResponse: &auth.CheckResponse_OkResponse{
			OkResponse: &auth.OkHttpResponse{Headers: headers, HeadersToRemove: headersToRemove},
		},
	}
}

// headerOption overwrites rather than appends: the default APPEND_IF_EXISTS_OR_ADD would otherwise leave the
// client's original header alongside ours instead of replacing it.
func headerOption(key, value string) *corev3.HeaderValueOption {
	return &corev3.HeaderValueOption{
		Header:       &corev3.HeaderValue{Key: key, Value: value},
		AppendAction: corev3.HeaderValueOption_OVERWRITE_IF_EXISTS_OR_ADD,
	}
}

type authServer struct {
	logger     *slog.Logger
	pdp        string
	timeout    time.Duration
	httpClient *http.Client
	pep        *pep.PEP
}
