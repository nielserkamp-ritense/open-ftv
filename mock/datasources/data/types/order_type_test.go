package types

import (
	"testing"

	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrderTypeFromString(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		want OrderType
	}{
		{name: "x", want: OrderDescending},
		{name: "bad name", want: OrderDescending},
		{name: "desc", want: OrderDescending},
		{name: "asc", want: OrderAscending},
		{name: "Ascending", want: OrderAscending},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := OrderTypeFromString(tc.name)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestOrderType_String(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		t    OrderType
		want string
	}{
		{name: "unknown", t: 99, want: "Descending"},
		{name: "asc", t: OrderAscending, want: "Ascending"},
		{name: "desc", t: OrderDescending, want: "Descending"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.t.String()
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestOrderType_JSON(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		t    OrderType
	}{
		{name: "asc", t: OrderAscending},
		{name: "desc", t: OrderDescending},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := json.Marshal(tc.t)
			require.NoError(t, err)
			require.NotNil(t, got)

			var t2 OrderType
			err = json.Unmarshal(got, &t2)
			require.NoError(t, err)
			assert.Equal(t, tc.t, t2)
		})
	}
}

func TestOrderType_YAML(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		t    OrderType
	}{
		{name: "asc", t: OrderAscending},
		{name: "desc", t: OrderDescending},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := yaml.Marshal(tc.t)
			require.NoError(t, err)
			require.NotNil(t, got)

			var t2 OrderType
			err = yaml.Unmarshal(got, &t2)
			require.NoError(t, err)
			assert.Equal(t, tc.t, t2)
		})
	}
}
