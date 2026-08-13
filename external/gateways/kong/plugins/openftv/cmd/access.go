package main

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/Kong/go-pdk"
	"github.com/goccy/go-json"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/authzen"
)

// Access implements the Kong-plugin Access handler.
func (c *Config) Access(kong *pdk.PDK) {
	c.ensurePEP()

	var method string
	var uri *url.URL

	defer func() {
		if e := recover(); e != nil {
			_ = kong.Log.Err(fmt.Sprintf("OpenFTV AuthZEN: access error: %v", e))
			kong.Response.Exit(http.StatusInternalServerError, nil, nil)
		}
	}()

	method, _ = kong.Request.GetMethod()
	scheme, _ := kong.Request.GetScheme()
	host, _ := kong.Request.GetHost()
	port, _ := kong.Request.GetPort()
	path, _ := kong.Request.GetPath()
	query, _ := kong.Request.GetRawQuery()
	headers, _ := kong.Request.GetHeaders(1000)
	body, _ := kong.Request.GetRawBody()

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

	traceParent := models.ResolveOutgoingTraceParent(baseTraceParent)
	models.SetHeader(headers, models.HeaderTraceParent, traceParent)

	uri = &url.URL{
		Scheme:   scheme,
		Host:     net.JoinHostPort(host, strconv.Itoa(port)),
		Path:     path,
		RawQuery: query,
	}

	now := time.Now().UTC()
	req := models.Request{
		RequestTime: &now,
		Method:      method,
		URL:         uri,
		Headers:     headers,
		Body:        body,
	}

	parc := c.pep.PARCFromRequest(&req, func(uid string) (*models.Entity, uint64, error) { return nil, 0, nil })

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
		c.reportError(kong, err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.Timeout)*time.Millisecond)
	defer cancel()

	authReq, err2 := http.NewRequestWithContext(ctx, http.MethodPost, c.PDP, bytes.NewBuffer(b))
	if err2 != nil {
		c.reportError(kong, err2)
		return
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
		c.reportError(kong, err3)
		return
	}

	defer resp.Body.Close()

	authResp := authzen.EvaluationResponse{}
	if err = json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		c.reportError(kong, err)
		return
	}

	if !authResp.Decision {
		_ = kong.Log.Err(fmt.Sprintf("OpenFTV AuthZEN: access denied: %v", authResp))
		kong.Response.Exit(http.StatusForbidden, nil, nil)

		return
	}

	if err = setUpstreamTrace(kong, upstream); err != nil {
		c.reportError(kong, err)
	}
}

// upstreamTrace is the trace context applied, via kong.ServiceRequest, to the request Kong forwards to the
// protected service.
type upstreamTrace struct {
	traceParent      string
	traceState       string
	fscTransactionID string
}

func setUpstreamTrace(kong *pdk.PDK, tc upstreamTrace) error {
	if err := kong.ServiceRequest.SetHeader(models.HeaderTraceParent, tc.traceParent); err != nil {
		return err
	}

	// Must actively clear, not skip: Kong forwards a client header unchanged unless told otherwise.
	if tc.traceState != "" {
		if err := kong.ServiceRequest.SetHeader(models.HeaderTraceState, tc.traceState); err != nil {
			return err
		}
	} else if err := kong.ServiceRequest.ClearHeader(models.HeaderTraceState); err != nil {
		return err
	}

	if tc.fscTransactionID != "" {
		return kong.ServiceRequest.SetHeader(models.HeaderFSCTransactionID, tc.fscTransactionID)
	}

	return nil
}

func (c *Config) reportError(kong *pdk.PDK, err error) {
	_ = kong.Log.Err(fmt.Sprintf("OpenFTV AuthZEN: access error: %v", err))
	kong.Response.Exit(http.StatusInternalServerError, nil, nil)
}
