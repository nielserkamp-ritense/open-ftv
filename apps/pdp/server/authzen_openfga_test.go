//go:build integration

package server

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/pdp/config"
	config2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/config"
)

// The OpenFGA store only defines service#can_update for doelbinding, with tuples for subsidies and
// burgerzaken: any other relation is not in the model, so OpenFGA fails to check it.
func TestAuthZEN_OpenFGA(t *testing.T) {
	const (
		resource = `"resource":{"type":"service","id":"https://inway-fsc-nlx-inway:443/brp-personen"}`
		allowed  = `"subject":{"type":"doelbinding","id":"subsidies"},"action":{"name":"can_update"},` + resource
		denied   = `"subject":{"type":"doelbinding","id":"onbekend"},"action":{"name":"can_update"},` + resource
		failed   = `"subject":{"type":"doelbinding","id":"subsidies"},"action":{"name":"can_fly"},` + resource
		noStore  = `"subject":{"type":"persoon","id":"subsidies"},"action":{"name":"can_update"},` + resource
	)

	testCases := []struct {
		name       string
		path       string
		body       string
		traceID    string
		wantStatus int
		want       string
		wantADL    string
	}{
		{
			name:       "allowed",
			path:       "/authzen/v1/evaluation",
			body:       `{` + allowed + `}`,
			traceID:    "7bf92f3577b34da6a3ce929d0e0e47c1",
			wantStatus: http.StatusOK,
			want:       `{"decision":true}`,
			wantADL:    "Ok",
		},
		{
			name:       "policy denial",
			path:       "/authzen/v1/evaluation",
			body:       `{` + denied + `}`,
			traceID:    "7bf92f3577b34da6a3ce929d0e0e47c2",
			wantStatus: http.StatusOK,
			want:       `{"decision":false}`,
			wantADL:    "Ok",
		},
		{
			name:       "engine failure",
			path:       "/authzen/v1/evaluation",
			body:       `{` + failed + `}`,
			traceID:    "7bf92f3577b34da6a3ce929d0e0e47c3",
			wantStatus: http.StatusInternalServerError,
			want:       `{"detail":"AuthZEN evaluation failed","status":500,"title":"Internal Server Error"}`,
			wantADL:    "Error",
		},
		{
			name:       "unknown store",
			path:       "/authzen/v1/evaluation",
			body:       `{` + noStore + `}`,
			traceID:    "7bf92f3577b34da6a3ce929d0e0e47c4",
			wantStatus: http.StatusInternalServerError,
			want:       `{"detail":"AuthZEN evaluation failed","status":500,"title":"Internal Server Error"}`,
			wantADL:    "Error",
		},
		{
			name:       "engine failure in batch",
			path:       "/authzen/v1/evaluations",
			body:       `{"evaluations":[{` + allowed + `},{` + failed + `}]}`,
			traceID:    "7bf92f3577b34da6a3ce929d0e0e47c5",
			wantStatus: http.StatusOK,
			want:       `{"evaluations":[{"decision":true},{"decision":false,"context":{"error":{"status":500,"message":"invalid relation: relation 'service#can_fly' not found"}}}]}`,
			wantADL:    "Ok",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dsn := newADLTestDatabaseURL(t)

			cfg := &config.Config{
				PAP:            config2.PAP{Language: "openfga", Store: "../../../testdata/unittest/openfga"},
				PIP:            config2.PIP{Store: "../../../testdata/pip", StoreRecurse: true},
				ADLService:     "pdp-openfga-test",
				ADLDatabaseURL: dsn,
			}

			app := newADLTestApp(t, cfg)

			req := httptest.NewRequest(http.MethodPost, tc.path, bytes.NewReader([]byte(tc.body)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("traceparent", "00-"+tc.traceID+"-00f067aa0ba902b7-01")

			resp, err := app.Test(req, -1)
			require.NoError(t, err)
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			assert.Equal(t, tc.wantStatus, resp.StatusCode)
			assert.JSONEq(t, tc.want, string(body))

			row := queryDecisionByTraceID(t, dsn, tc.traceID)
			assert.Equal(t, tc.wantADL, row.Status)
		})
	}
}
