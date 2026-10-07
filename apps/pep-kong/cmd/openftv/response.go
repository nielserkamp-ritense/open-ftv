package main

import (
	"fmt"
	"net/http"

	"github.com/Kong/go-pdk"
	"github.com/goccy/go-json"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

// Response implements the Kong-plugin Response handler. It only asks the PDP for a second decision, now
// including the upstream response, when evaluate_response is set: some rules depend on what the service
// returned rather than on what was asked.
func (c *Config) Response(kong *pdk.PDK) {
	if !c.EvaluateResponse {
		return
	}

	defer func() {
		if e := recover(); e != nil {
			_ = kong.Log.Err(fmt.Sprintf("OpenFTV AuthZEN: response error: %v", e))
			kong.Response.Exit(http.StatusInternalServerError, nil, nil)
		}
	}()

	c.ensurePEP()

	status, _ := kong.ServiceResponse.GetStatus()
	raw, _ := kong.ServiceResponse.GetRawBody()

	authResp, _, err := c.evaluate(kong, map[string]any{
		models.AttrStatus: status,
		models.AttrBody:   decodeResponseBody(raw),
	})
	if err != nil {
		c.reportError(kong, err)
		return
	}

	if !authResp.Decision {
		exitForbidden(kong, "response denied", authResp)
	}
}

// sharedRequestHeaders keeps the request headers from the access phase for the response evaluation: Kong
// has only one request, so a plugin running after this one (such as request-transformer removing
// Dpl-Core-User before it reaches the upstream service) also removes it from what the response phase reads.
const sharedRequestHeaders = "openftv.request_headers"

const maxRequestHeaders = 1000

func saveRequestHeaders(kong *pdk.PDK, headers map[string][]string) error {
	b, err := json.Marshal(headers)
	if err != nil {
		return err
	}

	return kong.Ctx.SetShared(sharedRequestHeaders, string(b))
}

func requestHeaders(kong *pdk.PDK, saved bool) (map[string][]string, error) {
	if !saved {
		headers, _ := kong.Request.GetHeaders(maxRequestHeaders)
		return headers, nil
	}

	s, err := kong.Ctx.GetSharedString(sharedRequestHeaders)
	if err != nil {
		return nil, fmt.Errorf("request headers from the access phase: %w", err)
	}

	var headers map[string][]string

	err = json.Unmarshal([]byte(s), &headers)

	return headers, err
}

// decodeResponseBody decodes a JSON body of any shape (a service may well return an array); anything else
// is passed on as the raw string.
func decodeResponseBody(raw []byte) any {
	var body any
	if err := json.Unmarshal(raw, &body); err != nil {
		return string(raw)
	}

	return body
}
