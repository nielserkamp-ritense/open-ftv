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

	c.ensurePEP()

	defer func() {
		if e := recover(); e != nil {
			_ = kong.Log.Err(fmt.Sprintf("OpenFTV AuthZEN: response error: %v", e))
			kong.Response.Exit(http.StatusInternalServerError, nil, nil)
		}
	}()

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

// decodeResponseBody decodes a JSON body of any shape (a service may well return an array); anything else
// is passed on as the raw string.
func decodeResponseBody(raw []byte) any {
	var body any
	if err := json.Unmarshal(raw, &body); err != nil {
		return string(raw)
	}

	return body
}
