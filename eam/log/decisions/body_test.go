package decisions

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestResponseFromBody_decodedMaps(t *testing.T) {
	t.Parallel()

	body := map[string]any{
		"request": map[string]any{
			"subject": map[string]any{"id": "alice", "type": "user"},
		},
		"response": map[string]any{"decision": true},
	}

	req, resp := RequestResponseFromBody(body)
	require.NotNil(t, req)
	require.NotNil(t, resp)
	assert.Equal(t, "alice", req["subject"].(map[string]any)["id"])
	assert.Equal(t, true, resp["decision"])
}

func TestRequestResponseFromBody_base64Encoded(t *testing.T) {
	t.Parallel()

	reqJSON := `{"subject":{"id":"trace-known-1","type":"medewerker"},"action":{"name":"GET"}}`
	respJSON := `{"decision":true,"context":{"id":"0"}}`
	body := map[string]any{
		"request":  base64.StdEncoding.EncodeToString([]byte(reqJSON)),
		"response": base64.StdEncoding.EncodeToString([]byte(respJSON)),
	}

	req, resp := RequestResponseFromBody(body)
	require.NotNil(t, req)
	require.NotNil(t, resp)
	assert.Equal(t, "trace-known-1", req["subject"].(map[string]any)["id"])
	assert.Equal(t, true, resp["decision"])
}
