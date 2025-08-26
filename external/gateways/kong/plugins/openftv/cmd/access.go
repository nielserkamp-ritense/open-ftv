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
	var method string
	var uri *url.URL

	defer func() {
		if e := recover(); e != nil {
			_ = kong.Log.Err(fmt.Sprintf("OpenFTV AuthZEN: access error: %v", e))
			kong.Response.ExitStatus(http.StatusInternalServerError)
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
		kong.Response.ExitStatus(http.StatusForbidden)
	}
}

func (c *Config) reportError(kong *pdk.PDK, err error) {
	_ = kong.Log.Err(fmt.Sprintf("OpenFTV AuthZEN: access error: %v", err))
	kong.Response.ExitStatus(http.StatusInternalServerError)
}
