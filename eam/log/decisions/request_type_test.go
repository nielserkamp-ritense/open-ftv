package decisions

import (
	"testing"

	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthRequestType_String(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		t    AuthRequestType
		want string
	}{
		{name: "zero", t: 0, want: "???"},
		{name: "evaluation", t: EvaluationEndpoint, want: "evaluation"},
		{name: "evaluations", t: EvaluationsEndpoint, want: "evaluations"},
		{name: "search subject", t: SearchSubjectEndpoint, want: "search_subject"},
		{name: "search action", t: SearchActionEndpoint, want: "search_action"},
		{name: "search resource", t: SearchResourceEndpoint, want: "search_resource"},
		{name: "invalid", t: 99, want: "???"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.t.String()
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestAuthRequestType_MarshalJSON(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		t    AuthRequestType
		want string
	}{
		{name: "zero", t: 0, want: `{"request_type":"???"}`},
		{name: "evaluation", t: EvaluationEndpoint, want: `{"request_type":"evaluation"}`},
		{name: "evaluations", t: EvaluationsEndpoint, want: `{"request_type":"evaluations"}`},
		{name: "search subject", t: SearchSubjectEndpoint, want: `{"request_type":"search_subject"}`},
		{name: "search action", t: SearchActionEndpoint, want: `{"request_type":"search_action"}`},
		{name: "search resource", t: SearchResourceEndpoint, want: `{"request_type":"search_resource"}`},
		{name: "invalid", t: 99, want: `{"request_type":"???"}`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			q := struct {
				RequestType AuthRequestType `json:"request_type"`
			}{tc.t}

			b, err := json.Marshal(q)
			require.NoError(t, err)
			require.Equal(t, tc.want, string(b))
		})
	}
}

func TestAuthRequestType_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		b       string
		wantErr bool
		want    AuthRequestType
	}{
		{name: "integer", b: `{"request_type":123"}`, wantErr: true, want: 0},
		{name: "invalid", b: `{"request_type":\1\2\3"}`, wantErr: true, want: 0},
		{name: "evaluation", b: `{"request_type":"evaluation"}`, want: EvaluationEndpoint},
		{name: "evaluations", b: `{"request_type":"evaluations"}`, want: EvaluationsEndpoint},
		{name: "search subject", b: `{"request_type":"search_subject"}`, want: SearchSubjectEndpoint},
		{name: "search action", b: `{"request_type":"search_action"}`, want: SearchActionEndpoint},
		{name: "search resource", b: `{"request_type":"search_resource"}`, want: SearchResourceEndpoint},
		{name: "unknown", b: `{"request_type":"unknown"}`, want: 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			q := struct {
				RequestType AuthRequestType `json:"request_type"`
			}{0}

			err := json.Unmarshal([]byte(tc.b), &q)
			require.Equal(t, tc.wantErr, err != nil)
			require.Equal(t, tc.want, q.RequestType)
		})
	}
}
