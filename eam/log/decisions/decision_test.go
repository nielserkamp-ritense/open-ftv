package decisions

import (
	"fmt"
	"testing"
	"time"

	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/authzen"
)

func TestDecision_MarshalJSON(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()

	req1 := oas.EvaluationRequest{
		Subject:  oas.Entity{Type: "gebruiker", Id: "0011"},
		Action:   oas.Action{Name: "GET"},
		Resource: oas.Entity{Type: "api", Id: "https://api.myorg.nl/v1/data"},
		Context:  map[string]any{"x": "hello world", "y": 12345, "z": true},
	}

	resp1 := oas.EvaluationResponse{
		Decision: true,
		Context: oas.ReasonObject{
			Id:          "13",
			ReasonAdmin: oas.ReasonField{"en-AU": "all good mate"},
			ReasonUser:  oas.ReasonField{"nl": "toegang verleend"},
		},
	}

	testCases := []struct {
		name    string
		d       *Decision
		want    string
		wantErr bool
	}{
		{name: "empty", d: &Decision{}, want: decisionJSON1},
		{
			name: "minimum",
			d: &Decision{
				Timestamp:   now,
				RequestType: EvaluationEndpoint,
				Request:     map[string]any{"x": "y"},
				Response:    map[string]any{"a": "b"},
				Policies:    15,
			},
			want: fmt.Sprintf(decisionJSON2, now.Format(time.RFC3339)),
		},
		{
			name: "all",
			d: &Decision{
				Timestamp:   now,
				RequestType: EvaluationEndpoint,
				Request:     map[string]any{"x": "y"},
				Response:    map[string]any{"a": "b"},
				Policies:    15,
				Information: map[string]any{"z": 123},
				Engine:      map[string]any{"c": true},
				TraceID:     "341d25f6ce326d77ff3a9004a0f45c2e",
				SpanID:      "e12fd367f790ae51",
			},
			want: fmt.Sprintf(decisionJSON3, now.Format(time.RFC3339)),
		},
		{
			name: "authzen request/response",
			d: &Decision{
				Timestamp:   now,
				RequestType: EvaluationEndpoint,
				Request:     req1,
				Response:    resp1,
				Policies:    15,
				TraceID:     "341d25f6ce326d77ff3a9004a0f45c2e",
				SpanID:      "e12fd367f790ae51",
			},
			want: fmt.Sprintf(decisionJSON4, now.Format(time.RFC3339)),
		},
		{
			name: "bad request",
			d: &Decision{
				Timestamp:   now,
				RequestType: SearchActionEndpoint,
				Request:     make(chan bool),
				Policies:    9,
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := tc.d.MarshalJSON()
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.want, string(got))
			}
		})
	}
}

func TestDecision_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC().Truncate(time.Second)

	testCases := []struct {
		name    string
		data    string
		want    *Decision
		wantErr bool
	}{
		{name: "empty", wantErr: true},
		{
			name:    "bad data",
			data:    `{"timestamp":"0001-01-01T00:00:00Z","request_type":\001\002,"request":null,"response":null,"policies":0}`,
			wantErr: true,
		},
		{
			name: "minimum",
			data: fmt.Sprintf(decisionJSON2, now.Format(time.RFC3339)),
			want: &Decision{
				Timestamp:   now,
				RequestType: EvaluationEndpoint,
				Request:     map[string]any{"x": "y"},
				Response:    map[string]any{"a": "b"},
				Policies:    15,
			},
		},
		{
			name: "authzen request/response",
			data: fmt.Sprintf(decisionJSON4, now.Format(time.RFC3339)),
			want: &Decision{
				Timestamp:   now,
				RequestType: EvaluationEndpoint,
				Request: map[string]any{
					"action":   map[string]any{"name": "GET"},
					"context":  map[string]any{"x": "hello world", "y": float64(12345), "z": true},
					"resource": map[string]any{"id": "https://api.myorg.nl/v1/data", "type": "api"},
					"subject":  map[string]any{"id": "0011", "type": "gebruiker"},
				},
				Response: map[string]any{
					"context":  map[string]any{"id": "13", "reasonAdmin": map[string]any{"en-AU": "all good mate"}, "reasonUser": map[string]any{"nl": "toegang verleend"}},
					"decision": true,
				},
				Policies: 15,
				TraceID:  "341d25f6ce326d77ff3a9004a0f45c2e",
				SpanID:   "e12fd367f790ae51",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := new(Decision)
			err := json.Unmarshal([]byte(tc.data), got)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.EqualValues(t, tc.want, got)
			}
		})
	}
}

const (
	decisionJSON1 = `{"timestamp":"0001-01-01T00:00:00Z","request_type":"???","request":null,"response":null,"policies":0}`
	decisionJSON2 = `{"timestamp":"%s","request_type":"evaluation","request":{"x":"y"},"response":{"a":"b"},"policies":15}`
	decisionJSON3 = `{"timestamp":"%s","request_type":"evaluation","request":{"x":"y"},"response":{"a":"b"},"policies":15,"information":{"z":123},` +
		`"engine":{"c":true},"trace_id":"341d25f6ce326d77ff3a9004a0f45c2e","span_id":"e12fd367f790ae51"}`
	decisionJSON4 = `{"timestamp":"%s","request_type":"evaluation","request":{"action":{"name":"GET"},"context":{"x":"hello world","y":12345,"z":true},` +
		`"resource":{"id":"https://api.myorg.nl/v1/data","type":"api"},"subject":{"id":"0011","type":"gebruiker"}},` +
		`"response":{"context":{"id":"13","reasonAdmin":{"en-AU":"all good mate"},"reasonUser":{"nl":"toegang verleend"}},"decision":true},` +
		`"policies":15,"trace_id":"341d25f6ce326d77ff3a9004a0f45c2e","span_id":"e12fd367f790ae51"}`
)
