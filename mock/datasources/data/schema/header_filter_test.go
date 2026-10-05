package schema

import (
	"testing"

	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHeaderFilter(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		pattern string
		wantErr bool
	}{
		{name: "valid", pattern: `^([0-9]+)\.[0-9]+$`},
		{name: "does not compile", pattern: `^([0-9]+`, wantErr: true},
		{name: "no capture group", pattern: `^[0-9]+$`, wantErr: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h, err := NewHeaderFilter("Afnemersindicatie", "inschrijving.gemeente", tc.pattern)
			if tc.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.pattern, h.Pattern)
		})
	}
}

func TestHeaderFilter_Value(t *testing.T) {
	t.Parallel()

	h, err := NewHeaderFilter("Afnemersindicatie", "inschrijving.gemeente", `^([0-9]+)\.[0-9]+$`)
	require.NoError(t, err)

	testCases := []struct {
		name    string
		header  string
		want    string
		wantErr bool
	}{
		{name: "three digits", header: "001.01", want: "001"},
		{name: "four digits", header: "0363.01", want: "0363"},
		{name: "missing", header: "", wantErr: true},
		{name: "no dot", header: "001", wantErr: true},
		{name: "not a number", header: "abc.01", wantErr: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := h.Value(tc.header)
			if tc.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestHeaderFilter_MarshalJSON(t *testing.T) {
	t.Parallel()

	h, err := NewHeaderFilter("Afnemersindicatie", "inschrijving.gemeente", `^([0-9]+)\.[0-9]+$`)
	require.NoError(t, err)

	got, err := json.Marshal(h)
	require.NoError(t, err)
	assert.Equal(t, `{"header":"Afnemersindicatie","field":"inschrijving.gemeente","pattern":"^([0-9]+)\\.[0-9]+$"}`, string(got))

	var h2 HeaderFilter
	require.NoError(t, json.Unmarshal(got, &h2))
	v, err := h2.Value("001.01")
	require.NoError(t, err)
	assert.Equal(t, "001", v)
}

func TestHeaderFilter_UnmarshalJSON_Error(t *testing.T) {
	t.Parallel()

	var h HeaderFilter
	require.Error(t, h.UnmarshalJSON([]byte("\000\001")))
	require.Error(t, h.UnmarshalJSON([]byte(`{"header":"X","field":"f","pattern":"^[0-9]+$"}`)))
}

func TestHeaderFilter_MarshalYAML(t *testing.T) {
	t.Parallel()

	h, err := NewHeaderFilter("Afnemersindicatie", "inschrijving.gemeente", `^([0-9]+)\.[0-9]+$`)
	require.NoError(t, err)

	got, err := yaml.Marshal(h)
	require.NoError(t, err)

	var h2 HeaderFilter
	require.NoError(t, yaml.Unmarshal(got, &h2))
	assert.Equal(t, h.Header, h2.Header)
	assert.Equal(t, h.Field, h2.Field)
	assert.Equal(t, h.Pattern, h2.Pattern)
}

func TestHeaderFilter_UnmarshalYAML_Error(t *testing.T) {
	t.Parallel()

	var h HeaderFilter
	require.Error(t, h.UnmarshalYAML([]byte("\000\001")))
	require.Error(t, h.UnmarshalYAML([]byte("header: X\nfield: f\npattern: '^([0-9]+'\n")))
}

func TestEndpoint_HeaderFilters_YAML(t *testing.T) {
	t.Parallel()

	in := `
version: 1
path: /personen
headerFilters:
  - header: Afnemersindicatie
    field: inschrijving.gemeente
    pattern: '^([0-9]+)\.[0-9]+$'
`

	var e Endpoint
	require.NoError(t, yaml.Unmarshal([]byte(in), &e))
	require.Len(t, e.HeaderFilters, 1)

	v, err := e.HeaderFilters[0].Value("001.01")
	require.NoError(t, err)
	assert.Equal(t, "001", v)
}
