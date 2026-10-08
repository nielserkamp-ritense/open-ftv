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

// The OPA store only holds doelbinding.subsidies: any other doelbinding has no decision path, so OPA
// fails to evaluate instead of denying.
func TestAuthZEN_PDPFailureIsServerError(t *testing.T) {
	const unknown = `"subject":{"type":"doelbinding","id":"onbekend"},"action":{"name":"can_read"},"resource":{"type":"service","id":"https://inway-fsc-nlx-inway:443/brp-personen"}`

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
			name:       "evaluation",
			path:       "/authzen/v1/evaluation",
			body:       `{` + unknown + `}`,
			traceID:    "6bf92f3577b34da6a3ce929d0e0e47b1",
			wantStatus: http.StatusInternalServerError,
			want:       `{"detail":"AuthZEN evaluation failed","status":500,"title":"Internal Server Error"}`,
			wantADL:    "Error",
		},
		{
			name:       "evaluations without evaluations",
			path:       "/authzen/v1/evaluations",
			body:       `{` + unknown + `}`,
			traceID:    "6bf92f3577b34da6a3ce929d0e0e47b2",
			wantStatus: http.StatusInternalServerError,
			want:       `{"detail":"AuthZEN evaluation failed","status":500,"title":"Internal Server Error"}`,
			wantADL:    "Error",
		},
		{
			name:       "evaluations denies each item it cannot evaluate",
			path:       "/authzen/v1/evaluations",
			body:       `{"evaluations":[{` + unknown + `}]}`,
			traceID:    "6bf92f3577b34da6a3ce929d0e0e47b3",
			wantStatus: http.StatusOK,
			want:       `{"evaluations":[{"decision":false,"context":{"error":{"status":500,"message":"opa_undefined_error: /doelbinding/onbekend decision was undefined"}}}]}`,
			wantADL:    "Ok",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dsn := newADLTestDatabaseURL(t)

			cfg := &config.Config{
				PAP:            config2.PAP{Language: "rego", Store: "../../../testdata/unittest/opa"},
				PIP:            config2.PIP{Store: "../../../testdata/pip", StoreRecurse: true},
				ADLService:     "pdp-failure-test",
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
