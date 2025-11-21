package main

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

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
	if request.Attributes.Request.Http.Method == "OPTIONS" {
		return allowed(), nil
	}

	if err := server.authorizeRequest(ctx, request); err != nil {
		server.logger.Error("authorization failed", "request", request, "error", err)
		return denied(http.StatusUnauthorized, "unauthorized"), nil
	}

	return allowed(), nil
}

func (server *authServer) authorizeRequest(ctx context.Context, request *auth.CheckRequest) error {
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
		Headers:     make(map[string][]string, len(httpReq.Headers)),
		Body:        httpReq.RawBody,
	}

	for k, v := range httpReq.Headers {
		req.Headers[k] = []string{v}
	}

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

func allowed() *auth.CheckResponse {
	return &auth.CheckResponse{
		Status: &status.Status{Code: int32(codes.OK)},
		HttpResponse: &auth.CheckResponse_OkResponse{
			OkResponse: &auth.OkHttpResponse{},
		},
	}
}

type authServer struct {
	logger     *slog.Logger
	pdp        string
	timeout    time.Duration
	httpClient *http.Client
	pep        *pep.PEP
}
