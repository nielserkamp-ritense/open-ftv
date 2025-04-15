package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLanguage_String(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   Language
		want string
	}{
		{name: "zero", want: "<unknown>"},
		{name: "XACML", in: XACML, want: "XACML"},
		{name: "ODRL", in: ODRL, want: "ODRL"},
		{name: "OPA", in: REGO, want: "OPA/Rego"},
		{name: "Cedar", in: CEDAR, want: "Cedar"},
		{name: "Cerbos", in: CERBOS, want: "Cerbos/CEL"},
		{name: "OpenFGA", in: OPENFGA, want: "OpenFGA"},
		{name: "high", in: 255, want: "<unknown>"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.in.String()
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestLanguage_Language(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   Language
		want string
	}{
		{name: "zero", want: "<unknown>"},
		{name: "XACML", in: XACML, want: "xacml"},
		{name: "ODRL", in: ODRL, want: "odrl"},
		{name: "OPA", in: REGO, want: "rego"},
		{name: "Cedar", in: CEDAR, want: "cedar"},
		{name: "Cerbos", in: CERBOS, want: "cerbos"},
		{name: "OpenFGA", in: OPENFGA, want: "openfga"},
		{name: "high", in: 255, want: "<unknown>"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.in.Language()
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestLanguageFromString(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   string
		want Language
	}{
		{name: "empty"},
		{name: "invalid", in: "MyPL"},
		{name: "xacml", in: "xacml", want: XACML},
		{name: "opa", in: "opa", want: REGO},
		{name: "Rego", in: "opa", want: REGO},
		{name: "opa-Rego", in: "opa", want: REGO},
		{name: "OdRl", in: "OdRl", want: ODRL},
		{name: "Cedar", in: "Cedar", want: CEDAR},
		{name: "cel", in: "cel", want: CERBOS},
		{name: "CERBOS", in: "CERBOS", want: CERBOS},
		{name: "Cerbos/CEL", in: "Cerbos/CEL", want: CERBOS},
		{name: "OpenFGA", in: "OpenFGA", want: OPENFGA},
		{name: "open-fga", in: "open-fga", want: OPENFGA},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := LanguageFromString(tc.in)
			assert.Equal(t, tc.want, got)
		})
	}
}
