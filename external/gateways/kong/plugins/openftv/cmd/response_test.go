package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Kong/go-pdk/test"
	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const vehicles = `[{"kenteken":"56-56-56","diplomatiek":true}]`

func vehiclesEnv(t *testing.T) *test.TestEnv {
	t.Helper()

	env, err := test.New(t, test.Request{
		Method:  "POST",
		Url:     "https://gw.test/v1/voertuigen",
		Headers: http.Header{"Dpl-Core-User": {"Diane"}},
		Body:    []byte(`{"@filter":"kenteken.bsn=999990299"}`),
	})
	require.NoError(t, err)

	env.ServiceRes = test.Response{Status: http.StatusOK, Body: []byte(vehicles)}

	return env
}

func TestResponse_Disabled_SkipsPDP(t *testing.T) {
	called := false

	pdp := decisionPDP(t, false, func(*http.Request) { called = true })
	defer pdp.Close()

	env := vehiclesEnv(t)
	env.DoResponse(&Config{PDP: pdp.URL, Timeout: 5000})

	assert.False(t, called)
	assert.Equal(t, http.StatusOK, env.ClientRes.Status)
	assert.JSONEq(t, vehicles, string(env.ClientRes.Body))
}

func TestResponse_Allowed_SendsUpstreamResponseToPDP(t *testing.T) {
	var evaluation struct {
		Context struct {
			Response struct {
				Status int `json:"status"`
				Body   any `json:"body"`
			} `json:"response"`
		} `json:"context"`
	}

	pdp := decisionPDP(t, true, func(r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		assert.NoError(t, json.Unmarshal(b, &evaluation))
	})
	defer pdp.Close()

	cfg := &Config{PDP: pdp.URL, Timeout: 5000, EvaluateResponse: true}
	env := vehiclesEnv(t)
	env.DoAccess(cfg)
	env.DoResponse(cfg)

	assert.Equal(t, http.StatusOK, evaluation.Context.Response.Status)
	assert.Equal(t, []any{map[string]any{"kenteken": "56-56-56", "diplomatiek": true}}, evaluation.Context.Response.Body)
	assert.Equal(t, http.StatusOK, env.ClientRes.Status)
	assert.JSONEq(t, vehicles, string(env.ClientRes.Body))
}

func TestResponse_Denied_ExitsForbiddenWithReason(t *testing.T) {
	pdp := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if b, _ := io.ReadAll(r.Body); !bytes.Contains(b, []byte(`"response"`)) {
			_, _ = w.Write([]byte(`{"decision":true}`))
			return
		}

		_, _ = w.Write([]byte(`{"decision":false,"context":{"reason_user":{"en":"diplomatiek voertuig"}}}`))
	}))
	defer pdp.Close()

	cfg := &Config{PDP: pdp.URL, Timeout: 5000, EvaluateResponse: true}
	env := vehiclesEnv(t)
	env.DoAccess(cfg)
	env.DoResponse(cfg)

	assert.Equal(t, http.StatusForbidden, env.ClientRes.Status)
	assert.JSONEq(t, `{"title":"Forbidden","status":403,"detail":"diplomatiek voertuig"}`, string(env.ClientRes.Body))
}

// Kong has one request for the client and the upstream service: a plugin that strips a header before
// forwarding (request-transformer on Dpl-Core-User) also strips it from what the response phase reads.
func TestResponse_HeaderRemovedAfterAccess_StillSentToPDP(t *testing.T) {
	var evaluation struct {
		Context struct {
			CoreUser string `json:"core_user"`
		} `json:"context"`
	}

	pdp := decisionPDP(t, true, func(r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		assert.NoError(t, json.Unmarshal(b, &evaluation))
	})
	defer pdp.Close()

	cfg := &Config{PDP: pdp.URL, Timeout: 5000, EvaluateResponse: true}
	env := vehiclesEnv(t)
	env.DoAccess(cfg)
	env.ClientReq.Headers.Del("Dpl-Core-User")
	env.DoResponse(cfg)

	assert.Equal(t, "Diane", evaluation.Context.CoreUser)
	assert.Equal(t, http.StatusOK, env.ClientRes.Status)
}
