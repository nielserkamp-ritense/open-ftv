package decisions

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBodyFromDecision_storesJSONObjects(t *testing.T) {
	t.Parallel()

	const span = `{"adl.core.request":{"subject":{"id":"alice"}},"adl.core.response":{"decision":true}}`

	d := new(Decision)
	applyBodyAttribute(span, d)

	// pgx marshals a map bound to a JSONB column with encoding/json, which
	// base64-encodes []byte but passes json.RawMessage through untouched.
	stored, err := json.Marshal(BodyFromDecision(d))
	require.NoError(t, err)
	assert.JSONEq(t, span, string(stored))

	var decoded map[string]any
	require.NoError(t, json.Unmarshal(stored, &decoded))

	req, resp := RequestResponseFromBody(decoded)
	require.NotNil(t, req)
	require.NotNil(t, resp)
	assert.Equal(t, "alice", req["subject"].(map[string]any)["id"])
	assert.Equal(t, true, resp["decision"])
}

func TestAttributesFromDecision(t *testing.T) {
	t.Parallel()

	assert.Equal(t, map[string]any{}, AttributesFromDecision(nil))
	assert.Equal(t, map[string]any{}, AttributesFromDecision(&Decision{}))
	assert.Equal(t,
		map[string]any{"adl.fsc.transaction_id": "abc-123"},
		AttributesFromDecision(&Decision{FSCTransactionID: "abc-123"}),
	)
}

func TestApplyAttributesAttribute(t *testing.T) {
	t.Parallel()

	d := new(Decision)
	applyAttributesAttribute(`{"adl.fsc.transaction_id":"abc-123"}`, d)
	assert.Equal(t, "abc-123", d.FSCTransactionID)

	d2 := new(Decision)
	applyAttributesAttribute("{}", d2)
	assert.Equal(t, "", d2.FSCTransactionID)

	d3 := new(Decision)
	applyAttributesAttribute("not json", d3)
	assert.Equal(t, "", d3.FSCTransactionID)
}

func TestRequestResponseFromBody_decodedMaps(t *testing.T) {
	t.Parallel()

	body := map[string]any{
		"adl.core.request": map[string]any{
			"subject": map[string]any{"id": "alice", "type": "user"},
		},
		"adl.core.response": map[string]any{"decision": true},
	}

	req, resp := RequestResponseFromBody(body)
	require.NotNil(t, req)
	require.NotNil(t, resp)
	assert.Equal(t, "alice", req["subject"].(map[string]any)["id"])
	assert.Equal(t, true, resp["decision"])
}
